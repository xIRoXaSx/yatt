package functions

import (
	"bytes"
	"strconv"
	"strings"
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
	v := bytes.Split(args[0], trimQuotes(args[1]))
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
	ret = bytes.Repeat(trimQuotes(args[0]), factor)
	return
}

func Replace(args [][]byte) (ret []byte, err error) {
	err = assertArgsLengthExact(args, 3)
	if err != nil {
		return
	}

	ret = bytes.ReplaceAll(
		trimQuotes(args[0]),
		trimQuotes(args[1]),
		trimQuotes(args[2]),
	)
	return
}

func Length(args [][]byte, globalVarLen int, localVarLenRetrieverFn func(name string) int) (ret []byte, err error) {
	err = assertArgsLengthExact(args, 1)
	if err != nil {
		return
	}

	const globalVarKey = "GLOBAL"
	var (
		length int
		arg    = args[0]
	)
	if !bytes.HasPrefix(arg, []byte(globalVarKey)) {
		length = len(arg)
	} else {
		varFile := strings.TrimPrefix(string(arg), globalVarKey+"_")
		if varFile == globalVarKey {
			length = globalVarLen
		} else {
			length = localVarLenRetrieverFn(varFile)
		}
	}
	ret = []byte(strconv.Itoa(length))

	return
}

//
// Helper.
//

// TrimQuotes trims spaces, prefix and suffix separately to allow symbol escaping.
func trimQuotes(val []byte) (ret []byte) {
	quoteSingle := '"'
	quoteDouble := '\''
	ret = bytes.TrimSpace(val)
	ret = bytes.TrimFunc(ret, func(r rune) bool {
		return r == quoteSingle || r == quoteDouble
	})
	return
}
