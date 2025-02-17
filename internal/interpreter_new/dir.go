package interpreter

import (
	"os"
	"path/filepath"
	"strings"
)

// runDirMode runs the interpreter for each file inside the given path.
func (i *Interpreter) runDirMode(inPath, outPath string) (err error) {
	const dirPerm = os.FileMode(0700)

	err = os.MkdirAll(inPath, dirPerm)
	if err != nil {
		return
	}

	err = filepath.WalkDir(inPath, func(inPath string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		dest := strings.ReplaceAll(inPath, inPath, outPath)
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
