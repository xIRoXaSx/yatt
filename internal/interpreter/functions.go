package interpreter

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/xiroxasx/fastplate/internal/common"
	"github.com/xiroxasx/fastplate/internal/interpreter/functions"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const (
	functionLower      = "lower"
	functionUpper      = "upper"
	functionCapitalize = "cap"
	functionAddition   = "add"
	functionSubtract   = "sub"
	functionDivide     = "div"
	functionMultiply   = "mult"
	functionMax        = "max"
	functionMin        = "min"
	functionModulus    = "mod"
	functionModulusMin = "modmin"
	functionFloor      = "floor"
	functionCeil       = "ceil"
	functionRound      = "round"
	functionToFixed    = "fixed"
	functionSha1       = "sha1"
	functionSha256     = "sha256"
	functionSha512     = "sha512"
	functionShaMd5     = "md5"
	functionSplit      = "split"
	functionRepeat     = "repeat"
	functionReplace    = "replace"
	functionLength     = "len"
	functionVar        = "var"
)

func (i *Interpreter) executeFunction(function string, args [][]byte, fileName string, additionalVars []common.Var) (ret []byte, err error) {
	if len(args) == 0 {
		err = fmt.Errorf("%v: func statement needs at least one arg", function)
		return
	}

	switch function {
	case functionLower:
		ret = bytes.ToLower(args[0])

	case functionUpper:
		ret = bytes.ToUpper(args[0])

	case functionCapitalize:
		ret = cases.Title(language.English, cases.NoLower).Bytes(args[0])

	case functionAddition:
		ret, err = functions.Add(function, args)

	case functionSubtract:
		ret, err = functions.Sub(function, args)

	case functionMultiply:
		ret, err = functions.Mult(function, args)

	case functionDivide:
		ret, err = functions.Div(function, args)

	case functionMax:
		ret, err = functions.Max(function, args)

	case functionMin:
		ret, err = functions.Min(function, args)

	case functionModulus:
		ret, err = functions.Mod(function, args)

	case functionModulusMin:
		ret, err = functions.ModMin(function, args)

	case functionFloor:
		ret, err = functions.Floor(function, args)

	case functionCeil:
		ret, err = functions.Ceil(function, args)

	case functionRound:
		ret, err = functions.Round(function, args)

	case functionToFixed:
		ret, err = functions.Fixed(function, args)

	case functionSha1:
		ret, err = functions.Sha1(function, args)

	case functionSha256:
		ret, err = functions.Sha256(function, args)

	case functionSha512:
		ret, err = functions.Sha512(function, args)

	case functionShaMd5:
		ret, err = functions.Md5(function, args)

	case functionSplit:
		ret, err = functions.Split(function, args)

	case functionRepeat:
		ret, err = functions.Repeat(function, args)

	case functionReplace:
		ret, err = functions.Replace(function, args)

	case functionLength:
		if len(args) != 1 {
			err = fmt.Errorf("%s: exactly 1 arg expected", function)
			return
		}

		var (
			l    int
			arg0 = args[0]
		)
		if bytes.HasPrefix(arg0, []byte(foreachUnscopedVars)) {
			varFile := strings.TrimPrefix(string(arg0), foreachUnscopedVars+"_")
			if varFile == foreachUnscopedVars {
				l = len(i.state.unscopedVars)
			} else {
				l = i.state.unscopedVarIndexes[strings.ToLower(varFile)].len
			}
		} else {
			l = len(arg0)
		}
		ret = []byte(fmt.Sprint(l))

	case functionVar:
		var vars [][]byte
		vars, err = functions.Var(function, args, additionalVars)
		if err != nil {
			return
		}
		i.setScopedVar(fileName, [][]byte{vars[0], []byte("="), vars[1]})
	}
	return
}
