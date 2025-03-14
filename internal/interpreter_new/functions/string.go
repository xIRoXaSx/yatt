package functions

import (
	"bytes"
	"strconv"
	"strings"

	"github.com/xiroxasx/fastplate/internal/common"
)

func Split(args [][]byte) (ret []byte, err error) {
	err = assertArgsLengthExact(args, 3)
	if err != nil {
		return
	}

	ind, err := strconv.Atoi(string(args[2]))
	if err != nil {
		return
	}
	v := bytes.Split(args[0], common.TrimQuotes(args[1]))
	if len(v) < ind {
		ret = v[0]
		return
	}
	ret = v[ind]
	return
}

func Repeat(args [][]byte) (ret []byte, err error) {
	err = assertArgsLengthExact(args, 2)
	if err != nil {
		return
	}

	factor, err := strconv.Atoi(string(args[1]))
	if err != nil {
		return
	}
	ret = bytes.Repeat(common.TrimQuotes(args[0]), factor)
	return
}

func Replace(args [][]byte) (ret []byte, err error) {
	err = assertArgsLengthExact(args, 3)
	if err != nil {
		return
	}

	ret = bytes.ReplaceAll(
		common.TrimQuotes(args[0]),
		common.TrimQuotes(args[1]),
		common.TrimQuotes(args[2]),
	)
	return
}

func Length(args [][]byte, globalVarLen int, localVarLenRetrieverFn func(name string) int) (ret []byte, err error) {
	err = assertArgsLengthExact(args, 1)
	if err != nil {
		return
	}

	const globalVarKey = "FASTPLATE_VARS"
	var (
		length int
		arg    = args[0]
	)
	if !bytes.HasPrefix(arg, []byte(globalVarKey)) {
		ret = []byte(strconv.Itoa(len(arg)))
		return
	}

	varFile := strings.TrimPrefix(string(arg), globalVarKey+"_")
	if varFile == globalVarKey {
		length = globalVarLen
	} else {
		length = localVarLenRetrieverFn(varFile)
	}
	ret = []byte(strconv.Itoa(length))

	return
}
