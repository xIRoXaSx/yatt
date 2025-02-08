package interpreter

import (
	"testing"

	r "github.com/stretchr/testify/require"
)

func TestDependenciesAreCyclic(t *testing.T) {
	t.Parallel()

	type importCase struct {
		src string
		imp string
	}

	type testCase struct {
		ic   importCase
		fail bool
	}

	testCases := [][]testCase{
		{
			{ic: importCase{src: "fileA", imp: "fileB"}, fail: true},
			{ic: importCase{src: "fileB", imp: "fileA"}, fail: true},
		},
		{
			{ic: importCase{src: "fileA", imp: "fileB"}, fail: true},
			{ic: importCase{src: "fileB", imp: "fileC"}, fail: true},
			{ic: importCase{src: "fileC", imp: "fileA"}, fail: true},
		},
		{
			{ic: importCase{src: "fileA", imp: "fileB"}, fail: true},
			{ic: importCase{src: "fileB", imp: "fileC"}, fail: true},
			{ic: importCase{src: "fileC", imp: "fileD"}, fail: true},
			{ic: importCase{src: "fileC", imp: "fileE"}, fail: true},
			{ic: importCase{src: "fileE", imp: "fileA"}, fail: true},
		},
		{
			{ic: importCase{src: "fileA", imp: "fileB"}, fail: true},
			{ic: importCase{src: "fileB", imp: "fileC"}, fail: true},
			{ic: importCase{src: "fileC", imp: "fileD"}, fail: true},
			{ic: importCase{src: "fileB", imp: "fileF"}, fail: true},
			{ic: importCase{src: "fileF", imp: "fileH"}, fail: true},
			{ic: importCase{src: "fileB", imp: "fileG"}, fail: true},
			{ic: importCase{src: "fileD", imp: "fileB"}, fail: true},
		},
		{
			{ic: importCase{src: "fileA", imp: "fileB"}, fail: true},
			{ic: importCase{src: "fileB", imp: "fileC"}, fail: true},
			{ic: importCase{src: "fileC", imp: "fileD"}, fail: true},
			{ic: importCase{src: "fileB", imp: "fileF"}, fail: true},
			{ic: importCase{src: "fileF", imp: "fileH"}, fail: true},
			{ic: importCase{src: "fileB", imp: "fileG"}, fail: true},
			{ic: importCase{src: "fileD", imp: "fileC"}, fail: true},
		},
		{
			{ic: importCase{src: "fileA", imp: "fileB"}},
			{ic: importCase{src: "fileB", imp: "fileC"}},
			{ic: importCase{src: "fileC", imp: "fileD"}},
			{ic: importCase{src: "fileC", imp: "fileF"}},
			{ic: importCase{src: "fileC", imp: "fileG"}},
			{ic: importCase{src: "fileF", imp: "fileH"}},
		},
	}

	for i, tcs := range testCases {
		startCase := tcs[0]
		dr := newDependencyResolver()
		for _, tc := range tcs {
			dr.addDependency(tc.ic.src, tc.ic.imp)
		}
		ok := dr.dependenciesAreCyclic(startCase.ic.src, startCase.ic.imp)
		r.Equal(t, ok, startCase.fail, "case=%d", i)
	}
}
