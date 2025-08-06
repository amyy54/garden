package types

import (
	"fmt"
	"strings"
)

func (i *Identifier) ToString() string {
	return fmt.Sprintf("%v/%v", i.Category, i.Name)
}

func ToModIdentifier(str string) (Identifier, error) {
	cat, mod, found := strings.Cut(str, "/")
	if !found {
		return Identifier{}, fmt.Errorf("Path did not specify a separator between category and module. Even on Windows, the separator should be \"/\". Example: net/nmap")
	} else {
		return Identifier{Category: cat, Name: mod}, nil
	}
}
