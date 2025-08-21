package reports_test

import (
	"fmt"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/amyy54/garden/internal/reports"
	"github.com/amyy54/garden/internal/types"
)

func TestReportOutput(t *testing.T) {
	result := types.ContainerResult{
		Identifier:  types.Identifier{Category: "net", Name: "hello"},
		BuildOutput: "Hello\nThis is a build script\n",
		Output:      "Hello World!\n",
		Error:       fmt.Errorf("An error occurred :(\n"),
	}
	report_path, err := reports.CreateReportsOutput("./tmp_tests", time.Now(), result)
	if err != nil {
		t.Errorf("An error occurred: %v", err)
	}

	output_path := path.Join(report_path, "net", "hello.txt")

	_, err = os.Stat(output_path)
	if err != nil {
		t.Errorf("File \"hello.txt\" does not exist in the correct path")
	}

	file, err := os.ReadFile(output_path)
	if err != nil {
		t.Errorf("Failed to read file, %v", err)
	}

	if !strings.Contains(string(file), "Hello World!") {
		t.Errorf("File does not contain the correct content, %v", string(file))
	}

	err = os.RemoveAll("./tmp_tests")
	if err != nil {
		t.Errorf("Failed to cleanup tests, %v", err)
	}
}
