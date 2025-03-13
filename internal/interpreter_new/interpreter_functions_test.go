package interpreter

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"testing"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	r "github.com/stretchr/testify/require"
	"github.com/xiroxasx/fastplate/internal/common"
)

const floatThreshold = 1e-9

func TestFunctions(t *testing.T) {
	t.Parallel()

	l := log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	i := New(&Options{NoStats: true}, l)

	type test struct {
		funcName string
		fileName string
		args     []string
		expected any
		vars     []common.Variable
		fail     bool
	}
	tests := []test{
		{
			funcName: functionNameMathAdd,
			args:     []string{"2", "3", "4"},
			expected: 9,
		},
		{
			funcName: functionNameMathAdd,
			args:     []string{"  2  ", "  3  ", "  -4  "},
			expected: 1,
		},
		{
			funcName: functionNameMathSub,
			args:     []string{"2", "3", "4"},
			expected: -5,
		},
		{
			funcName: functionNameMathSub,
			args:     []string{"  2  ", "  3  ", "  -4  "},
			expected: 3,
		},
		{
			funcName: functionNameMathMult,
			args:     []string{"2", "3", "4"},
			expected: 24,
		},
		// 5.
		{
			funcName: functionNameMathMult,
			args:     []string{"  2  ", "  3  ", "  -4  "},
			expected: -24,
		},
		{
			funcName: functionNameMathDiv,
			args:     []string{"2", "3", "4"},
			expected: 0.166666667,
		},
		{
			funcName: functionNameMathDiv,
			args:     []string{"  2  ", "  3  ", "  -4  "},
			expected: -0.166666667,
		},
		{
			funcName: functionNameMathPow,
			args:     []string{"3", "4"},
			expected: 81,
		},
		{
			funcName: functionNameMathPow,
			args:     []string{"  3  ", "  -4  "},
			expected: 0.012345679,
		},
		// 10.
		{
			funcName: functionNameMathSqrt,
			args:     []string{"3"},
			expected: 1.73205080757,
		},
		{
			funcName: functionNameMathSqrt,
			args:     []string{"  -3  "},
			expected: math.NaN(),
		},
		{
			funcName: functionNameMathRound,
			args:     []string{"1.73205080757"},
			expected: 2,
		},
		{
			funcName: functionNameMathRound,
			args:     []string{"  -1.73205080757  "},
			expected: -2,
		},
		{
			funcName: functionNameMathCeil,
			args:     []string{"1.23205080757"},
			expected: 2,
		},
		// 15.
		{
			funcName: functionNameMathCeil,
			args:     []string{"  -1.23205080757  "},
			expected: -1,
		},
		{
			funcName: functionNameMathFloor,
			args:     []string{"1.73205080757"},
			expected: 1,
		},
		{
			funcName: functionNameMathFloor,
			args:     []string{"  -1.73205080757  "},
			expected: -2,
		},
		{
			funcName: functionNameMathFixed,
			args:     []string{"1.73205080757", "  5  "},
			expected: 1.73205,
		},
		{
			funcName: functionNameMathFixed,
			args:     []string{"  -1.73205080757  ", "  5  "},
			expected: -1.73205,
		},
		// 20.
		{
			funcName: functionNameMathMax,
			args:     []string{"2", "3", "4"},
			expected: 4,
		},
		{
			funcName: functionNameMathMax,
			args:     []string{"  2  ", "  3  ", "  -4  "},
			expected: 3,
		},
		{
			funcName: functionNameMathMin,
			args:     []string{"2", "3", "4"},
			expected: 2,
		},
		{
			funcName: functionNameMathMin,
			args:     []string{"  2  ", "  3  ", "  -4  "},
			expected: -4,
		},
		{
			funcName: functionNameMathMod,
			args:     []string{"2", "3"},
			expected: 2,
		},
		// 25.
		{
			funcName: functionNameMathMod,
			args:     []string{"  2  ", "  -4  "},
			expected: 2,
		},
		{
			funcName: functionNameMathAdd,
			args:     []string{},
			fail:     true,
		},
		{
			funcName: "n/a",
			fail:     true,
		},
	}

	for j, test := range tests {
		args := make([][]byte, len(test.args))
		for k, arg := range test.args {
			args[k] = []byte(arg)
		}

		ret, err := i.state.executeFunction(interpreterFunc(test.funcName), test.fileName, args, test.vars)
		if test.fail {
			r.Error(t, err)
			continue
		}

		expected, ok := test.expected.(float64)
		if ok {
			retFloat, err := strconv.ParseFloat(string(ret), 64)
			r.NoError(t, err, "case=%d, expected=%v, actual=%v", j, expected, retFloat)
			r.True(t, floatCompareOK(expected, retFloat), "case=%d, expected=%v, actual=%v", j, expected, retFloat)
			continue
		}
		r.NoError(t, err, "case=%d", j)
		r.Exactly(t, fmt.Sprintf("%v", test.expected), string(ret), "case=%d", j)
	}
}

func floatCompareOK(expected, actual float64) bool {
	return (math.IsNaN(expected) && math.IsNaN(actual)) ||
		math.Abs(expected-actual) <= floatThreshold
}
