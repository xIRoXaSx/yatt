package functions

import (
	"errors"
	"fmt"
	"math"
)

func Add(args [][]byte) (ret []byte, err error) {
	var (
		floats []float64
		sum    float64
	)
	if len(args) < 2 {
		err = errors.New("at least 2 args expected")
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

func Ceil(args [][]byte) (ret []byte, err error) {
	var floats []float64
	if len(args) != 1 {
		err = errors.New("exactly 1 arg expected")
		return
	}

	floats, err = ParseFloats(args)
	if err != nil {
		return
	}
	ret = []byte(fmt.Sprint(math.Ceil(floats[0])))
	return
}

func Div(args [][]byte) (ret []byte, err error) {
	var (
		floats []float64
		sum    float64
	)
	if len(args) < 2 {
		err = errors.New("at least 2 args expected")
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

func Floor(args [][]byte) (ret []byte, err error) {
	var floats []float64
	if len(args) != 1 {
		err = errors.New("exactly 1 arg expected")
		return
	}

	floats, err = ParseFloats(args)
	if err != nil {
		return
	}
	ret = []byte(fmt.Sprint(math.Floor(floats[0])))
	return
}

func Max(args [][]byte) (ret []byte, err error) {
	var floats []float64
	if len(args) < 2 {
		err = errors.New("at least 2 args expected")
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

func Min(args [][]byte) (ret []byte, err error) {
	var floats []float64
	if len(args) < 2 {
		err = errors.New("at least 2 args expected")
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

func Mod(args [][]byte) (ret []byte, err error) {
	var floats []float64
	if len(args) != 2 {
		err = errors.New("exactly 2 args expected")
		return
	}

	floats, err = ParseFloats(args)
	if err != nil {
		return
	}
	ret = []byte(fmt.Sprint(math.Mod(floats[0], floats[1])))
	return
}

func ModMin(args [][]byte) (ret []byte, err error) {
	var floats []float64
	if len(args) != 3 {
		err = errors.New("exactly 3 args expected")
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

func Mult(args [][]byte) (ret []byte, err error) {
	var (
		floats []float64
		sum    float64
	)
	if len(args) < 2 {
		err = errors.New("at least 2 args expected")
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

func Pow(args [][]byte) (ret []byte, err error) {
	var floats []float64
	if len(args) < 2 {
		err = errors.New("at least 2 args expected")
		return
	}

	floats, err = ParseFloats(args)
	if err != nil {
		return
	}
	ret = []byte(fmt.Sprint(math.Pow(floats[0], floats[1])))
	return
}

func Sqrt(args [][]byte) (ret []byte, err error) {
	var floats []float64
	if len(args) != 1 {
		err = errors.New("exactly 1 arg expected")
		return
	}

	floats, err = ParseFloats(args)
	if err != nil {
		return
	}
	ret = []byte(fmt.Sprint(math.Sqrt(floats[0])))
	return
}

func Round(args [][]byte) (ret []byte, err error) {
	var floats []float64
	if len(args) != 1 {
		err = errors.New("exactly 1 arg expected")
		return
	}

	floats, err = ParseFloats(args)
	if err != nil {
		return
	}
	ret = []byte(fmt.Sprint(math.Round(floats[0])))
	return
}

func Sub(args [][]byte) (ret []byte, err error) {
	var (
		floats []float64
		sum    float64
	)
	if len(args) < 2 {
		err = errors.New("at least 2 args expected")
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

func Fixed(args [][]byte) (ret []byte, err error) {
	var floats []float64
	if len(args) != 2 {
		err = errors.New("exactly 2 arg expected")
		return
	}

	floats, err = ParseFloats(args)
	if err != nil {
		return
	}
	decPlace := math.Pow10(int(floats[1]))
	ret = []byte(fmt.Sprint(math.Floor(floats[0]*decPlace) / decPlace))
	return
}
