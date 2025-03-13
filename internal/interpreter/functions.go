package interpreter

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xiroxasx/fastplate/internal/common"
	"github.com/xiroxasx/fastplate/internal/interpreter/functions"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const (
	functionEnvVar       = "env"
	functionFileBaseName = "fbasename"
	functionFileName     = "fname"
	functionLower        = "lower"
	functionUpper        = "upper"
	functionCapitalize   = "cap"
	functionAddition     = "add"
	functionSubtract     = "sub"
	functionDivide       = "div"
	functionMultiply     = "mult"
	functionPower        = "pow"
	functionSquareRoot   = "sqrt"
	functionMax          = "max"
	functionMin          = "min"
	functionModulus      = "mod"
	functionModulusMin   = "modmin"
	functionFloor        = "floor"
	functionCeil         = "ceil"
	functionRound        = "round"
	functionToFixed      = "fixed"
	functionSha1         = "sha1"
	functionSha256       = "sha256"
	functionSha512       = "sha512"
	functionShaMd5       = "md5"
	functionNow          = "now"
	functionSplit        = "split"
	functionRepeat       = "repeat"
	functionReplace      = "replace"
	functionLength       = "len"
	functionVar          = "var"
)

func (i *Interpreter) executeFunction(function string, args [][]byte, fileName string, additionalVars []common.Variable) (ret []byte, err error) {
	if len(args) == 0 {
		err = fmt.Errorf("%v: func statement needs at least one arg", function)
		return
	}

	defer func() {
		if err != nil {
			err = fmt.Errorf("%s: %v: %v", fileName, function, err)
		}
	}()

	switch function {
	case functionFileName:
		ret = []byte(fileName)

	case functionEnvVar:
		ret = []byte(os.Getenv(string(args[0])))

	case functionFileBaseName:
		ret = []byte(filepath.Base(fileName))

	case functionLower:
		ret = bytes.ToLower(args[0])

	case functionUpper:
		ret = bytes.ToUpper(args[0])

	case functionCapitalize:
		ret = cases.Title(language.English, cases.NoLower).Bytes(args[0])

	case functionAddition:
		ret, err = functions.Add(args)

	case functionSubtract:
		ret, err = functions.Sub(args)

	case functionMultiply:
		ret, err = functions.Mult(args)

	case functionPower:
		ret, err = functions.Pow(args)

	case functionSquareRoot:
		ret, err = functions.Sqrt(args)

	case functionDivide:
		ret, err = functions.Div(args)

	case functionMax:
		ret, err = functions.Max(args)

	case functionMin:
		ret, err = functions.Min(args)

	case functionModulus:
		ret, err = functions.Mod(args)

	case functionModulusMin:
		ret, err = functions.ModMin(args)

	case functionFloor:
		ret, err = functions.Floor(args)

	case functionCeil:
		ret, err = functions.Ceil(args)

	case functionRound:
		ret, err = functions.Round(args)

	case functionToFixed:
		ret, err = functions.Fixed(args)

	case functionSha1:
		ret, err = functions.Sha1(args)

	case functionSha256:
		ret, err = functions.Sha256(args)

	case functionSha512:
		ret, err = functions.Sha512(args)

	case functionShaMd5:
		ret, err = functions.Md5(args)

	case functionNow:
		ret, err = functions.Now(args)

	case functionSplit:
		ret, err = functions.Split(args)

	case functionRepeat:
		ret, err = functions.Repeat(args)

	case functionReplace:
		ret, err = functions.Replace(args)

	case functionLength:
		ret, err = functions.Length(args, foreachUnscopedVars, i.state.unscopedVars, func(name string) int {
			return i.state.unscopedVarIndexes[strings.ToLower(name)].len
		})

	case functionVar:
		err = functions.SetScopedVar(i, fileName, args, additionalVars)
	}
	return
}
