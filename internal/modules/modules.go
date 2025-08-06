package modules

import (
	"log/slog"
	"strings"

	"github.com/amyy54/garden/internal/types"
)

func LoadModules(module_path string, options ModuleOptions) ([]types.ContainerModule, error) {
	var categories []Category
	meta, err := loadMeta(module_path)
	slog.Debug("Meta loaded", "meta", meta)
	if err != nil {
		return []types.ContainerModule{}, err
	}
	if len(options.Categories) > 0 {
		cat, err := loadCategories(meta, options.Categories, options.IgnoreHashes)
		slog.Debug("Loaded categories", "cat", cat)
		if err != nil {
			return []types.ContainerModule{}, err
		}
		categories = append(categories, cat...)
	}
	if len(options.Modules) > 0 {
		for _, spec := range options.Modules {
			cat, err := loadModuleFromSpecifier(meta, spec, options.IgnoreHashes)
			slog.Debug("Loaded single module", "mod", cat)
			if err != nil {
				return []types.ContainerModule{}, err
			}
			categories = append(categories, cat)
		}
	}

	// Now we have the list of categories, put it in a useful format
	var modres []types.ContainerModule
	for _, cat := range categories {
		cat_name := cat.Name
		if cat.Version == -5 {
			cat_name = strings.Split(cat_name, "-")[0]
		}
		for _, mod := range cat.Modules {
			dockerfile, err := getDockerfilePath(meta, cat, mod, options.IgnoreHashes)
			slog.Debug("Found dockerfile path for module", "module", mod, "dockerfile", dockerfile)
			if err != nil {
				return []types.ContainerModule{}, err
			}
			modres = append(modres, types.ContainerModule{
				Identifier: types.Identifier{Category: cat_name, Name: mod.Name},
				Dockerfile: dockerfile,
				Command:    mod.Command,
			})
		}
	}
	return modres, nil
}
