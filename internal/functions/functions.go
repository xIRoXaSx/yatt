package functions

import (
	"bytes"
	"fmt"
	"strconv"
)

func assertArgsLengthExact(args [][]byte, num int) (err error) {
	if len(args) != num {
		return fmt.Errorf("length assertion: exactly %d args required", num)
	}
	return
}

func assertArgsLengthAtLeast(args [][]byte, num int) (err error) {
	if len(args) < num {
		return fmt.Errorf("length assertion: at least %d args required", num)
	}
	return
}

func parseFloats(args [][]byte) (values []float64, err error) {
	values = make([]float64, len(args))
	for i := range args {
		values[i], err = strconv.ParseFloat(string(bytes.TrimSpace(args[i])), 64)
		if err != nil {
			err = fmt.Errorf("args='%s': %v", string(bytes.Join(args, []byte(", "))), err)
			return
		}
	}
	return
}

func floatToBytes(v float64) []byte {
	return []byte(strconv.FormatFloat(v, 'f', -1, 64))
}
