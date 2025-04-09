package core

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"path/filepath"
)

const (
	directiveNameForeach    = "foreach"
	directiveNameForeachEnd = "foreachend"
	directiveNameIgnore     = "ignore"
	directiveNameIgnoreEnd  = "ignoreend"
	directiveNameImport     = "import"
	directiveNameVariable   = "var"
)

type PreprocessorDirective struct {
	name     string
	fileName string
	args     [][]byte
	indent   []byte
	lineNum  int
	buf      *bytes.Buffer
}

func newPreprocessorDirective(name, fileName string, lineNum int, args [][]byte, indent []byte) *PreprocessorDirective {
	return &PreprocessorDirective{
		name:     name,
		fileName: fileName,
		args:     args,
		indent:   indent,
		lineNum:  lineNum,
		buf:      &bytes.Buffer{},
	}
}

func (p *PreprocessorDirective) WriteTo(w io.Writer) (n int64, err error) {
	return p.buf.WriteTo(w)
}

// Implements the Preprocessor interface.
func (c *Core) Preprocess(pd *PreprocessorDirective, importPathFunc func(pd *PreprocessorDirective) error) (err error) {
	return c.preprocess(importPathFunc, pd)
}

func (c *Core) preprocess(importPathFunc func(pd *PreprocessorDirective) error, pd *PreprocessorDirective) (err error) {
	callID := fmt.Sprintf("%s: %s: %d", pd.fileName, pd.name, pd.lineNum)
	defer func() {
		if err != nil {
			err = fmt.Errorf("%s: %v", callID, err)
			return
		}
	}()

	switch pd.name {
	case directiveNameForeach:
		return c.foreachStart(pd)

	case directiveNameForeachEnd:
		return c.foreachEnd(pd)

	case directiveNameIgnore:
		return c.ignoreStart(pd)

	case directiveNameIgnoreEnd:
		return c.ignoreEnd(pd)

	case directiveNameImport:
		return importPathFunc(pd)

	case directiveNameVariable:
		return c.setLocalVarByArg(filepath.Clean(pd.fileName), bytes.Join(pd.args, []byte{' '}))

	default:
		return errors.New("unknown preprocessor directive")
	}
}

type recurringToken uint8

const (
	RecurringTokenIgnore recurringToken = iota + 1
	RecurringTokenForeach
)

func (c *Core) preprocessorState(fileName string) (t recurringToken) {
	// Line does not contain one of the required prefixes.
	if c.ignoreIndex[fileName].isActive() {
		return RecurringTokenIgnore
	}
	if c.feb.IsActive() {
		return RecurringTokenForeach
	}

	return
}
