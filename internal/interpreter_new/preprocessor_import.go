package interpreter

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
)

const (
	preprocessorImportName = "import"
)

func (i *Interpreter) importPath(pd *preprocessorDirective) (err error) {
	if len(pd.args) != 1 {
		err = fmt.Errorf("unknown syntax: %s <file path>", preprocessorImportName)
		return
	}

	// Open the import file.
	// The interpret method will close it afterwards.
	path := filepath.Clean(string(pd.args[0]))
	importFile, err := os.Open(path)
	if err != nil {
		return
	}

	interFile := interpreterFile{
		name: filepath.Clean(pd.fileName),
		rc:   importFile,
	}
	err = i.interpret(interFile, pd.indent)
	if err != nil {
		return err
	}
	return
}

func (i *Interpreter) importPathCheckCyclicDependencies(startPath string) (err error) {
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
		i.l.Err(cErr).Str("file", startPath).Msg("closing file reader")
	}()

	var (
		// Limits reads to 65536 bytes per line.
		scanner = bufio.NewScanner(file)
	)
	for scanner.Scan() {
		err = scanner.Err()
		if err != nil {
			return
		}

		line := bytes.TrimSpace(scanner.Bytes())
		prefix := i.matchedPrefixToken(line)
		if len(prefix) == 0 {
			continue
		}

		statement := i.trimLine(line, prefix)
		split := bytes.Split(statement, []byte{' '})
		if len(split) == 0 {
			err = errDependencyUnknownSyntax
			return
		}

		pd := &preprocessorDirective{
			name:     string(split[0]),
			fileName: startPath,
			args:     split[1:],
			buf:      &bytes.Buffer{},
		}
		err = i.readDependency(pd)
		if err != nil {
			return
		}
	}
	return
}

func (i *Interpreter) readDependency(pd *preprocessorDirective) (err error) {
	if len(pd.args) != 1 {
		return errDependencyUnknownSyntax
	}

	path := filepath.Clean(string(pd.args[0]))
	importFile, err := os.Open(path)
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
			i.l.Err(cErr).Str("path", path).Msg("closing dependency file on defer")
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
		prefix := i.matchedPrefixToken(line)
		if len(prefix) == 0 {
			continue
		}

		// Check if line contains import statement.
		importSplit := bytes.Split(line, []byte(fmt.Sprintf("%s %s", prefix, preprocessorImportName)))
		if len(importSplit) == 0 {
			continue
		} else if len(importSplit) < 2 {
			// Incorrect usage of the import statement.
			return errDependencyUnknownSyntax
		}

		// We got the path to import.
		// Now we need to check if the file does also import the current one.
		err = func() (err error) {
			importPath := filepath.Clean(string(bytes.TrimSpace(importSplit[1])))
			i.state.Lock()
			defer i.state.Unlock()

			i.state.depsResolver.addDependency(pd.fileName, importPath)
			cyclic := i.state.depsResolver.dependenciesAreCyclic(pd.fileName, importPath)
			if cyclic {
				return fmt.Errorf("cyclic dependency")
			}
			return
		}()
		if err != nil {
			return
		}
	}

	return
}
