package functions

import (
	"errors"
	"time"

	"github.com/xiroxasx/godate"
)

func Now(args [][]byte) (ret []byte, err error) {
	if len(args) != 1 {
		err = errors.New("exactly 1 arg expected")
		return
	}

	date := godate.New(time.Now())
	ret = []byte(date.Format(string(args[0])))
	return
}
