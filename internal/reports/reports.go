package reports

import (
	"path"
	"strings"
	"time"

	"github.com/amyy54/garden/internal/types"
)

func CreateReportsOutput(dir string, t time.Time, result types.ContainerResult) (string, error) {
	timestring := t.Format(time.RFC822)
	timestring = strings.ReplaceAll(timestring, ":", "-") // Thanks, Windows
	timestring = strings.ReplaceAll(timestring, " ", "_") // Thanks, Linux
	err := writeResult(dir, timestring, result)
	return path.Join(dir, timestring), err
}
