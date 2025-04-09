package core

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

func TestDependenciesAreCyclic(t *testing.T) {
	t.Parallel()

	type importCase struct {
		src string
		imp string
	}

	type testCaseWrapper struct {
		fail  bool
		cases []importCase
	}

	testCases := []testCaseWrapper{
		{
			fail: true,
			cases: []importCase{
				{src: "fileA", imp: "fileB"},
				{src: "fileB", imp: "fileA"},
			},
		},
		{
			fail: true,
			cases: []importCase{
				{src: "fileA", imp: "fileB"},
				{src: "fileB", imp: "fileC"},
				{src: "fileC", imp: "fileA"},
			},
		},
		{
			fail: true,
			cases: []importCase{
				{src: "fileA", imp: "fileB"},
				{src: "fileB", imp: "fileC"},
				{src: "fileC", imp: "fileD"},
				{src: "fileC", imp: "fileE"},
				{src: "fileE", imp: "fileA"},
			},
		},
		{
			fail: true,
			cases: []importCase{
				{src: "fileA", imp: "fileB"},
				{src: "fileB", imp: "fileC"},
				{src: "fileC", imp: "fileD"},
				{src: "fileB", imp: "fileF"},
				{src: "fileF", imp: "fileH"},
				{src: "fileB", imp: "fileG"},
				{src: "fileD", imp: "fileB"},
			},
		},
		{
			fail: true,
			cases: []importCase{
				{src: "fileA", imp: "fileB"},
				{src: "fileB", imp: "fileC"},
				{src: "fileC", imp: "fileD"},
				{src: "fileB", imp: "fileF"},
				{src: "fileF", imp: "fileH"},
				{src: "fileB", imp: "fileG"},
				{src: "fileD", imp: "fileC"},
			},
		},
		{
			fail: true,
			cases: []importCase{
				{src: "fileA", imp: "fileB"},
				{src: "fileB", imp: "fileC"},
				{src: "fileC", imp: "fileD"},
				{src: "fileD", imp: "fileE"},
				{src: "fileE", imp: "fileC"},
				{src: "fileB", imp: "fileF"},
				{src: "fileF", imp: "fileH"},
				{src: "fileB", imp: "fileG"},
			},
		},
		{
			cases: []importCase{
				{src: "fileA", imp: "fileB"},
				{src: "fileB", imp: "fileC"},
				{src: "fileC", imp: "fileD"},
				{src: "fileC", imp: "fileF"},
				{src: "fileC", imp: "fileG"},
				{src: "fileF", imp: "fileH"},
				{src: "fileF", imp: "fileI"},
				{src: "fileG", imp: "fileJ"},
				{src: "fileY", imp: "fileZ"},
			},
		},
	}

	for i, tcw := range testCases {
		dr := newDependencyResolver()

		// Add all dependencies accordingly.
		for _, tc := range tcw.cases {
			dr.addDependency(tc.src, tc.imp)
		}

		startCase := tcw.cases[0]
		ok := dr.CheckForCyclicDependencies(startCase.src, startCase.imp)
		r.Equal(t, tcw.fail, ok, "case=%d", i)
	}
}

func TestImportPaths(t *testing.T) {
	t.Parallel()

	l := log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	c := New(l, []string{"# fastplate"}, Options{})
	err := c.ImportPathCheckCyclicDependencies("testdata/imports/fileA.txt")
	r.ErrorIs(t, err, errDependencyCyclic)
}

func TestSetLocalVarByArg(t *testing.T) {
	t.Parallel()

	const (
		keyRegisterName = "testName"
		keyName         = "name"
		keyValue        = "value"
		keyNameNew      = "newName"
		keyValueNew     = "newValue"
	)

	l := zerolog.New(os.Stderr)
	c := New(l, nil, Options{})
	err := c.setLocalVarByArg(keyRegisterName, []byte(keyName+"="+keyValue))
	r.NoError(t, err)
	localVars := c.varRegistryLocal.entries[keyRegisterName]
	r.Len(t, localVars, 1)
	v := localVars[0]
	r.Exactly(t, keyName, v.Name())
	r.Exactly(t, keyValue, v.Value())

	// Variables with the equal names should be updated.
	err = c.setLocalVarByArg(keyRegisterName, []byte(keyName+"="+keyValueNew))
	r.NoError(t, err)
	localVars = c.varRegistryLocal.entries[keyRegisterName]
	r.Len(t, localVars, 1)
	v = localVars[0]
	r.Exactly(t, keyName, v.Name())
	r.Exactly(t, keyValueNew, v.Value())

	// New variables with the same value should be added.
	err = c.setLocalVarByArg(keyRegisterName, []byte(keyNameNew+"="+keyValueNew))
	r.NoError(t, err)
	localVars = c.varRegistryLocal.entries[keyRegisterName]
	r.Len(t, localVars, 2)
	v = localVars[1]
	r.Exactly(t, keyNameNew, v.Name())
	r.Exactly(t, keyValueNew, v.Value())

	// Empty variable names and values should return an error.
	err = c.setLocalVarByArg(keyRegisterName, []byte("="+keyValueNew))
	r.ErrorIs(t, err, errEmptyVariableParameter)
	err = c.setLocalVarByArg(keyRegisterName, []byte(keyName+"="))
	r.ErrorIs(t, err, errEmptyVariableParameter)
}

func TestVariableSetAndGet(t *testing.T) {
	t.Parallel()

	type testCase struct {
		content       []byte
		expectedName  string
		expectedValue string
	}

	const (
		keyName  = "name"
		keyValue = "value"
	)

	l := log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	c := New(l, nil, Options{})

	testCases := []testCase{
		{
			expectedName:  keyName,
			expectedValue: keyValue,
			content:       []byte(keyName + "=" + keyValue),
		},
		{
			expectedName:  keyName,
			expectedValue: keyValue,
			content:       []byte(keyName + " = " + keyValue),
		},
		{
			expectedName:  keyName,
			expectedValue: keyValue,
			content:       []byte(keyName + "  =  " + keyValue),
		},
		{
			expectedName:  keyName,
			expectedValue: keyValue,
			content:       []byte(keyName + " =  " + keyValue),
		},
		{
			expectedName:  keyName,
			expectedValue: keyValue,
			content:       []byte(keyName + "  = " + keyValue),
		},
	}

	for i, tc := range testCases {
		v := common.VarFromArg(tc.content)

		// Global variables.
		// Set a new variable.
		c.setGlobalVar(v)
		rvar := c.varLookupGlobal(v.Name())
		r.Exactly(t, v.Name(), rvar.Name(), "global var name differs: case=%d, expected=%s, actual=%s", i, v.Name(), rvar.Name())
		r.Exactly(t, v.Value(), rvar.Value(), "global var value differs: case=%d, expected=%s, actual=%s", i, v.Value(), rvar.Value())
		// Unset the variable's value.
		c.setGlobalVar(variable{name: v.Name()})
		rvar = c.varLookupGlobal(v.Name())
		r.Exactly(t, v.Name(), rvar.Name(), "cleared global var name differs: case=%d, expected=%s, actual=%s", i, v.Name(), rvar.Name())
		r.Exactly(t, "", rvar.Value(), "cleared global var value differs: case=%d, expected=%s, actual=%s", i, v.Value(), rvar.Value())

		// Local variables.
		// Set a new variable.
		register := "test123"
		c.setLocalVar(register, v)
		rvar = c.varLookupLocal(register, v.Name())
		r.Exactly(t, v.Name(), rvar.Name(), "local var name differs: case=%d, expected=%s, actual=%s", i, v.Name(), rvar.Name())
		r.Exactly(t, v.Value(), rvar.Value(), "local var value differs: case=%d, expected=%s, actual=%s", i, v.Value(), rvar.Value())
		// Unset the variable's value.
		c.setLocalVar(register, variable{name: v.Name()})
		rvar = c.varLookupLocal(register, v.Name())
		r.Exactly(t, v.Name(), rvar.Name(), "cleared local var name differs: case=%d, expected=%s, actual=%s", i, v.Name(), rvar.Name())
		r.Exactly(t, "", rvar.Value(), "cleared local var value differs: case=%d, expected=%s, actual=%s", i, v.Value(), rvar.Value())
	}
}

func TestExecuteFunctions(t *testing.T) {
	t.Parallel()

	const localTestVarFileName = "test-filename.txt"
	l := log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	c := New(l, nil, Options{})
	c.setGlobalVar(common.NewVar("test123", "321test"))
	c.setLocalVar(localTestVarFileName, common.NewVar("test123", "321test"))
	c.setLocalVar(localTestVarFileName, common.NewVar("123", "321"))

	const testFileName = "testdata/test.txt"

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
			funcName: functionNameCryptSHA1,
			args:     []string{testFileName},
			expected: "a94a8fe5ccb19ba61c4c0873d391e987982fbbd3",
		},
		{
			funcName: functionNameCryptSHA256,
			args:     []string{testFileName},
			expected: "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08",
		},
		{
			funcName: functionNameCryptSHA512,
			args:     []string{testFileName},
			expected: "ee26b0dd4af7e749aa1a8ee3c10ae9923f618980772e473f8819a5d4940e0db27ac185f8a0e1d5f84f88bc887fd67b143732c304cc5fa9ad8e6f57f50028a8ff",
		},
		// 30.
		{
			funcName: functionNameCryptMD5,
			args:     []string{testFileName},
			expected: "098f6bcd4621d373cade4e832627b4f6",
		},
		{
			funcName: functionNameStringSplit,
			args:     []string{"test|123", "|", "1"},
			expected: "123",
		},
		{
			funcName: functionNameStringRepeat,
			args:     []string{"test,", "3"},
			expected: "test,test,test,",
		},
		{
			funcName: functionNameStringReplace,
			args:     []string{"test test test ", " ", "+"},
			expected: "test+test+test+",
		},
		{
			funcName: functionNameStringLength,
			args:     []string{"test"},
			expected: "4",
		},
		// 35.
		{
			funcName: functionNameInternalVar,
			args:     []string{"test", "value"},
			expected: "",
		},
		{
			funcName: functionNameInternalVar,
			args:     []string{"test", "value2"},
			expected: "",
		},
		{
			funcName: functionNameStringLength,
			fileName: localTestVarFileName,
			args:     []string{"FASTPLATE_VARS"},
			expected: "1",
		},
		{
			funcName: functionNameStringLength,
			fileName: localTestVarFileName,
			args:     []string{"FASTPLATE_VARS_" + localTestVarFileName},
			expected: "2",
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

		ret, err := c.executeFunction(parserFunc(test.funcName), test.fileName, args, test.vars)
		if test.fail {
			r.Error(t, err, "function=%s", test.funcName)
			continue
		}

		expected, ok := test.expected.(float64)
		if ok {
			retFloat, err := strconv.ParseFloat(string(ret), 64)
			r.NoError(t, err, "case=%d, function=%s, expected=%v, actual=%v", j, test.funcName, expected, retFloat)
			r.True(t, floatCompareOK(expected, retFloat), "case=%d, function=%s, expected=%v, actual=%v", j, test.funcName, expected, retFloat)
			continue
		}
		r.NoError(t, err, "case=%d, function=%s", j, test.funcName)
		r.Exactly(t, fmt.Sprintf("%v", test.expected), string(ret), "case=%d, function=%s", j, test.funcName)
	}
}

func TestInterpreterResolveNested(t *testing.T) {
	t.Parallel()

	l := log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	c := New(l, nil, Options{})
	ret, err := c.resolve(resolveArgs{
		fileName: "test.txt",
		line:     []byte("test 123 {{add(1,2,{{mult(2,3)}})}}"),
	})
	r.NoError(t, err)
	r.Exactly(t, string(ret), "test 123 9")
}

//
// Helper
//

func floatCompareOK(expected, actual float64) bool {
	return (math.IsNaN(expected) && math.IsNaN(actual)) ||
		math.Abs(expected-actual) <= floatThreshold
}
