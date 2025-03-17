package interpreter

import (
	"bytes"
	"crypto/sha1"
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

	rootTestDir := filepath.Join("testdata", "imports", "in")
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
		InPath:  filepath.Join(rootTestDir, "in", "startOk.txt"),
		OutPath: filepath.Join(rootTestDir, "out", "finished.txt"),
		Indent:  true,
		NoStats: true,
	})
	err := ip.Start()
	r.NoError(t, err)

	ok, err := areFileContentsTheSame(filepath.Join("testdata", "imports", "gold", "finished.txt"), ip.opts.OutPath)
	r.NoError(t, err)
	r.True(t, ok)
}

func TestIgnore(t *testing.T) {
	t.Parallel()

	l := log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	ip := New(l, &Options{
		InPath:  filepath.Join("testdata", "ignore", "in", "ignore.yaml"),
		OutPath: filepath.Join("testdata", "ignore", "out", "ignore.yaml"),
		Indent:  true,
		NoStats: true,
	})

	err := ip.Start()
	r.NoError(t, err)

	ok, err := areFileContentsTheSame(filepath.Join("testdata", "ignore", "gold", "ignore.yaml"), ip.opts.OutPath)
	r.NoError(t, err)
	r.True(t, ok)
}

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

	r.NoError(t, ip.core.Interpret(nil, core.InterpreterFile{
		Name:   rootFileIn,
		RC:     f,
		Writer: out,
	}))
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

//
// Helper
//

func areFileContentsTheSame(expectedPath, actualPath string) (equal bool, err error) {
	expectedContent, err := os.ReadFile(expectedPath)
	if err != nil {
		return
	}
	actualContent, err := os.ReadFile(actualPath)
	if err != nil {
		return
	}

	expected := sha1.Sum(expectedContent)
	actual := sha1.Sum(actualContent)

	equal = bytes.Equal(expected[:], actual[:])
	return
}
