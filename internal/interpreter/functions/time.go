package functions

import (
	"errors"
	"time"
)

func Now(args [][]byte) (ret []byte, err error) {
	if len(args) != 1 {
		err = errors.New("exactly 1 arg expected")
		return
	}

	ret = []byte(time.Now().Format(string(args[0])))
	return
}
