package interpreter

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"math"
	"os"
	"strconv"
	"strings"

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

func (i *Interpreter) executeFunction(function string, args [][]byte, fileName string, additionalVars []variable) (ret []byte, err error) {
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
		var (
			floats []float64
			sum    float64
		)
		if len(args) < 2 {
			err = fmt.Errorf("%s: at least 2 args expected", function)
			return
		}

		floats, err = parseFloats(args)
		if err != nil {
			return
		}
		for j := range floats {
			sum += floats[j]
		}
		ret = []byte(fmt.Sprint(sum))

	case functionSubtract:
		var (
			floats []float64
			sum    float64
		)
		if len(args) < 2 {
			err = fmt.Errorf("%s: at least 2 args expected", function)
			return
		}

		floats, err = parseFloats(args)
		if err != nil {
			return
		}
		sum = floats[0]
		for _, f := range floats[1:] {
			sum -= f
		}
		ret = []byte(fmt.Sprint(sum))

	case functionMultiply:
		var (
			floats []float64
			sum    float64
		)
		if len(args) < 2 {
			err = fmt.Errorf("%s: at least 2 args expected", function)
			return
		}

		floats, err = parseFloats(args)
		if err != nil {
			return
		}
		sum = floats[0]
		for _, f := range floats[1:] {
			sum *= f
		}
		ret = []byte(fmt.Sprint(sum))

	case functionDivide:
		var (
			floats []float64
			sum    float64
		)
		if len(args) < 2 {
			err = fmt.Errorf("%s: at least 2 args expected", function)
			return
		}

		floats, err = parseFloats(args)
		if err != nil {
			return
		}
		sum = floats[0]
		for _, f := range floats[1:] {
			sum /= f
		}
		ret = []byte(fmt.Sprint(sum))

	case functionMax:
		var floats []float64
		if len(args) < 2 {
			err = fmt.Errorf("%s: at least 2 args expected", function)
			return
		}

		floats, err = parseFloats(args)
		if err != nil {
			return
		}
		max := floats[0]
		for _, f := range floats[1:] {
			max = math.Max(max, f)
		}
		ret = []byte(fmt.Sprint(max))

	case functionMin:
		var floats []float64
		if len(args) < 2 {
			err = fmt.Errorf("%s: at least 2 args expected", function)
			return
		}

		floats, err = parseFloats(args)
		if err != nil {
			return
		}
		min := floats[0]
		for _, f := range floats[1:] {
			min = math.Min(min, f)
		}
		ret = []byte(fmt.Sprint(min))

	case functionModulus:
		var floats []float64
		if len(args) != 2 {
			err = fmt.Errorf("%s: exactly 2 args expected", function)
			return
		}

		floats, err = parseFloats(args)
		if err != nil {
			return
		}
		ret = []byte(fmt.Sprint(math.Mod(floats[0], floats[1])))

	case functionModulusMin:
		var floats []float64
		if len(args) != 3 {
			err = fmt.Errorf("%s: exactly 3 args expected", function)
			return
		}

		floats, err = parseFloats(args)
		if err != nil {
			return
		}
		mod := math.Mod(floats[0], floats[1])
		ret = []byte(fmt.Sprint(math.Max(floats[2], mod)))

	case functionFloor:
		var floats []float64
		if len(args) != 1 {
			err = fmt.Errorf("%s: exactly 1 arg expected", function)
			return
		}

		floats, err = parseFloats(args)
		if err != nil {
			return
		}
		ret = []byte(fmt.Sprint(math.Floor(floats[0])))

	case functionCeil:
		var floats []float64
		if len(args) != 1 {
			err = fmt.Errorf("%s: exactly 1 arg expected", function)
			return
		}

		floats, err = parseFloats(args)
		if err != nil {
			return
		}
		ret = []byte(fmt.Sprint(math.Ceil(floats[0])))

	case functionRound:
		var floats []float64
		if len(args) != 1 {
			err = fmt.Errorf("%s: exactly 1 arg expected", function)
			return
		}

		floats, err = parseFloats(args)
		if err != nil {
			return
		}
		ret = []byte(fmt.Sprint(math.Round(floats[0])))

	case functionToFixed:
		var floats []float64
		if len(args) != 2 {
			err = fmt.Errorf("%s: exactly 2 arg expected", function)
			return
		}

		floats, err = parseFloats(args)
		if err != nil {
			return
		}
		decPlace := floats[1] * 10
		ret = []byte(fmt.Sprint(math.Round(floats[0]*decPlace) / decPlace))

	case functionSha1:
		if len(args) != 1 {
			err = fmt.Errorf("%s: exactly 1 arg expected", function)
			return
		}

		ret, err = encodeHashToHex(sha1.New(), string(args[0]))
		if err != nil {
			return
		}

	case functionSha256:
		if len(args) != 1 {
			err = fmt.Errorf("%s: exactly 1 arg expected", function)
			return
		}

		ret, err = encodeHashToHex(sha256.New(), string(args[0]))
		if err != nil {
			return
		}

	case functionSha512:
		if len(args) != 1 {
			err = fmt.Errorf("%s: exactly 1 arg expected", function)
			return
		}

		ret, err = encodeHashToHex(sha512.New(), string(args[0]))
		if err != nil {
			return
		}

	case functionShaMd5:
		if len(args) != 1 {
			err = fmt.Errorf("%s: exactly 1 arg expected", function)
			return
		}

		ret, err = encodeHashToHex(md5.New(), string(args[0]))
		if err != nil {
			return
		}

	case functionSplit:
		if len(args) != 3 {
			err = fmt.Errorf("%s: exactly 3 args expected", function)
			return
		}

		var ind int
		ind, err = strconv.Atoi(string(args[2]))
		if err != nil {
			return nil, err
		}
		split := bytes.Split(args[0], trimQuotes(args[1]))
		if len(split) < ind {
			return
		}
		ret = split[ind]

	case functionRepeat:
		if len(args) != 2 {
			err = fmt.Errorf("%s: exactly 2 args expected", function)
			return
		}

		var factor int
		factor, err = strconv.Atoi(string(args[1]))
		if err != nil {
			return nil, err
		}
		ret = bytes.Repeat(trimQuotes(args[0]), factor)

	case functionReplace:
		if len(args) != 3 {
			err = fmt.Errorf("%s: exactly 3 args expected", function)
			return
		}

		if err != nil {
			return nil, err
		}
		ret = bytes.ReplaceAll(trimQuotes(args[0]), trimQuotes(args[1]), trimQuotes(args[2]))

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
		if len(args) != 2 {
			err = fmt.Errorf("%s: exactly 2 args expected", function)
			return
		}
		// Check if any additional variable matches.
		arg0 := args[0]
		arg1 := args[1]
		for _, v := range additionalVars {
			if string(arg0) == v.name {
				arg1 = []byte(v.value)
				break
			}
		}
		i.setScopedVar(fileName, [][]byte{arg0, []byte("="), arg1})
	}
	return
}

func parseFloats(args [][]byte) (floats []float64, err error) {
	floats = make([]float64, len(args))
	for i := range args {
		floats[i], err = strconv.ParseFloat(string(bytes.TrimSpace(args[i])), 64)
		if err != nil {
			err = fmt.Errorf("%v: args=%s", err, bytes.Join(args, []byte(", ")))
			return
		}
	}
	return
}

func encodeHashToHex(h hash.Hash, file string) (sum []byte, err error) {
	b, err := os.ReadFile(file)
	if err != nil {
		return
	}

	h.Write(b)
	s := h.Sum(nil)
	sum = make([]byte, hex.EncodedLen(len(s)))
	hex.Encode(sum, s)
	return
}

// trimQuotes trims spaces, prefix and suffix separately to allow symbol escaping.
func trimQuotes(val []byte) (ret []byte) {
	sq := []byte("'")
	dq := []byte("\"")
	ret = bytes.TrimSpace(val)
	ret = bytes.TrimPrefix(ret, dq)
	ret = bytes.TrimPrefix(ret, sq)
	ret = bytes.TrimSuffix(ret, dq)
	ret = bytes.TrimSuffix(ret, sq)
	return
}
