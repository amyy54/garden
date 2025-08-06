package types_test

import (
	"testing"

	"github.com/amyy54/garden/internal/types"
)

func TestToString(t *testing.T) {
	input := types.Identifier{Category: "spiderweb", Name: "muffet"}
	res := input.ToString()
	if res != "spiderweb/muffet" {
		t.Errorf("Failed to format identifier. Res: %v", res)
	}
}

func TestToIdentifier(t *testing.T) {
	res, err := types.ToModIdentifier("spiderweb/muffet")
	if err != nil {
		t.Errorf("Failed to parse identifier, %v", err)
	}
	if res.Name != "muffet" {
		t.Errorf("Result does not match input, %v", res)
	}
}

func TestFailIdentifier(t *testing.T) {
	what, err := types.ToModIdentifier("thiswillfail")
	if err == nil {
		t.Errorf("Somehow gave a positive result for this, %v", what)
	}
}
