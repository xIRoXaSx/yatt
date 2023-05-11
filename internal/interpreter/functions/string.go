package functions

import (
	"bytes"
	"fmt"
	"strconv"
)

func Split(fn string, args [][]byte) (ret []byte, err error) {
	if len(args) != 3 {
		err = fmt.Errorf("%s: exactly 3 args expected", fn)
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

func Repeat(fn string, args [][]byte) (ret []byte, err error) {
	if len(args) != 2 {
		err = fmt.Errorf("%s: exactly 2 args expected", fn)
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

func Replace(fn string, args [][]byte) (ret []byte, err error) {
	if len(args) != 3 {
		err = fmt.Errorf("%s: exactly 3 args expected", fn)
		return
	}

	if err != nil {
		return nil, err
	}
	ret = bytes.ReplaceAll(TrimQuotes(args[0]), TrimQuotes(args[1]), TrimQuotes(args[2]))
	return
}
