package modules_test

import (
	"testing"

	"github.com/amyy54/garden/internal/modules"
)

func TestLoadModules(t *testing.T) {
	modres, err := modules.LoadModules("./tests/modules", modules.ModuleOptions{
		Categories:   []string{"net"},
		Modules:      []string{"web/nuclei"},
		IgnoreHashes: false,
	})
	if err != nil {
		t.Errorf("Failed to load modules, %v", err)
	} else if len(modres) == 0 || modres[0].Identifier.Name != "nmap" {
		t.Errorf("No modules were found, or the first module isn't nmap")
	}
}
