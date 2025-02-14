package interpreter

import (
	"errors"
	"fmt"
	"strings"

	"github.com/xiroxasx/fastplate/internal/common"
	"github.com/xiroxasx/fastplate/internal/interpreter_new/functions"
)

const (
	functionNameAdd   = "add"
	functionNameSub   = "sub"
	functionNameMult  = "mult"
	functionNameDiv   = "div"
	functionNamePow   = "pow"
	functionNameSqrt  = "sqrt"
	functionNameRound = "round"
	functionNameCeil  = "ceil"
	functionNameFloor = "floor"
	functionNameFixed = "fixed"
	functionNameMax   = "max"
	functionNameMin   = "min"
	functionNameMod   = "mod"
)

func (i *Interpreter) executeFunction(funcName interpreterFunc, fileName string, args [][]byte, additionalVars []common.Variable) (ret []byte, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("%s: %v", funcName, err)
		}
	}()
	switch strings.ToLower(funcName.string()) {

	// Math.
	case functionNameAdd:
		return functions.Add(args)
	case functionNameSub:
		return functions.Sub(args)
	case functionNameMult:
		return functions.Mult(args)
	case functionNameDiv:
		return functions.Div(args)
	case functionNamePow:
		return functions.Pow(args)
	case functionNameSqrt:
		return functions.Sqrt(args)
	case functionNameRound:
		return functions.Round(args)
	case functionNameCeil:
		return functions.Ceil(args)
	case functionNameFloor:
		return functions.Floor(args)
	case functionNameFixed:
		return functions.Fixed(args)
	case functionNameMax:
		return functions.Max(args)
	case functionNameMin:
		return functions.Min(args)
	case functionNameMod:
		return functions.Mod(args)
	default:
		err = errors.New("unknown function")
		return
	}
}
