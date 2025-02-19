package interpreter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// runDirMode runs the interpreter for each file inside the given path.
func (i *Interpreter) runDirMode(sourcePath, outPath string) (err error) {
	const dirPerm = os.FileMode(0700)

	err = os.MkdirAll(sourcePath, dirPerm)
	if err != nil {
		return
	}

	err = filepath.WalkDir(sourcePath, func(inPath string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// First check if we have cyclic dependencies.
		err = i.importPathCheckCyclicDependencies(inPath)
		if err != nil {
			return fmt.Errorf("dependency check: %v", err)
		}

		dest := strings.ReplaceAll(sourcePath, inPath, outPath)
		if entry.IsDir() {
			if dest == "" {
				// The root path can be skipped.
				return nil
			}

			// Create dirs along the way.
			err = os.MkdirAll(dest, dirPerm)
			if err != nil {
				return err
			}
			return nil
		}

		err = i.writeInterpretedFile(inPath, dest)
		if err != nil {
			return err
		}

		return nil
	})
	return
}
