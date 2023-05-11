package functions

import (
	"fmt"
	"math"
)

func Add(fn string, args [][]byte) (ret []byte, err error) {
	var (
		floats []float64
		sum    float64
	)
	if len(args) < 2 {
		err = fmt.Errorf("%s: at least 2 args expected", fn)
		return
	}

	floats, err = ParseFloats(args)
	if err != nil {
		return
	}
	for j := range floats {
		sum += floats[j]
	}
	ret = []byte(fmt.Sprint(sum))
	return
}

func Ceil(fn string, args [][]byte) (ret []byte, err error) {
	var floats []float64
	if len(args) != 1 {
		err = fmt.Errorf("%s: exactly 1 arg expected", fn)
		return
	}

	floats, err = ParseFloats(args)
	if err != nil {
		return
	}
	ret = []byte(fmt.Sprint(math.Ceil(floats[0])))
	return
}

func Div(fn string, args [][]byte) (ret []byte, err error) {
	var (
		floats []float64
		sum    float64
	)
	if len(args) < 2 {
		err = fmt.Errorf("%s: at least 2 args expected", fn)
		return
	}

	floats, err = ParseFloats(args)
	if err != nil {
		return
	}
	sum = floats[0]
	for _, f := range floats[1:] {
		sum /= f
	}
	ret = []byte(fmt.Sprint(sum))
	return
}

func Floor(fn string, args [][]byte) (ret []byte, err error) {
	var floats []float64
	if len(args) != 1 {
		err = fmt.Errorf("%s: exactly 1 arg expected", fn)
		return
	}

	floats, err = ParseFloats(args)
	if err != nil {
		return
	}
	ret = []byte(fmt.Sprint(math.Floor(floats[0])))
	return
}

func Max(fn string, args [][]byte) (ret []byte, err error) {
	var floats []float64
	if len(args) < 2 {
		err = fmt.Errorf("%s: at least 2 args expected", fn)
		return
	}

	floats, err = ParseFloats(args)
	if err != nil {
		return
	}
	max := floats[0]
	for _, f := range floats[1:] {
		max = math.Max(max, f)
	}
	ret = []byte(fmt.Sprint(max))
	return
}

func Min(fn string, args [][]byte) (ret []byte, err error) {
	var floats []float64
	if len(args) < 2 {
		err = fmt.Errorf("%s: at least 2 args expected", fn)
		return
	}

	floats, err = ParseFloats(args)
	if err != nil {
		return
	}
	min := floats[0]
	for _, f := range floats[1:] {
		min = math.Min(min, f)
	}
	ret = []byte(fmt.Sprint(min))
	return
}

func Mod(fn string, args [][]byte) (ret []byte, err error) {
	var floats []float64
	if len(args) != 2 {
		err = fmt.Errorf("%s: exactly 2 args expected", fn)
		return
	}

	floats, err = ParseFloats(args)
	if err != nil {
		return
	}
	ret = []byte(fmt.Sprint(math.Mod(floats[0], floats[1])))
	return
}

func ModMin(fn string, args [][]byte) (ret []byte, err error) {
	var floats []float64
	if len(args) != 3 {
		err = fmt.Errorf("%s: exactly 3 args expected", fn)
		return
	}

	floats, err = ParseFloats(args)
	if err != nil {
		return
	}
	mod := math.Mod(floats[0], floats[1])
	ret = []byte(fmt.Sprint(math.Max(floats[2], mod)))
	return
}

func Mult(fn string, args [][]byte) (ret []byte, err error) {
	var (
		floats []float64
		sum    float64
	)
	if len(args) < 2 {
		err = fmt.Errorf("%s: at least 2 args expected", fn)
		return
	}

	floats, err = ParseFloats(args)
	if err != nil {
		return
	}
	sum = floats[0]
	for _, f := range floats[1:] {
		sum *= f
	}
	ret = []byte(fmt.Sprint(sum))
	return
}

func Round(fn string, args [][]byte) (ret []byte, err error) {
	var floats []float64
	if len(args) != 1 {
		err = fmt.Errorf("%s: exactly 1 arg expected", fn)
		return
	}

	floats, err = ParseFloats(args)
	if err != nil {
		return
	}
	ret = []byte(fmt.Sprint(math.Round(floats[0])))
	return
}

func Sub(fn string, args [][]byte) (ret []byte, err error) {
	var (
		floats []float64
		sum    float64
	)
	if len(args) < 2 {
		err = fmt.Errorf("%s: at least 2 args expected", fn)
		return
	}

	floats, err = ParseFloats(args)
	if err != nil {
		return
	}
	sum = floats[0]
	for _, f := range floats[1:] {
		sum -= f
	}
	ret = []byte(fmt.Sprint(sum))
	return
}

func Fixed(fn string, args [][]byte) (ret []byte, err error) {
	var floats []float64
	if len(args) != 2 {
		err = fmt.Errorf("%s: exactly 2 arg expected", fn)
		return
	}

	floats, err = ParseFloats(args)
	if err != nil {
		return
	}
	decPlace := floats[1] * 10
	ret = []byte(fmt.Sprint(math.Round(floats[0]*decPlace) / decPlace))
	return
}
