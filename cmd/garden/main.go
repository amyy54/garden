package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/amyy54/garden/internal/docker"
	"github.com/amyy54/garden/internal/modules"
)

var (
	Version     string
	VersionLong string

	ModulesPath string
	ReportsPath string
)

func main() {
	var target string         // -target
	var docker_host string    // -host, default docker context
	var docker_context string // -context
	var categories string     // -category
	var single string         // -single
	var reports_dir string    // -output, defualt ./reports/
	var modules_dir string    // -modules, default ./modules/
	var modargs string        // -modargs

	var list bool        // -list
	var skip_hashes bool // -ignore-hash
	var v bool
	var vv bool
	var version bool

	log.SetFlags(log.LstdFlags)

	if len(ModulesPath) == 0 {
		ModulesPath = "./modules"
	}
	if len(ReportsPath) == 0 {
		ReportsPath = "./reports"
	}

	flag.StringVar(&target, "target", "", "Host/IP to scan")
	flag.StringVar(&docker_host, "host", "", "Docker endpoint to use")
	flag.StringVar(&docker_context, "context", "", "Docker context to use. Overrides -host")
	flag.StringVar(&categories, "category", "", "Categories to execute, separated by commas")
	flag.StringVar(&single, "single", "", "Individual modules to execute, separated by commas")
	flag.StringVar(&modules_dir, "modules", ModulesPath, "Directory to look for modules")
	flag.StringVar(&reports_dir, "output", ReportsPath, "Directory to output results")
	flag.StringVar(&modargs, "modargs", "", "Additional module-specific arguments")

	flag.BoolVar(&list, "list", false, "List loaded modules and their information")
	flag.BoolVar(&skip_hashes, "ignore-hash", false, "Skips checking module hashes")
	flag.BoolVar(&v, "v", false, "Increase verbosity to info")
	flag.BoolVar(&vv, "vv", false, "Increase verbosity to debug")
	flag.BoolVar(&version, "version", false, "Print the version and exit")

	flag.Parse()

	if version {
		if v || vv {
			fmt.Printf("garden: %s\n", VersionLong)
		} else {
			fmt.Printf("garden: %s\n", Version)
		}
		os.Exit(0)
	}

	if vv {
		slog.SetLogLoggerLevel(slog.LevelDebug)
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	} else if v {
		slog.SetLogLoggerLevel(slog.LevelInfo)
	} else {
		slog.SetLogLoggerLevel(slog.LevelWarn)
	}

	modules_path, err := filepath.Abs(modules_dir)
	if err != nil {
		log.Fatalf("Could not resolve modules path, %v", err)
	}

	reports_path, err := filepath.Abs(reports_dir)
	if err != nil {
		log.Fatalf("Could not resolve reports path, %v", err)
	}

	var cats []string
	var mods []string

	if list {
		cats = []string{"*"}
	} else if len(categories) == 0 && len(single) == 0 {
		log.Fatal("No categories/modules were supplied, cannot continue")
	} else {
		if len(categories) > 0 {
			cats = strings.Split(categories, ",")
		}
		if len(single) > 0 {
			mods = strings.Split(single, ",")
		}
	}

	loaded_modules, err := modules.LoadModules(modules_path, modules.ModuleOptions{
		Categories:   cats,
		Modules:      mods,
		IgnoreHashes: skip_hashes,
	})
	if err != nil {
		log.Fatalf("Failed to load modules, %v", err)
	}

	if list {
		for _, mod := range loaded_modules {
			fmt.Printf("%s\t%v\n", mod.Identifier.ToString(), mod.Command)
		}
		os.Exit(0)
	}

	if len(target) == 0 {
		log.Fatal("No target supplied, cannot continue")
	}

	starting_time := time.Now()

	var docker_client docker.ClientOptions
	if len(docker_context) != 0 {
		docker_client = docker.ClientOptions{
			IsContext: true,
			Runner:    docker_context,
		}
	} else {
		docker_client = docker.ClientOptions{
			IsContext: false,
			Runner:    docker_host,
		}
	}

	var args []docker.ModArg
	if len(modargs) > 0 {
		args, err = docker.ParseModArgs(modargs)
		if err != nil {
			log.Fatalf("Additional module arguments could not be parsed, %v", err)
		}
	}

	version_docker := Version
	if len(version_docker) == 0 {
		version_docker = "v0.0.1"
	}

	slog.Debug("Parsed module arguments", "args", args)

	slog.Warn("Starting docker and all modules")

	_, err = docker.RunCategories(docker_client, loaded_modules, docker.RunOptions{
		Target:    target,
		Time:      starting_time,
		ReportDir: reports_path,
		Args:      args,
		Version:   version_docker,
	})
	if err != nil {
		log.Fatalf("Failed to run docker, %v", err)
	}

	os.Exit(0)
}
