package modules

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"slices"
	"strings"

	"github.com/go-playground/validator/v10"

	"github.com/amyy54/garden/internal/types"
)

// Load the meta file based on the supplied directory
// Default directory will be handled by other functions
func loadMeta(dirpath string) (Meta, error) {
	var metapath string
	var meta Meta

	// Check the string supplied to see if it exists somewhere
	dirinfo, err := os.Stat(dirpath)
	if err != nil {
		return Meta{}, err
	}

	// The string is something. This identifies if it's a file or a directory
	// If it's a file, ensure that the file is a meta.json file
	// If it's a directory, assume meta.json is in the directory supplied
	if dirinfo.IsDir() {
		metapath = path.Join(dirpath, "meta.json")
	} else {
		if strings.Contains(dirpath, "meta.json") {
			metapath = dirpath
		} else {
			return Meta{}, fmt.Errorf("Did not supply directory, and no meta.json was referenced.")
		}
	}

	// Check if metapath exists, if it doesn't, return an error
	_, err = os.Stat(metapath)
	if err != nil {
		return Meta{}, err
	}

	// Read the file. If, for some reason, it can't read it, return an error
	file, err := os.ReadFile(metapath)
	if err != nil {
		return Meta{}, err
	}

	// Marshall the data. If it worked, return the module. If it didn't, return the error
	err = json.Unmarshal(file, &meta)
	if err != nil {
		return Meta{}, fmt.Errorf("Cannot unmarshal data, %v", err)
	} else {
		meta.Path = path.Dir(metapath)
		// Validate the data. All data should be checked and validated to be correct
		validate := validator.New()
		err = validate.Struct(meta)
		if err != nil {
			return Meta{}, fmt.Errorf("File loaded was not the correct file type, %v", err)
		}
		return meta, nil
	}
}

// Takes the main meta file, filters the categories by what's needed, and returns the categories
func loadCategories(meta Meta, filter []string, ignore_hash bool) ([]Category, error) {
	var res []Category
	for _, category := range meta.Categories {
		// Before anything, check that the category was actually requested
		// But also! If any entry is a wildcard, allow all through
		if !slices.Contains(filter, "*") && !slices.Contains(filter, category.Name) {
			continue
		}

		// Obtain the category path by joining the location of the meta file with the category name
		// Ensure that meta.json is loaded here, as we can check the file all at once
		cat_path := path.Join(meta.Path, category.Name, "meta.json")

		// Check if we successfully found the meta file
		_, err := os.Stat(cat_path)
		if err != nil {
			return []Category{}, err
		}

		// Read the file. If, for some reason, it can't read it, return an error
		file, err := os.ReadFile(cat_path)
		if err != nil {
			return []Category{}, err
		}

		// Check the hash of the meta file to ensure it hasn't been changed
		if !ignore_hash {
			sum := sha256.Sum256(file)
			if fmt.Sprintf("%x", sum) != category.Hash {
				return []Category{}, fmt.Errorf("Hash failed for category file \"%s\". Calculated hash: %x", category.Name, sum)
			}
		}

		// Marshall the data. If it worked, add it to res, if it didn't, error
		var category_data Category
		err = json.Unmarshal(file, &category_data)
		if err != nil {
			return []Category{}, fmt.Errorf("Cannot unmarshal data, %v", err)
		} else {
			category_data.Name = category.Name
			// Validate that the data is correct
			validate := validator.New()
			err = validate.Struct(category_data)
			if err != nil {
				return []Category{}, fmt.Errorf("File loaded was not the correct file type, %v", err)
			}
			res = append(res, category_data)
		}
	}
	return res, nil
}

// Function to load single module from supplied input
// It returns a category with the name "single"
func loadModuleFromSpecifier(meta Meta, path string, ignore_hash bool) (Category, error) {
	// Example: net/nmap. As modules don't have a specific configuration file, this'll do
	module_id, err := types.ToModIdentifier(path)
	if err != nil {
		return Category{}, err
	}
	cat, err := loadCategories(meta, []string{module_id.Category}, ignore_hash)
	if err != nil {
		return Category{}, err
	}
	if len(cat) == 0 {
		return Category{}, fmt.Errorf("Could not find the category specified")
	}
	modules := loadCategoryModules(cat)
	index := slices.IndexFunc(modules, func(m Module) bool { return m.Name == module_id.Name })
	if index == -1 {
		return Category{}, fmt.Errorf("Could not find specified module in the provided category")
	}
	res := Category{
		Version: -5,
		Name:    fmt.Sprintf("single-%v", module_id.Category),
		Modules: []Module{modules[index]},
	}
	return res, nil
}

// Basic helper function to turn a list of categories into a list of modules
func loadCategoryModules(categories []Category) []Module {
	var res []Module
	for _, category := range categories {
		res = append(res, category.Modules...)
	}
	return res
}

func getDockerfilePath(meta Meta, category Category, mod Module, ignore_hash bool) (string, error) {
	cat_name := category.Name
	// Checking if its a single category
	if category.Version == -5 {
		cat_split := strings.Split(cat_name, "-")
		if len(cat_split) > 1 {
			cat_name = cat_split[1]
		}
	}
	dockerfilepath := path.Join(meta.Path, cat_name, fmt.Sprintf("%v.Dockerfile", mod.Name))
	_, err := os.Stat(dockerfilepath)
	if err != nil {
		return "", err
	}

	file, err := os.ReadFile(dockerfilepath)
	if err != nil {
		return "", err
	}

	// Check the hash of the meta file to ensure it hasn't been changed
	if !ignore_hash {
		sum := sha256.Sum256(file)
		if fmt.Sprintf("%x", sum) != mod.Hash {
			return "", fmt.Errorf("Hash failed for module file \"%s\". Calculated hash: %x", mod.Name, sum)
		}
	}

	return dockerfilepath, nil
}
