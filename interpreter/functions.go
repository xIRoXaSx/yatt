package importer

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
	functionModulus    = "mod"
	functionSha1       = "sha1"
	functionSha256     = "sha256"
	functionSha512     = "sha512"
	functionShaMd5     = "md5"
)

func (i *Interpreter) executeFunction(function string, args [][]byte) (ret []byte, err error) {
	if len(args) == 0 {
		err = fmt.Errorf("%v: func statement needs at least one char", function)
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

	case functionModulus:
		var floats []float64
		if len(args) != 2 {
			err = fmt.Errorf("%s: exactly 1 arg expected", function)
			return
		}
		floats, err = parseFloats(args)
		if err != nil {
			return
		}
		ret = []byte(fmt.Sprint(math.Mod(floats[0], floats[1])))

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
	}
	return
}

func parseFloats(args [][]byte) (floats []float64, err error) {
	floats = make([]float64, len(args))
	for i := range args {
		floats[i], err = strconv.ParseFloat(string(bytes.TrimSpace(args[i])), 64)
		if err != nil {
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
