package functions

import (
	"fmt"

	"github.com/xiroxasx/fastplate/internal/common"
)

func Var(fn string, args [][]byte, additionalVars []common.Var) (ret [][]byte, err error) {
	if len(args) != 2 {
		err = fmt.Errorf("%s: exactly 2 args expected", fn)
		return
	}
	// Check if any additional variable matches.
	arg0 := args[0]
	arg1 := args[1]
	for _, v := range additionalVars {
		if string(arg0) == v.Name() {
			arg1 = []byte(v.Value())
			break
		}
	}
	ret = [][]byte{arg0, arg1}
	return
}
