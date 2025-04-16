package core

import (
	"fmt"
	"os"
	"path/filepath"
)

func (c *Core) importPath(pd *PreprocessorDirective) (err error) {
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

	interFile := InterpreterFile{
		Name:   path,
		RC:     importFile,
		Writer: pd.buf,
	}
	err = c.interpret(interFile, pd.indent)
	if err != nil {
		return err
	}
	return
}
