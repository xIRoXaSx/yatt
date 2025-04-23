package interpreter

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	r "github.com/stretchr/testify/require"
	"github.com/xiroxasx/fastplate/internal/interpreter/core"
)

func TestFileInterpretation(t *testing.T) {
	rootDir := filepath.Join("testdata", "interpret")
	rootInDir := filepath.Join(rootDir, "in")
	rootOutDir := filepath.Join(rootDir, "out")
	l := log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	ip := New(l, &Options{
		VarFilePaths: []string{filepath.Join(rootInDir, "fastplate.var")},
		Indent:       true,
		NoStats:      true,
	})

	// file gets closed by the Interpret call.
	rootFileIn := filepath.Join(rootInDir, "rootfile.yaml")
	f, err := os.Open(rootFileIn)
	r.NoError(t, err)

	out, err := os.OpenFile(filepath.Join(rootOutDir, "rootfile.yaml"), os.O_TRUNC|os.O_CREATE|os.O_RDWR, 0o755)
	r.NoError(t, err)
	defer func() {
		out.Close()
	}()

	err = os.Setenv("TEST", "environment_variable")
	r.NoError(t, err)
	r.NoError(t, ip.core.Interpret(core.InterpreterFile{
		Name: rootFileIn,
		RC:   f,
		Buf:  out,
	}))
}

func TestStart(t *testing.T) {
	rootDir := filepath.Join("testdata", "interpret")
	rootInDir := filepath.Join(rootDir, "in")
	rootOutDir := filepath.Join(rootDir, "out")
	l := log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	ip := New(l, &Options{
		InPath:       filepath.Join(rootInDir, "rootfile.yaml"),
		OutPath:      filepath.Join(rootOutDir, "rootfile.yaml"),
		VarFilePaths: []string{filepath.Join(rootInDir, "fastplate.var")},
		Indent:       true,
		NoStats:      true,
	})

	out, err := os.OpenFile(filepath.Join(rootOutDir, "rootfile.yaml"), os.O_TRUNC|os.O_CREATE|os.O_RDWR, 0o755)
	r.NoError(t, err)
	defer func() {
		out.Close()
	}()

	err = os.Setenv("TEST", "environment_variable")
	r.NoError(t, err)
	r.NoError(t, ip.Start())
}

//
// Benchmarks
//

func BenchmarkFileInterpretation(b *testing.B) {
	rootDir := filepath.Join("testdata", "interpret", "in")
	l := log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	ip := New(l, &Options{
		VarFilePaths: []string{filepath.Join(rootDir, "fastplate.var")},
		Indent:       true,
		NoStats:      true,
	})

	// file gets closed by the Interpret call.
	path := filepath.Join(rootDir, "rootfile.yaml")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		f, err := os.Open(path)
		r.NoError(b, err)

		b.StartTimer()
		r.NoError(b, ip.core.Interpret(core.InterpreterFile{
			Name: path,
			RC:   f,
			Buf:  &bytes.Buffer{},
		}))
	}
}

func BenchmarkFileWrites(b *testing.B) {
	var testDir = filepath.Join("testdata", "interpret")
	l := log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	ip := New(l, &Options{
		InPath:       filepath.Join(testDir, "in", "rootfile.yaml"),
		OutPath:      filepath.Join(testDir, "out", "rootfile.yaml"),
		VarFilePaths: []string{filepath.Join(testDir, "in", "fastplate.var")},
		Indent:       true,
		NoStats:      true,
	})
	r.NoError(b, os.MkdirAll(testDir, 0700))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.NoError(b, ip.Start())
	}
}
