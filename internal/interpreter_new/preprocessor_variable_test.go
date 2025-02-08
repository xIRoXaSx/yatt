package interpreter

import (
	"fmt"
	"os"
	"testing"

	"github.com/rs/zerolog"
	r "github.com/stretchr/testify/require"
)

func TestVariableFromArgs(t *testing.T) {
	t.Parallel()

	type testCase struct {
		content       [][]byte
		expectedName  string
		expectedValue string
	}

	const (
		keyName  = "name"
		keyValue = "value"
	)

	keyValuePair := func(k, v string) []byte {
		return []byte(fmt.Sprintf("%s=%s", k, v))
	}

	testCases := []testCase{
		{content: nil},
		{content: [][]byte{[]byte("")}},
		{content: [][]byte{[]byte("=")}},
		{
			expectedValue: keyValue,
			content:       [][]byte{keyValuePair("", keyValue)},
		},
		{
			expectedValue: "=",
			content:       [][]byte{keyValuePair("", "=")},
		},
		{
			expectedName:  keyName,
			expectedValue: keyValue,
			content:       [][]byte{keyValuePair(keyName, keyValue)},
		},
		{
			expectedName:  keyName,
			expectedValue: keyValue,
			content: [][]byte{
				[]byte(keyName),
				[]byte(""),
				[]byte("="),
				[]byte(""),
				[]byte(keyValue),
			},
		},
		{
			expectedName:  keyName,
			expectedValue: keyValue,
			content: [][]byte{
				[]byte(keyName),
				[]byte(""),
				[]byte(""),
				[]byte(""),
				[]byte("="),
				[]byte(keyValue),
			},
		},
		{
			expectedName:  keyName,
			expectedValue: keyValue,
			content: [][]byte{
				[]byte(keyName),
				[]byte(""),
				[]byte(""),
				[]byte("="),
				[]byte(""),
				[]byte(""),
				[]byte(keyValue),
			},
		},
		{
			expectedName:  keyName,
			expectedValue: keyValue,
			content: [][]byte{
				[]byte(keyName),
				[]byte(""),
				[]byte("="),
				[]byte(""),
				[]byte(""),
				[]byte(""),
				[]byte(keyValue),
			},
		},
		{
			expectedName:  keyName,
			expectedValue: keyValue,
			content: [][]byte{
				[]byte(keyName + "="),
				[]byte(""),
				[]byte(""),
				[]byte(keyValue),
			},
		},
		{
			expectedName:  keyName,
			expectedValue: keyValue,
			content: [][]byte{
				[]byte(keyName),
				[]byte(""),
				[]byte(""),
				[]byte(""),
				[]byte("=" + keyValue),
			},
		},
	}

	for i, tc := range testCases {
		v := variableFromArgs(tc.content)
		r.Exactly(t, tc.expectedName, v.name, "names do not match: case=%d, expected=%s, actual=%s", i, tc.expectedName, v.name)
		r.Exactly(t, tc.expectedValue, v.value, "names do not match: case=%d, expected=%s, actual=%s", i, tc.expectedValue, v.value)
	}
}

func TestSetLocalVar(t *testing.T) {
	t.Parallel()

	const (
		keyRegisterName = "testName"
		keyName         = "name"
		keyValue        = "value"
		keyNameNew      = "newName"
		keyValueNew     = "newValue"
	)

	i := New(&Options{NoStats: true}, zerolog.New(os.Stderr))
	err := i.setLocalVar(keyRegisterName, [][]byte{[]byte(keyName + "=" + keyValue)})
	r.NoError(t, err)
	localVars := i.state.varRegistryLocal.entries[keyRegisterName]
	r.Len(t, localVars, 1)
	v := localVars[0]
	r.Exactly(t, keyName, v.Name())
	r.Exactly(t, keyValue, v.Value())

	// Variables with the equal names should be updated.
	err = i.setLocalVar(keyRegisterName, [][]byte{[]byte(keyName + "=" + keyValueNew)})
	r.NoError(t, err)
	localVars = i.state.varRegistryLocal.entries[keyRegisterName]
	r.Len(t, localVars, 1)
	v = localVars[0]
	r.Exactly(t, keyName, v.Name())
	r.Exactly(t, keyValueNew, v.Value())

	// New variables with the same value should be added.
	err = i.setLocalVar(keyRegisterName, [][]byte{[]byte(keyNameNew + "=" + keyValueNew)})
	r.NoError(t, err)
	localVars = i.state.varRegistryLocal.entries[keyRegisterName]
	r.Len(t, localVars, 2)
	v = localVars[1]
	r.Exactly(t, keyNameNew, v.Name())
	r.Exactly(t, keyValueNew, v.Value())

	// Empty variable names and values should return an error.
	err = i.setLocalVar(keyRegisterName, [][]byte{[]byte("=" + keyValueNew)})
	r.ErrorIs(t, err, errEmptyVariableParameter)
	err = i.setLocalVar(keyRegisterName, [][]byte{[]byte(keyName + "=")})
	r.ErrorIs(t, err, errEmptyVariableParameter)
}
