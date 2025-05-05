package functions

import (
	"time"

	"github.com/xiroxasx/godate"
)

func Now(args [][]byte) (ret []byte, err error) {
	err = assertArgsLengthExact(args, 1)
	if err != nil {
		return
	}

	date := godate.New(time.Now())
	ret = []byte(date.Format(string(args[0])))
	return
}
