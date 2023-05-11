package interpreter

import (
	"archive/zip"
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	r "github.com/stretchr/testify/require"
	"github.com/xiroxasx/fastplate/internal/common"
)

type zipWriter struct {
	isClosed bool
	*zip.Writer
}

func (zw *zipWriter) CloseWriter() (err error) {
	if !zw.isClosed {
		zw.isClosed = true
		err = zw.Close()
	}
	return
}

func TestInterpreter(t *testing.T) {
	t.Parallel()

	var (
		testRootFile     = "rootfile.yaml"
		timeSeed         = time.UnixMilli(1673911642000)
		testDestDir      = filepath.Join("testdata", "dest")
		testSrcDir       = filepath.Join("testdata", "src")
		testGoldDir      = filepath.Join("testdata", "gold")
		testGoldDirMode  = filepath.Join(testGoldDir, "dirMode.sum")
		testGoldFileMode = filepath.Join(testGoldDir, "fileMode.sum")
	)
	r.NoError(t, os.MkdirAll(testGoldDir, 0700))
	ip := New(&Options{
		InPath:       testSrcDir,
		OutPath:      testDestDir,
		Indent:       true,
		NoStats:      true,
		VarFilePaths: []string{filepath.Join("testdata", "fastplate.var")},
	})
	r.NoError(t, ip.Start())

	// Run in dir mode.
	testInterpreter(t, ip, timeSeed, testGoldDirMode)

	// Run in file mode.
	ip.opts.InPath = filepath.Join(ip.opts.InPath, testRootFile)
	ip.opts.OutPath = filepath.Join(ip.opts.OutPath, testRootFile)
	testInterpreter(t, ip, timeSeed, testGoldFileMode)
}

func testInterpreter(t *testing.T, ip Interpreter, timeSeed time.Time, goldPath string) {
	buf := &bytes.Buffer{}
	zw := &zipWriter{Writer: zip.NewWriter(buf)}
	defer func() {
		// Make sure the writer gets closed in any case.
		r.NoError(t, zw.CloseWriter())
	}()

	// Write the in-memory zip file.
	r.NoError(t, filepath.WalkDir(ip.opts.OutPath, func(p string, d os.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}
		f, err := os.Open(p)
		r.NoError(t, err)
		defer func() {
			r.NoError(t, f.Close())
		}()

		stat, err := f.Stat()
		r.NoError(t, err)
		w, err := zw.CreateHeader(&zip.FileHeader{
			Name:               f.Name(),
			Modified:           timeSeed,
			UncompressedSize64: uint64(stat.Size()),
		})
		b, err := io.ReadAll(f)
		r.NoError(t, err)
		_, err = w.Write(b)
		r.NoError(t, err)
		return nil
	}))
	r.NoError(t, zw.Flush())
	r.NoError(t, zw.CloseWriter())

	// Comment back in if gold sample(s) need(s) to be updated.
	//r.NoError(t, updateGoldSample(goldPath, buf.Bytes()))

	goldSum, err := os.ReadFile(goldPath)
	r.NoError(t, err)
	r.Exactly(t, goldSum, generateHashSum(buf.Bytes()))
}

func BenchmarkFileInterpretation(b *testing.B) {
	ip := Interpreter{
		opts:     &Options{Indent: true},
		prefixes: defaultImportPrefixes(),
		state: state{
			ignoreIndex: map[string]int8{},
			scopedRegistry: scopedRegistry{
				scopedVars: map[string][]common.Var{},
				Mutex:      &sync.Mutex{},
			},
			unscopedVars: []common.Var{},
			dependencies: map[string][]string{},
			buf:          &bytes.Buffer{},
			Mutex:        &sync.Mutex{},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.NoError(b, ip.interpretFile(filepath.Join("testdata", "src", "rootfile.yaml"), nil))
	}
}

func BenchmarkFileWrites(b *testing.B) {
	var testDir = filepath.Join("testdata", "dest")
	imp := New(&Options{
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

func generateHashSum(b []byte) (sum []byte) {
	h := sha1.New()
	h.Write(b)
	sb := h.Sum(nil)
	sum = make([]byte, hex.EncodedLen(len(sb)))
	hex.Encode(sum, sb)
	return
}

func updateGoldSample(file string, b []byte) (err error) {
	err = os.WriteFile(file, generateHashSum(b), 0500)
	return
}
