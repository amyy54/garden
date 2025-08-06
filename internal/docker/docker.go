package docker

import (
	"fmt"
	"log/slog"
	"reflect"
	"regexp"
	"strings"

	"github.com/amyy54/garden/internal/reports"
	"github.com/amyy54/garden/internal/types"
)

func RunCategories(runner ClientOptions, modules []types.ContainerModule, options RunOptions) ([]types.ContainerResult, error) {
	var cases []reflect.SelectCase
	var res []types.ContainerResult
	cli, err := createClient(runner)
	if err != nil {
		return []types.ContainerResult{}, err
	}
	slog.Info("Client created", "client", cli.ClientVersion())

	for _, mod := range modules {
		modargs := options.Args.FindAll(mod.Identifier)
		slog.Debug("Found the following module arguments", "modargs", modargs)

		for i, cmd := range mod.Command {
			for _, arg := range modargs {
				formatted_key := fmt.Sprintf("<%s>", arg.Key)
				if strings.Contains(cmd, formatted_key) {
					cmd = strings.ReplaceAll(cmd, formatted_key, arg.Value)
				}
			}

			if strings.Contains(cmd, "<TARGET>") {
				cmd = strings.ReplaceAll(cmd, "<TARGET>", options.Target)
			}

			mod.Command[i] = cmd
		}

		ch := make(chan types.ContainerResult)
		go runModule(ch, cli, DockerOptions{
			Version:        strings.ReplaceAll(options.Version, "v", ""),
			DockerfilePath: mod.Dockerfile,
			Identifier:     mod.Identifier,
			Command:        mod.Command,
		})

		cases = append(cases, reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(ch),
		})
	}

	for range cases {
		_, recv, ok := reflect.Select(cases)
		if ok {
			single := recv.Interface().(types.ContainerResult)
			if len(options.ReportDir) > 0 {
				_, err := reports.CreateReportsOutput(options.ReportDir, options.Time, single)
				if err != nil {
					slog.Warn("An error occurred when saving report", "error", err)
				} else {
					slog.Info("Reports generated", "identifier", single.Identifier.ToString())
				}
			}
			res = append(res, single)
		}
	}
	return res, nil
}

func ParseModArgs(args string) ([]ModArg, error) {
	slog.Debug("Module arguments passed", "args", args)
	var res []ModArg
	re := regexp.MustCompile("(.+)/(.+)-(.+)=(.+)")

	for arg := range strings.SplitSeq(args, ",") {
		matches := re.FindStringSubmatch(arg)
		if matches == nil || len(matches) <= 4 {
			return []ModArg{}, fmt.Errorf("Could not find any matches for module args")
		}
		res = append(res, ModArg{
			Module: types.Identifier{Category: matches[1], Name: matches[2]},
			Key:    matches[3],
			Value:  matches[4],
		})
	}
	return res, nil
}
