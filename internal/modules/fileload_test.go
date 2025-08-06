package modules

import (
	"strings"
	"testing"
)

func TestMetaFileLoad(t *testing.T) {
	res, err := loadMeta("./tests/modules/meta.json")
	if err != nil {
		t.Errorf("Failed to load meta, %v", err)
	} else {
		if res.Version != 1 {
			t.Error("Version is not 1, failed to load configuration.")
		}
	}
}

func TestMetaFail(t *testing.T) {
	// Fails because no meta.json file. It will only check one directory
	_, err := loadMeta("./tests")
	if err.Error() != "stat tests/meta.json: no such file or directory" {
		t.Errorf("Error was not correct, %v", err)
	}
}

func TestMetaWrongFile(t *testing.T) {
	// Fails because the meta file is a category meta file, thus failing to unmarshall
	_, err := loadMeta("./tests/modules/net/meta.json")
	if !strings.Contains(err.Error(), "File loaded was not the correct file type") {
		t.Errorf("Incorrect error for wrong file, %v", err)
	}
}

func TestCategoryFilter(t *testing.T) {
	meta, _ := loadMeta("./tests/modules")
	res, err := loadCategories(meta, []string{"not-a-category"}, false)
	if err != nil {
		t.Errorf("Failed to load categories, %v", err)
	} else {
		// Length of categories should be 0
		if len(res) != 0 {
			t.Errorf("Categories should be 0, but is instead %d", len(res))
		}
	}
}

func TestCategoryToModule(t *testing.T) {
	cat := Category{
		Version: 1,
		Modules: []Module{
			{
				Name:    "nmap",
				Command: []string{"nmap", "-sV", "scanme.nmap.org"},
				Hash:    "we-aren't-testing-hashes-here",
			},
		},
	}
	input := []Category{cat}
	res := loadCategoryModules(input)
	if len(res) <= 0 || res[0].Name != "nmap" {
		t.Error("Category was not successfully turned into modules")
	}
}

func TestSpecifierModule(t *testing.T) {
	meta, _ := loadMeta("./tests/modules")
	cat, err := loadModuleFromSpecifier(meta, "net/nmap", false)
	if err != nil {
		t.Errorf("Failed to load specifier, %v", err)
	} else if len(cat.Modules) == 0 || cat.Modules[0].Name != "nmap" {
		t.Errorf("Module name does not match what was specified, got %v", cat)
	}
}
