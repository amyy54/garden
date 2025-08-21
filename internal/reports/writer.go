package reports

import (
	"fmt"
	"os"
	"path"

	"github.com/amyy54/garden/internal/types"
)

func writeResult(dir string, timestring string, res types.ContainerResult) error {
	dirinfo, err := os.Stat(dir)
	if err != nil {
		// Create directory if doesn't exist
		err = os.Mkdir(dir, os.ModePerm)
		if err != nil {
			return err
		}
		dirinfo, err = os.Stat(dir)
		if err != nil {
			return err
		}
	}
	if !dirinfo.IsDir() {
		return fmt.Errorf("Supplied reports path is not a directory")
	}

	mod_identifier := res.Identifier

	cat_path := path.Join(dir, timestring, mod_identifier.Category)

	_, err = os.Stat(cat_path)
	// If error, the category folder doesn't exist yet
	if err != nil {
		err = os.MkdirAll(cat_path, os.ModePerm)
		if err != nil {
			return err
		}
	}

	if len(res.BuildOutput) > 0 {
		build_output := path.Join(cat_path, fmt.Sprintf("%v_build.log", mod_identifier.Name))
		err = os.WriteFile(build_output, []byte(res.BuildOutput), 0440)
		if err != nil {
			return err
		}
	}

	if len(res.Output) > 0 {
		output := path.Join(cat_path, fmt.Sprintf("%v.txt", mod_identifier.Name))
		err = os.WriteFile(output, []byte(res.Output), 0440)
		if err != nil {
			return err
		}
	}

	if res.Error != nil {
		error_log := path.Join(cat_path, fmt.Sprintf("%v_error.log", mod_identifier.Name))
		err = os.WriteFile(error_log, []byte(res.Error.Error()), 0440)
		if err != nil {
			return err
		}
	}

	return nil
}
