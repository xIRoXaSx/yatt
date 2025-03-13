package interpreter

import (
	"bytes"
	"fmt"
)

const (
	directiveNameForeach    = "foreach"
	directiveNameForeachEnd = "foreachend"
	directiveNameIgnore     = "ignore"
	directiveNameIgnoreEnd  = "ignoreend"
	directiveNameImport     = "import"
	directiveNameVariable   = "var"
)

type preprocessorDirective struct {
	name     string
	fileName string
	args     [][]byte
	indent   []byte
	buf      *bytes.Buffer
}

func (s *state) preprocess(pd *preprocessorDirective, lineDisplayNum int, importPathFunc func(pd *preprocessorDirective) error) (err error) {
	callID := fmt.Sprintf("%s: %d", pd.name, lineDisplayNum)
	defer func() {
		if err != nil {
			err = fmt.Errorf("%s: %v", callID, err)
			return
		}
	}()

	switch pd.name {
	case directiveNameForeach:
		return s.foreachStart(pd)

	case directiveNameForeachEnd:
		return s.foreachEnd(pd)

	case directiveNameIgnore:
		return s.ignoreStart(pd)

	case directiveNameIgnoreEnd:
		return s.ignoreEnd(pd)

	case directiveNameImport:
		return importPathFunc(pd)

	case directiveNameVariable:
		return s.setLocalVarByArgs(pd.fileName, pd.args)

	default:
		return fmt.Errorf("unknown preprocessor directive: %s", pd.name)
	}
}
