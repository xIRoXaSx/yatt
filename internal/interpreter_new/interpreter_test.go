package interpreter

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	r "github.com/stretchr/testify/require"
)

func TestInterpreterImportCycle(t *testing.T) {
	t.Parallel()

	l := log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	ip := New(&Options{}, l)

	rootTestDir := filepath.Join("testdata", "imports")
	err := ip.importPathCheckCyclicDependencies(filepath.Join(rootTestDir, "startOk.txt"))
	r.NoError(t, err)

	ip = New(&Options{}, l)
	err = ip.importPathCheckCyclicDependencies(filepath.Join(rootTestDir, "startFail.txt"))
	r.Error(t, err)
}

func TestInterpreterImport(t *testing.T) {
	t.Parallel()

	l := log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	rootTestDir := filepath.Join("testdata", "imports")
	ip := New(&Options{
		InPath:  filepath.Join(rootTestDir, "startOk.txt"),
		OutPath: filepath.Join(rootTestDir, "bin", "finished.txt"),
	}, l)
	err := ip.Start()
	r.NoError(t, err)
}
