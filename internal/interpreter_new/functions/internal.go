package functions

import (
	"github.com/xiroxasx/fastplate/internal/common"
)

func Var(fileName string, args [][]byte, additionalVars []common.Variable, localVarSetter func(name, value []byte) error) (ret []byte, err error) {
	err = assertArgsLengthAtLeast(args, 2)
	if err != nil {
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

	err = localVarSetter(arg0, arg1)
	return
}
