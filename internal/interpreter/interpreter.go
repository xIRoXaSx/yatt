package interpreter

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/xiroxasx/yatt/internal/core"
)

type Interpreter struct {
	l    zerolog.Logger
	core *core.Core

	opts *Options
}

type Options struct {
	InPath        string
	OutPath       string
	FileWhitelist []string
	FileBlacklist []string
	VarFilePaths  []string
	Indent        bool
	NoStats       bool
	Verbose       bool
}

func defaultPrefixTokens() []string {
	const prefixName = "yatt"

	return []string{
		fmt.Sprintf("#%s", prefixName),
		fmt.Sprintf("# %s", prefixName),
		fmt.Sprintf("//%s", prefixName),
		fmt.Sprintf("// %s", prefixName),
	}
}

func New(l zerolog.Logger, opts *Options) (i *Interpreter) {
	i = &Interpreter{
		opts: opts,
		l:    l,
		core: core.New(l, defaultPrefixTokens(), core.Options{
			PreserveIndent: opts.Indent,
		}),
	}

	i.initScopedVars()
	return
}

func (i *Interpreter) Start() (err error) {
	i.opts.InPath = filepath.Clean(i.opts.InPath)

	stat, err := os.Stat(i.opts.InPath)
	if err != nil {
		i.l.Fatal().Err(err).Msg("unable to stat input path")
	}

	start := time.Now()
	defer func() {
		if err != nil {
			return
		}

		elapsed := time.Since(start)
		if !i.opts.NoStats {
			i.l.Info().Dur("elapsed", elapsed).Msg("finished")
		}
	}()

	if stat.IsDir() {
		i.opts.OutPath = filepath.Clean(i.opts.OutPath)
		err = os.MkdirAll(i.opts.OutPath, 0o755)
		if err != nil {
			return
		}

		err = i.runDirMode(i.opts.InPath, i.opts.OutPath)
		return
	}

	outDir := filepath.Dir(filepath.Clean(i.opts.OutPath))
	err = os.MkdirAll(outDir, 0o755)
	if err != nil {
		return
	}
	err = i.runFileMode(i.opts.InPath, i.opts.OutPath)
	return
}

func (i *Interpreter) initScopedVars() {
	// Cleanup filepaths.
	vFiles := i.opts.VarFilePaths
	for i, vFile := range vFiles {
		vFiles[i] = filepath.Clean(vFile)
	}

	i.core.InitGlobalVariablesByFiles(vFiles...)
}

func (i *Interpreter) writeInterpretedFile(inPath, outPath string) (err error) {
	out, err := os.OpenFile(outPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0700)
	if err != nil {
		return err
	}
	defer func() {
		cErr := out.Close()
		if err == nil {
			err = cErr
		}
	}()

	// Copy file contents if the current file is matching the filters,
	// we don't need to interpret them.
	isRaw, err := i.rawCopyOnListMatch(inPath, out)
	if err != nil {
		return err
	}
	if isRaw {
		return nil
	}

	// Open the input file.
	// The interpret method will close it afterwards.
	inFile, err := os.Open(inPath)
	if err != nil {
		return
	}

	buf := &bytes.Buffer{}
	interFile := core.InterpreterFile{
		Name: inPath,
		RC:   inFile,
		Buf:  buf,
	}
	// Write to the buffer to ensure that files don't get partially written.
	err = i.core.Interpret(interFile)
	if err != nil {
		return
	}

	if buf.Len() > 0 {
		// Write buffer to the file and cut last new line.
		_, err = out.Write(buf.Bytes()[:buf.Len()-1])
	}
	return
}

func (i *Interpreter) rawCopyOnListMatch(inPath string, out io.Writer) (isRaw bool, err error) {
	writeTo := func(inPath string, out io.Writer) (err error) {
		var b []byte
		b, err = os.ReadFile(inPath)
		if err != nil {
			return
		}

		// Write file content to the output.
		_, err = out.Write(b)
		if err != nil {
			return
		}
		return
	}

	if i.matchedBlacklist(inPath) {
		log.Debug().Str("file", inPath).Msg("matched blacklist, plain copy")
		isRaw = true
		err = writeTo(inPath, out)
		return
	}

	if len(i.opts.FileWhitelist) == 0 || i.matchedWhitelist(inPath) {
		return
	}

	log.Debug().Str("file", inPath).Msg("does not match whitelist, plain copy")
	isRaw = true
	err = writeTo(inPath, out)
	return
}

func (i *Interpreter) matchedBlacklist(v string) (matched bool) {
	if i.opts == nil {
		return
	}

	for _, f := range i.opts.FileBlacklist {
		reg := regexp.MustCompile(f)
		m := reg.MatchString(v)
		if m {
			return true
		}
	}

	return
}

func (i *Interpreter) matchedWhitelist(v string) (matched bool) {
	if i.opts == nil {
		return
	}

	for _, f := range i.opts.FileWhitelist {
		reg := regexp.MustCompile(f)
		m := reg.MatchString(v)
		if m {
			return true
		}
	}

	return
}
