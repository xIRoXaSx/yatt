package functions

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/xiroxasx/fastplate/internal/common"
)

func Split(args [][]byte) (ret []byte, err error) {
	if len(args) != 3 {
		err = errors.New("exactly 3 args expected")
		return
	}

	var ind int
	ind, err = strconv.Atoi(string(args[2]))
	if err != nil {
		return nil, err
	}
	split := bytes.Split(args[0], TrimQuotes(args[1]))
	if len(split) < ind {
		return
	}
	ret = split[ind]
	return
}

func Repeat(args [][]byte) (ret []byte, err error) {
	if len(args) != 2 {
		err = errors.New("exactly 2 args expected")
		return
	}

	var factor int
	factor, err = strconv.Atoi(string(args[1]))
	if err != nil {
		return nil, err
	}
	ret = bytes.Repeat(TrimQuotes(args[0]), factor)
	return
}

func Replace(args [][]byte) (ret []byte, err error) {
	if len(args) != 3 {
		err = errors.New("exactly 3 args expected")
		return
	}

	ret = bytes.ReplaceAll(TrimQuotes(args[0]), TrimQuotes(args[1]), TrimQuotes(args[2]))
	return
}

func Length(args [][]byte, unscopedVarsKey string, unscoped []common.Variable, scopedVarIndexLengthFn func(name string) int) (ret []byte, err error) {
	if len(args) != 1 {
		err = fmt.Errorf("exactly 1 arg expected")
		return
	}

	var (
		l    int
		arg0 = args[0]
	)
	if bytes.HasPrefix(arg0, []byte(unscopedVarsKey)) {
		varFile := strings.TrimPrefix(string(arg0), unscopedVarsKey+"_")
		if varFile == unscopedVarsKey {
			l = len(unscoped)
		} else {
			l = scopedVarIndexLengthFn(varFile)
		}
	} else {
		l = len(arg0)
	}
	ret = []byte(strconv.Itoa(l))

	return
}
