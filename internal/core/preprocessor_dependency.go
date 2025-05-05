package core

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
)

func (c *Core) ImportPathCheckCyclicDependencies(startPath string) (err error) {
	file, err := os.Open(startPath)
	if err != nil {
		return
	}
	defer func() {
		cErr := file.Close()
		if err == nil {
			err = cErr
			return
		}
		c.l.Err(cErr).Str("file", startPath).Msg("closing file reader")
	}()

	var (
		// Limits reads to 65536 bytes per line.
		scanner = bufio.NewScanner(file)
		ln      int
	)
	for scanner.Scan() {
		err = scanner.Err()
		if err != nil {
			return
		}

		ln++
		line := bytes.TrimSpace(scanner.Bytes())
		prefix := c.matchedPrefixToken(line)
		if len(prefix) == 0 {
			continue
		}

		statement := trimLine(line, prefix)
		split := bytes.Split(statement, []byte{' '})
		if len(split) == 0 {
			continue
		}

		preprocessor := bytes.TrimSpace(split[0])
		if string(preprocessor) != directiveNameImport {
			continue
		}

		pd := &PreprocessorDirective{
			name:     string(preprocessor),
			fileName: startPath,
			args:     split[1:],
			buf:      &bytes.Buffer{},
			lineNum:  ln,
		}
		err = c.walkDependency(pd)
		if err != nil {
			return
		}
	}
	return
}

func (c *Core) walkDependency(pd *PreprocessorDirective) (err error) {
	if len(pd.args) != 1 {
		return errDependencyUnknownSyntax
	}

	sourcePath := pd.fileName
	importingPath := filepath.Clean(string(pd.args[0]))
	c.depsResolver.addDependency(sourcePath, importingPath)
	importFile, err := os.Open(importingPath)
	if err != nil {
		return
	}
	defer func() {
		cErr := importFile.Close()
		if cErr != nil {
			if err == nil {
				err = cErr
				return
			}
			c.l.Err(cErr).Str("path", importingPath).Msg("closing dependency file on defer")
		}
	}()

	// Limits reads to 65536 bytes per line.
	scanner := bufio.NewScanner(importFile)
	for scanner.Scan() {
		err = scanner.Err()
		if err != nil {
			return
		}

		line := bytes.TrimSpace(scanner.Bytes())
		prefix := c.matchedPrefixToken(line)
		if len(prefix) == 0 {
			continue
		}

		// Check if line contains import statement.
		importSplit := bytes.Split(line, fmt.Appendf(nil, "%s %s", prefix, preprocessorImportName))
		if len(importSplit) != 2 {
			continue
		}

		// We got the path to import.
		// Now we need to check if the file does also import the current one.
		foundImportPath := bytes.TrimSpace(importSplit[1])
		importPathOfLine := filepath.Clean(string(bytes.TrimSpace(foundImportPath)))
		cyclic := c.depsResolver.CheckForCyclicDependencies(sourcePath, importPathOfLine)
		if cyclic {
			return fmt.Errorf("%w: %s -> %s", errDependencyCyclic, sourcePath, importingPath)
		}

		err = c.walkDependency(&PreprocessorDirective{
			fileName: importingPath,
			args:     [][]byte{[]byte(importPathOfLine)},
		})
		if err != nil {
			return
		}
	}

	return
}
