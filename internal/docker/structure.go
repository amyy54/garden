package docker

import (
	"time"

	"github.com/amyy54/garden/internal/types"
)

type DockerOptions struct {
	Version        string
	DockerfilePath string
	Identifier     types.Identifier
	Command        []string
}

type InspectHost struct {
	Host          string
	SkipTLSVerify bool
}

type Inspect struct {
	Name      string
	Endpoints map[string]InspectHost
}

type ClientOptions struct {
	Runner    string
	IsContext bool
}

type ModArg struct {
	Module types.Identifier
	Key    string
	Value  string
}

type ModArgs []ModArg

func (mal *ModArgs) FindAll(identifier types.Identifier) []ModArg {
	var res []ModArg
	for _, arg := range *mal {
		// TODO Fix this. It's gross
		if identifier.Category == "single" {
			if identifier.Name == arg.Module.Name {
				res = append(res, arg)
			}
		} else {
			if identifier.ToString() == arg.Module.ToString() {
				res = append(res, arg)
			}
		}
	}
	return res
}

type RunOptions struct {
	Target    string
	Time      time.Time
	ReportDir string
	Args      ModArgs
	Version   string
}
