package docker

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/docker/docker/api/types/build"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"

	"github.com/amyy54/garden/internal/connhelper"
	"github.com/amyy54/garden/internal/types"
)

// Creates the docker API client. Handles SSH with connhelper
func createClient(options ClientOptions) (*client.Client, error) {
	var cli *client.Client
	var host string

	// If runner is blank, presume that the user wants us to read docker context
	// If the user specifically supplies a context as well, run this too
	if len(options.Runner) == 0 || options.IsContext {
		context, err := getContext(options.Runner)
		if err != nil {
			return cli, err
		} else {
			host = context
		}
	} else {
		host = options.Runner
	}

	// Docker API does not have support for SSH
	// Needs to be handled with connhelper
	if strings.HasPrefix(host, "ssh") {
		helper, err := connhelper.GetConnectionHelper(host)
		if err != nil {
			return cli, err
		}
		httpClient := &http.Client{
			Transport: &http.Transport{
				DialContext: helper.Dialer,
			},
		}

		cli, err = client.NewClientWithOpts(
			client.WithHost(helper.Host),
			client.WithHTTPClient(httpClient),
			client.WithAPIVersionNegotiation(),
		)
		if err != nil {
			return cli, err
		}
	} else {
		// If not using SSH, everything else will work
		// In theory
		var err error
		cli, err = client.NewClientWithOpts(
			client.WithHost(host),
			client.WithAPIVersionNegotiation(),
		)
		if err != nil {
			return cli, err
		}
	}
	return cli, nil
}

// Gets the context by parsing "docker context"
func getContext(context string) (string, error) {
	_, err := exec.LookPath("docker")
	if err != nil {
		return "", fmt.Errorf("Docker does not exist in path")
	}
	var exec_run *exec.Cmd
	if len(context) == 0 {
		exec_run = exec.Command("docker", "context", "inspect", "-f=json")
	} else {
		exec_run = exec.Command("docker", "context", "inspect", context, "-f=json")
	}

	output, err := exec_run.Output()
	if err != nil {
		return "", fmt.Errorf("Failed to get output from docker, %v", err)
	}

	var res []Inspect

	err = json.Unmarshal(output, &res)
	if err != nil {
		return "", fmt.Errorf("Failed to unmarshal docker result, %v", err)
	}

	if len(res) > 0 {
		context := res[0]
		host, ok := context.Endpoints["docker"]
		if !ok {
			return "", fmt.Errorf("\"docker\" endpoint does not exist")
		} else {
			return host.Host, nil
		}
	} else {
		return "", fmt.Errorf("The supplied context doesn't exist, %v", res)
	}
}

// Runs the specified module with docker in a goroutine
// Requires running in a channel that is passed
func runModule(ch chan types.ContainerResult, cli *client.Client, options DockerOptions) {
	// Create a lot of the initial variables for use later
	ctx := context.Background()
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)

	identifier := fmt.Sprintf("garden/%s/%s", options.Version, options.Identifier.ToString())
	slog.Debug("Generated docker identifier", "identifier", identifier)

	// Read the Dockerfile path from passed options
	// The "modules" package already checked for a valid hash
	dockerfileContents, err := os.ReadFile(options.DockerfilePath)
	if err != nil {
		ch <- types.ContainerResult{Identifier: options.Identifier, Error: err}
		return
	}

	// Docker's rest API expects everything to be contained in a tar file
	// Here the dockerfile is renamed to simply "Dockerfile" for ease of use
	err = tw.WriteHeader(&tar.Header{
		Name: "Dockerfile",
		Size: int64(len(dockerfileContents)),
		Mode: 0600,
	})
	if err != nil {
		ch <- types.ContainerResult{Identifier: options.Identifier, Error: err}
		return
	}

	_, err = tw.Write(dockerfileContents)
	if err != nil {
		ch <- types.ContainerResult{Identifier: options.Identifier, Error: err}
		return
	}
	tw.Close()

	// Builds the actual container. Gives the tar buffer
	buildResponse, err := cli.ImageBuild(ctx, buf, build.ImageBuildOptions{
		Tags:       []string{identifier},
		Remove:     true,
		Dockerfile: "Dockerfile",
	})
	if err != nil {
		ch <- types.ContainerResult{Identifier: options.Identifier, Error: err}
		return
	}
	defer buildResponse.Body.Close()

	var buildbuf strings.Builder
	_, err = io.Copy(&buildbuf, buildResponse.Body)
	if err != nil {
		ch <- types.ContainerResult{Identifier: options.Identifier, Error: err}
		return
	}

	slog.Info("Successfully built container", "identifier", options.Identifier.ToString())

	// TTY is set to true here as it makes it easier to collect all container logs
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: identifier,
		Cmd:   options.Command,
		Tty:   true,
	}, nil, nil, nil, "")
	if err != nil {
		ch <- types.ContainerResult{Identifier: options.Identifier, BuildOutput: buildbuf.String(), Error: err}
		return
	}

	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		ch <- types.ContainerResult{Identifier: options.Identifier, BuildOutput: buildbuf.String(), Error: err}
		return
	}

	slog.Info("Successfully started container", "identifier", options.Identifier.ToString())

	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			ch <- types.ContainerResult{Identifier: options.Identifier, BuildOutput: buildbuf.String(), Error: err}
			return
		}
	case <-statusCh:
	}

	out, err := cli.ContainerLogs(ctx, resp.ID, container.LogsOptions{ShowStdout: true})
	if err != nil {
		ch <- types.ContainerResult{Identifier: options.Identifier, BuildOutput: buildbuf.String(), Error: err}
		return
	}
	defer out.Close()

	var outbuf bytes.Buffer
	_, err = io.Copy(&outbuf, out)
	if err != nil {
		ch <- types.ContainerResult{Identifier: options.Identifier, BuildOutput: buildbuf.String(), Error: err}
		return

	}

	slog.Info("Container finished, creating reports", "identifier", options.Identifier.ToString())

	ch <- types.ContainerResult{
		Identifier:  options.Identifier,
		Output:      outbuf.String(),
		BuildOutput: buildbuf.String(),
		Error:       nil,
	}
}
