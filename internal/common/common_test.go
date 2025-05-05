package common

import (
	"fmt"
	"testing"

	r "github.com/stretchr/testify/require"
)

func TestVariableFromArg(t *testing.T) {
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

	keyValuePair := func(k, v string) []byte {
		return fmt.Appendf(nil, "%s=%s", k, v)
	}

	testCases := []testCase{
		{content: nil},
		{content: []byte{}},
		{content: []byte("")},
		{content: []byte("=")},
		{content: []byte(" = ")},
		{content: []byte("  =  ")},
		{
			expectedValue: keyValue,
			content:       keyValuePair("", keyValue),
		},
		{
			expectedValue: "=",
			content:       keyValuePair("", "="),
		},
		{
			expectedName:  keyName,
			expectedValue: keyValue,
			content:       fmt.Appendf(nil, "%s=%s", keyName, keyValue),
		},
		{
			expectedName:  keyName,
			expectedValue: keyValue,
			content:       fmt.Appendf(nil, "%s= %s", keyName, keyValue),
		},
		{
			expectedName:  keyName,
			expectedValue: keyValue,
			content:       fmt.Appendf(nil, "%s =%s", keyName, keyValue),
		},
		{
			expectedName:  keyName,
			expectedValue: keyValue,
			content:       fmt.Appendf(nil, "%s = %s", keyName, keyValue),
		},
		{
			expectedName:  keyName,
			expectedValue: keyValue,
			content:       fmt.Appendf(nil, "%s  =  %s", keyName, keyValue),
		},
		{
			expectedName:  keyName,
			expectedValue: keyValue,
			content:       fmt.Appendf(nil, "%s=  %s", keyName, keyValue),
		},
		{
			expectedName:  keyName,
			expectedValue: keyValue,
			content:       fmt.Appendf(nil, "%s  =%s", keyName, keyValue),
		},
	}

	for i, tc := range testCases {
		v := VarFromArg(tc.content)
		r.Exactly(t, tc.expectedName, v.Name(), "names do not match: case=%d, expected=%s, actual=%s", i, tc.expectedName, v.Name())
		r.Exactly(t, tc.expectedValue, v.Value(), "names do not match: case=%d, expected=%s, actual=%s", i, tc.expectedValue, v.Value())
	}
}

func TestGetLeadingWhitespaceSpaces(t *testing.T) {
	t.Parallel()

	s := GetLeadingWhitespace([]byte("   test"))
	r.Exactly(t, []byte("   "), s)
}

func TestGetLeadingWhitespaceTabs(t *testing.T) {
	t.Parallel()

	s := GetLeadingWhitespace([]byte("\t\ttest"))
	r.Exactly(t, []byte("\t\t"), s)
}
