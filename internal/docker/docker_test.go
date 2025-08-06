package docker_test

import (
	"strings"
	"testing"
	"time"

	"github.com/amyy54/garden/internal/docker"
	"github.com/amyy54/garden/internal/types"
)

func TestRunCategories(t *testing.T) {
	mod := types.ContainerModule{
		Identifier: types.Identifier{Name: "hello", Category: "single"},
		Dockerfile: "./tests/hello.Dockerfile",
		Command:    []string{"echo", "<TARGET>"},
	}
	res, err := docker.RunCategories(docker.ClientOptions{IsContext: false, Runner: ""}, []types.ContainerModule{mod}, docker.RunOptions{Version: "v0.0.1", Target: "hello world!", Time: time.Now(), ReportDir: ""})
	if err != nil {
		t.Errorf("Failed to run docker, %v", err)
	} else if len(res) > 0 && res[0].Error != nil {
		t.Errorf("Failed to run docker container, %v", res[0].Error)
	} else if len(res) == 0 || strings.TrimSpace(res[0].Output) != "hello world!" {
		t.Errorf("Output did not format correctly, %v", res[0].Output)
	}
}
