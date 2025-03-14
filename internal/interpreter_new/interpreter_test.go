package interpreter

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	r "github.com/stretchr/testify/require"
	"github.com/xiroxasx/fastplate/internal/interpreter_new/core"
)

func TestInterpreterImportCycle(t *testing.T) {
	t.Parallel()

	l := log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	ip := New(l, &Options{})

	rootTestDir := filepath.Join("testdata", "imports")
	err := ip.core.ImportPathCheckCyclicDependencies(filepath.Join(rootTestDir, "startOk.txt"))
	r.NoError(t, err)

	ip = New(l, &Options{})
	err = ip.core.ImportPathCheckCyclicDependencies(filepath.Join(rootTestDir, "startFail.txt"))
	r.Error(t, err)
}

func TestInterpreterImport(t *testing.T) {
	t.Parallel()

	l := log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	rootTestDir := filepath.Join("testdata", "imports")
	ip := New(l, &Options{
		InPath:  filepath.Join(rootTestDir, "startOk.txt"),
		OutPath: filepath.Join(rootTestDir, "bin", "finished.txt"),
		Indent:  true,
	})
	err := ip.Start()
	r.NoError(t, err)
}

func BenchmarkFileInterpretation(b *testing.B) {
	l := log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	ip := New(l, &Options{
		VarFilePaths: []string{filepath.Join("testdata", "fastplate.var")},
		Indent:       true,
		NoStats:      true,
	})

	path := filepath.Join("testdata", "src", "rootfile.yaml")
	f, err := os.Open(path)
	r.NoError(b, err)
	defer func() {
		f.Close()
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.NoError(b, ip.core.Interpret(nil, core.InterpreterFile{
			Name:   path,
			RC:     f,
			Writer: &bytes.Buffer{},
		}))
	}
}

func BenchmarkFileWrites(b *testing.B) {
	var testDir = filepath.Join("testdata", "dest")
	l := log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	imp := New(l, &Options{
		InPath:       filepath.Join("testdata", "src", "rootfile.yaml"),
		OutPath:      filepath.Join(testDir, "rootfile.yaml"),
		VarFilePaths: []string{filepath.Join("testdata", "fastplate.var")},
		Indent:       true,
		NoStats:      true,
	})
	r.NoError(b, os.MkdirAll(testDir, 0700))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.NoError(b, imp.Start())
	}
}
