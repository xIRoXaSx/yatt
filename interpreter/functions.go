package importer

import (
	"bytes"
	"fmt"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const (
	functionLower      = "lower"
	functionUpper      = "upper"
	functionCapitalize = "cap"
)

func (i *Interpreter) executeFunction(function string, arg []byte) (ret []byte, err error) {
	if len(arg) == 0 {
		err = fmt.Errorf("%v: func statement needs at least one char", function)
		return
	}

	switch function {
	case functionLower:
		ret = bytes.ToLower(arg)

	case functionUpper:
		ret = bytes.ToUpper(arg)

	case functionCapitalize:
		ret = cases.Title(language.English, cases.NoLower).Bytes(arg)
	}
	return
}
