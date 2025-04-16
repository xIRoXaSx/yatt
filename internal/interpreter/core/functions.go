package core

import (
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/xiroxasx/fastplate/internal/common"
	"github.com/xiroxasx/fastplate/internal/interpreter/functions"
)

const (
	functionNameCryptSHA1   = "sha1"
	functionNameCryptSHA256 = "sha256"
	functionNameCryptSHA512 = "sha512"
	functionNameCryptMD5    = "md5"

	functionNameInternalEnv          = "env"
	functionNameInternalFileBaseName = "fbasename"
	functionNameInternalFileName     = "fname"
	functionNameInternalVar          = "var"

	functionNameMathAdd   = "add"
	functionNameMathSub   = "sub"
	functionNameMathMult  = "mult"
	functionNameMathDiv   = "div"
	functionNameMathPow   = "pow"
	functionNameMathSqrt  = "sqrt"
	functionNameMathRound = "round"
	functionNameMathCeil  = "ceil"
	functionNameMathFloor = "floor"
	functionNameMathFixed = "fixed"
	functionNameMathMax   = "max"
	functionNameMathMin   = "min"
	functionNameMathMod   = "mod"

	functionNameStringCapitalize = "capitalize"
	functionNameStringRepeat     = "repeat"
	functionNameStringReplace    = "replace"
	functionNameStringSplit      = "split"
	functionNameStringToLower    = "lower"
	functionNameStringToUpper    = "upper"
	functionNameStringLength     = "len"
)

func (c *Core) executeFunction(funcName parserFunc, fileName string, args [][]byte, additionalVars []common.Variable) (ret []byte, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("%s: %v", funcName, err)
		}
	}()

	switch strings.ToLower(funcName.string()) {

	// Crypt.
	case functionNameCryptSHA1:
		return functions.SHA1(args)
	case functionNameCryptSHA256:
		return functions.SHA256(args)
	case functionNameCryptSHA512:
		return functions.SHA512(args)
	case functionNameCryptMD5:
		return functions.MD5(args)

	// Internal.
	case functionNameInternalEnv:
		return functions.Env(args)
	case functionNameInternalFileBaseName:
		return functions.FileBaseName(fileName)
	case functionNameInternalFileName:
		return functions.FileName(fileName)
	case functionNameInternalVar:
		varSetter := func(name, value []byte) error {
			c.setLocalVar(filepath.Clean(fileName), common.NewVar(string(name), string(value)))
			return nil
		}
		if c.feb.StateIndex() > -1 {
			varSetter = func(name, value []byte) error {
				reg := strconv.Itoa(c.feb.StateIndex())
				c.setForeachVar(reg, common.NewVar(string(name), string(value)))
				return nil
			}
		}

		return functions.Var(fileName, args, additionalVars, varSetter)

	// Math.
	case functionNameMathAdd:
		return functions.Add(args)
	case functionNameMathSub:
		return functions.Sub(args)
	case functionNameMathMult:
		return functions.Mult(args)
	case functionNameMathDiv:
		return functions.Div(args)
	case functionNameMathPow:
		return functions.Pow(args)
	case functionNameMathSqrt:
		return functions.Sqrt(args)
	case functionNameMathRound:
		return functions.Round(args)
	case functionNameMathCeil:
		return functions.Ceil(args)
	case functionNameMathFloor:
		return functions.Floor(args)
	case functionNameMathFixed:
		return functions.Fixed(args)
	case functionNameMathMax:
		return functions.Max(args)
	case functionNameMathMin:
		return functions.Min(args)
	case functionNameMathMod:
		return functions.Mod(args)

	// String.
	case functionNameStringCapitalize:
		return functions.Capitalize(args)
	case functionNameStringRepeat:
		return functions.Repeat(args)
	case functionNameStringReplace:
		return functions.Replace(args)
	case functionNameStringSplit:
		return functions.Split(args)
	case functionNameStringToLower:
		return functions.ToLower(args)
	case functionNameStringToUpper:
		return functions.ToUpper(args)
	case functionNameStringLength:
		return functions.Length(args, len(c.varRegistryGlobal.entries), func(name string) int {
			return len(c.varRegistryLocal.entries[strings.ToLower(name)])
		})

	default:
		err = errors.New("unknown function")
		return
	}
}
