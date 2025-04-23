package core

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"math"
	"os"
	"path/filepath"
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
	err := c.ImportPathCheckCyclicDependencies("testdata/deps/fileA.txt")
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

	const testFileName = "testdata/functions/test.txt"

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
			expected: "62b8c0e420a152e68182cd3fa32947dc628eccc5",
		},
		{
			funcName: functionNameCryptSHA256,
			args:     []string{testFileName},
			expected: "0c6fb147ed8471d30fe62bdefce2becc6a589fc16442738e313165abba527e99",
		},
		{
			funcName: functionNameCryptSHA512,
			args:     []string{testFileName},
			expected: "08f92f25ebe532d24c57c56a8467714e2b3ab2958f622a93e47ce073116e78f75ec3073c5ff8737a9223019c901754ffc96fed60b71e7af163fef28be6a98266",
		},
		// 30.
		{
			funcName: functionNameCryptMD5,
			args:     []string{testFileName},
			expected: "a8b518cddc851290ab1e1bb6b0b41072",
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

func TestInterpreterImport(t *testing.T) {
	t.Parallel()

	l := log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	rootTestDir := filepath.Join("testdata", "imports")
	in := filepath.Join(rootTestDir, "in", "startOk.txt")
	c := New(l, []string{"# fastplate"}, Options{PreserveIndent: true})

	rc, err := os.Open(in)
	r.NoError(t, err)
	defer rc.Close()

	buf := &bytes.Buffer{}
	err = c.Interpret(InterpreterFile{
		Name: in,
		Buf:  buf,
		RC:   rc,
	})
	r.NoError(t, err)

	h := sha1.New()
	_, err = h.Write(buf.Bytes())
	r.NoError(t, err)
	s := h.Sum(nil)
	sum := make([]byte, hex.EncodedLen(len(s)))
	_ = hex.Encode(sum, s)

	r.True(t, bytes.Equal(sum, []byte("c5f70bc07c8a941f1091727896d3a7e495725abf")))
}

func TestInterpreterImportCycle(t *testing.T) {
	t.Parallel()

	prefixes := []string{"# fastplate"}

	l := log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	c := New(l, prefixes, Options{})

	rootTestDir := filepath.Join("testdata", "imports", "in")
	err := c.ImportPathCheckCyclicDependencies(filepath.Join(rootTestDir, "startOk.txt"))
	r.NoError(t, err)

	c = New(l, prefixes, Options{})
	err = c.ImportPathCheckCyclicDependencies(filepath.Join(rootTestDir, "startFail.txt"))
	r.Error(t, err)
}

func TestIgnore(t *testing.T) {
	t.Parallel()

	l := log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	inPath := filepath.Join("testdata", "ignore", "in", "ignore.yaml")
	inFile, err := os.Open(inPath)
	r.NoError(t, err)
	defer func() {
		_ = inFile.Close()
	}()

	c := New(l, []string{"# fastplate"}, Options{PreserveIndent: true})
	buf := &bytes.Buffer{}
	err = c.Interpret(InterpreterFile{
		Name: inPath,
		Buf:  buf,
		RC:   inFile,
	})
	r.NoError(t, err)

	h := sha1.New()
	_, err = h.Write(buf.Bytes())
	r.NoError(t, err)
	s := h.Sum(nil)
	sum := make([]byte, hex.EncodedLen(len(s)))
	_ = hex.Encode(sum, s)

	r.True(t, bytes.Equal(sum, []byte("9b12b4004fdca14cd81c9378b6dc040feb39e730")))
}

//
// Helper
//

func floatCompareOK(expected, actual float64) bool {
	return (math.IsNaN(expected) && math.IsNaN(actual)) ||
		math.Abs(expected-actual) <= floatThreshold
}
