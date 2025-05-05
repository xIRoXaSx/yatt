package functions

import (
	"math"
)

func Add(args [][]byte) (ret []byte, err error) {
	err = assertArgsLengthAtLeast(args, 2)
	if err != nil {
		return
	}

	floats, err := parseFloats(args)
	if err != nil {
		return
	}

	sum := floats[0]
	for _, f := range floats[1:] {
		sum += f
	}
	ret = floatToBytes(sum)
	return
}

func Sub(args [][]byte) (ret []byte, err error) {
	err = assertArgsLengthAtLeast(args, 2)
	if err != nil {
		return
	}

	floats, err := parseFloats(args)
	if err != nil {
		return
	}

	sum := floats[0]
	for _, f := range floats[1:] {
		sum -= f
	}
	ret = floatToBytes(sum)
	return
}

func Mult(args [][]byte) (ret []byte, err error) {
	err = assertArgsLengthAtLeast(args, 2)
	if err != nil {
		return
	}

	floats, err := parseFloats(args)
	if err != nil {
		return
	}

	sum := floats[0]
	for _, f := range floats[1:] {
		sum *= f
	}
	ret = floatToBytes(sum)
	return
}

func Div(args [][]byte) (ret []byte, err error) {
	err = assertArgsLengthAtLeast(args, 2)
	if err != nil {
		return
	}

	floats, err := parseFloats(args)
	if err != nil {
		return
	}

	sum := floats[0]
	for _, f := range floats[1:] {
		sum /= f
	}
	ret = floatToBytes(sum)
	return
}

func Pow(args [][]byte) (ret []byte, err error) {
	err = assertArgsLengthExact(args, 2)
	if err != nil {
		return
	}

	floats, err := parseFloats(args)
	if err != nil {
		return
	}

	ret = floatToBytes(math.Pow(floats[0], floats[1]))
	return
}

func Sqrt(args [][]byte) (ret []byte, err error) {
	err = assertArgsLengthExact(args, 1)
	if err != nil {
		return
	}

	floats, err := parseFloats(args)
	if err != nil {
		return
	}

	ret = floatToBytes(math.Sqrt(floats[0]))
	return
}

func Round(args [][]byte) (ret []byte, err error) {
	err = assertArgsLengthExact(args, 1)
	if err != nil {
		return
	}

	floats, err := parseFloats(args)
	if err != nil {
		return
	}

	ret = floatToBytes(math.Round(floats[0]))
	return
}

func Ceil(args [][]byte) (ret []byte, err error) {
	err = assertArgsLengthAtLeast(args, 1)
	if err != nil {
		return
	}

	floats, err := parseFloats(args)
	if err != nil {
		return
	}

	ret = floatToBytes(math.Ceil(floats[0]))
	return
}

func Floor(args [][]byte) (ret []byte, err error) {
	err = assertArgsLengthAtLeast(args, 1)
	if err != nil {
		return
	}

	floats, err := parseFloats(args)
	if err != nil {
		return
	}

	ret = floatToBytes(math.Floor(floats[0]))
	return
}

func Fixed(args [][]byte) (ret []byte, err error) {
	err = assertArgsLengthExact(args, 2)
	if err != nil {
		return
	}

	floats, err := parseFloats(args)
	if err != nil {
		return
	}

	decPlace := math.Pow10(int(floats[1]))
	value := floats[0]
	if value < 0 {
		ret = floatToBytes(math.Ceil(value*decPlace) / decPlace)
		return
	}

	ret = floatToBytes(math.Floor(value*decPlace) / decPlace)
	return
}

func Max(args [][]byte) (ret []byte, err error) {
	err = assertArgsLengthAtLeast(args, 2)
	if err != nil {
		return
	}

	floats, err := parseFloats(args)
	if err != nil {
		return
	}

	max := floats[0]
	for _, f := range floats[1:] {
		max = math.Max(max, f)
	}
	ret = floatToBytes(max)
	return
}

func Min(args [][]byte) (ret []byte, err error) {
	err = assertArgsLengthAtLeast(args, 2)
	if err != nil {
		return
	}

	floats, err := parseFloats(args)
	if err != nil {
		return
	}

	min := floats[0]
	for _, f := range floats[1:] {
		min = math.Min(min, f)
	}
	ret = floatToBytes(min)
	return
}

func Mod(args [][]byte) (ret []byte, err error) {
	err = assertArgsLengthAtLeast(args, 2)
	if err != nil {
		return
	}

	floats, err := parseFloats(args)
	if err != nil {
		return
	}

	ret = floatToBytes(math.Mod(floats[0], floats[1]))
	return
}
