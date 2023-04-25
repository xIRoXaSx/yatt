package importer

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"strconv"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const (
	functionLower      = "lower"
	functionUpper      = "upper"
	functionCapitalize = "cap"
	functionModulus    = "mod"
)

func (i *Interpreter) executeFunction(function string, args [][]byte) (ret []byte, err error) {
	if len(args) == 0 {
		err = fmt.Errorf("%v: func statement needs at least one char", function)
		return
	}

	switch function {
	case functionLower:
		ret = bytes.ToLower(args[0])

	case functionUpper:
		ret = bytes.ToUpper(args[0])

	case functionCapitalize:
		ret = cases.Title(language.English, cases.NoLower).Bytes(args[0])

	case functionModulus:
		var (
			m1 float64
			m2 float64
		)
		if len(args) != 2 {
			err = errors.New("modulus args must have exactly 2 args")
			return
		}
		m1, err = strconv.ParseFloat(string(bytes.TrimSpace(args[0])), 32)
		if err != nil {
			return
		}
		m2, err = strconv.ParseFloat(string(bytes.TrimSpace(args[1])), 32)
		if err != nil {
			return
		}
		ret = []byte(fmt.Sprint(math.Mod(m1, m2)))
	}
	return
}
