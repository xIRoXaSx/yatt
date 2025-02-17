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
			funcName: functionNameAdd,
			args:     []string{"2", "3", "4"},
			expected: 9,
		},
		{
			funcName: functionNameAdd,
			args:     []string{"  2  ", "  3  ", "  -4  "},
			expected: 1,
		},
		{
			funcName: functionNameSub,
			args:     []string{"2", "3", "4"},
			expected: -5,
		},
		{
			funcName: functionNameSub,
			args:     []string{"  2  ", "  3  ", "  -4  "},
			expected: 3,
		},
		{
			funcName: functionNameMult,
			args:     []string{"2", "3", "4"},
			expected: 24,
		},
		// 5.
		{
			funcName: functionNameMult,
			args:     []string{"  2  ", "  3  ", "  -4  "},
			expected: -24,
		},
		{
			funcName: functionNameDiv,
			args:     []string{"2", "3", "4"},
			expected: 0.166666667,
		},
		{
			funcName: functionNameDiv,
			args:     []string{"  2  ", "  3  ", "  -4  "},
			expected: -0.166666667,
		},
		{
			funcName: functionNamePow,
			args:     []string{"3", "4"},
			expected: 81,
		},
		{
			funcName: functionNamePow,
			args:     []string{"  3  ", "  -4  "},
			expected: 0.012345679,
		},
		// 10.
		{
			funcName: functionNameSqrt,
			args:     []string{"3"},
			expected: 1.73205080757,
		},
		{
			funcName: functionNameSqrt,
			args:     []string{"  -3  "},
			expected: math.NaN(),
		},
		{
			funcName: functionNameRound,
			args:     []string{"1.73205080757"},
			expected: 2,
		},
		{
			funcName: functionNameRound,
			args:     []string{"  -1.73205080757  "},
			expected: -2,
		},
		{
			funcName: functionNameCeil,
			args:     []string{"1.23205080757"},
			expected: 2,
		},
		// 15.
		{
			funcName: functionNameCeil,
			args:     []string{"  -1.23205080757  "},
			expected: -1,
		},
		{
			funcName: functionNameFloor,
			args:     []string{"1.73205080757"},
			expected: 1,
		},
		{
			funcName: functionNameFloor,
			args:     []string{"  -1.73205080757  "},
			expected: -2,
		},
		{
			funcName: functionNameFixed,
			args:     []string{"1.73205080757", "  5  "},
			expected: 1.73205,
		},
		{
			funcName: functionNameFixed,
			args:     []string{"  -1.73205080757  ", "  5  "},
			expected: -1.73205,
		},
		// 20.
		{
			funcName: functionNameMax,
			args:     []string{"2", "3", "4"},
			expected: 4,
		},
		{
			funcName: functionNameMax,
			args:     []string{"  2  ", "  3  ", "  -4  "},
			expected: 3,
		},
		{
			funcName: functionNameMin,
			args:     []string{"2", "3", "4"},
			expected: 2,
		},
		{
			funcName: functionNameMin,
			args:     []string{"  2  ", "  3  ", "  -4  "},
			expected: -4,
		},
		{
			funcName: functionNameMod,
			args:     []string{"2", "3"},
			expected: 2,
		},
		// 25.
		{
			funcName: functionNameMod,
			args:     []string{"  2  ", "  -4  "},
			expected: 2,
		},
		{
			funcName: functionNameAdd,
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

		ret, err := i.executeFunction(interpreterFunc(test.funcName), test.fileName, args, test.vars)
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
