package interpreter

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/xiroxasx/fastplate/internal/common"
)

const varFileName = "fastplate.var"

type Interpreter struct {
	opts       *Options
	prefixes   []string
	state      state
	lineEnding []byte
}

type Options struct {
	InPath        string
	OutPath       string
	FileWhitelist []string
	FileBlacklist []string
	VarFilePaths  []string
	Indent        bool
	UseCRLF       bool
	NoStats       bool
	Verbose       bool
}

type indexers map[string]indexer

type indexer struct {
	start int
	len   int
	mx    *sync.Mutex
}

type scopedRegistry struct {
	scopedVars map[string][]common.Variable
	*sync.Mutex
}

type state struct {
	ignoreIndex        map[string]int8
	scopedRegistry     scopedRegistry
	dependencies       map[string][]string
	unscopedVarIndexes indexers
	unscopedVars       []common.Variable
	foreach            sync.Map
	dirMode            bool
	buf                *bytes.Buffer
	*sync.Mutex
}

func defaultImportPrefixes() []string {
	return []string{"#fastplate", "# fastplate", "//fastplate", "// fastplate"}
}

func New(opts *Options) (i *Interpreter) {
	i = &Interpreter{
		opts:       opts,
		prefixes:   defaultImportPrefixes(),
		lineEnding: []byte("\n"),
		state: state{
			ignoreIndex: map[string]int8{},
			scopedRegistry: scopedRegistry{
				scopedVars: map[string][]common.Variable{},
				Mutex:      &sync.Mutex{},
			},
			dependencies:       map[string][]string{},
			unscopedVarIndexes: indexers{},
			foreach:            sync.Map{},
			buf:                &bytes.Buffer{},
			Mutex:              &sync.Mutex{},
		},
	}

	if opts.UseCRLF {
		i.lineEnding = []byte("\r\n")
	}

	// Look in the current working directory.
	vFiles := opts.VarFilePaths
	if len(vFiles) == 0 {
		_, err := os.Stat(varFileName)
		if err == nil {
			vFiles = []string{varFileName}
		}
	}

	// Check if the global var files exist and read it into the memory.
	for _, vf := range vFiles {
		_, err := os.Stat(vf)
		if err != nil {
			return
		}
		cont, err := os.ReadFile(vf)
		if err != nil {
			log.Fatal().Err(err).Str("path", vf).Msg("unable to read global variable file")
		}
		lines := bytes.Split(cont, i.lineEnding)
		for _, l := range lines {
			split := bytes.Split(i.CutPrefix(l), []byte{' '})
			if string(split[0]) != commandVar {
				continue
			}
			// Skip the var declaration keyword.
			i.setUnscopedVar(strings.TrimSuffix(filepath.Base(vf), filepath.Ext(vf)), split[1:])
		}
	}
	return
}

func (i *Interpreter) TrimLine(b, prefix []byte) []byte {
	return bytes.Trim(bytes.TrimPrefix(b, prefix), string(i.lineEnding)+" ")
}

func (i *Interpreter) CutPrefix(b []byte) (ret []byte) {
	prefix := i.matchedImportPrefix(b)
	if prefix == nil {
		return
	}
	return i.TrimLine(b, prefix)
}

func (i *Interpreter) Start() (err error) {
	stat, err := os.Stat(i.opts.InPath)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to stat input path")
	}

	start := time.Now()
	i.state.dirMode = stat.IsDir()
	if i.state.dirMode {
		err = i.runDirMode()
	} else {
		err = i.runFileMode()
	}
	el := time.Since(start)
	if err != nil {
		return
	}

	if !i.opts.NoStats {
		fmt.Println("Execution took", el)
	}
	return
}

// runDirMode runs the import for each file inside the Options.InPath.
func (i *Interpreter) runDirMode() (err error) {
	const dirPerm = os.FileMode(0700)

	err = os.MkdirAll(i.opts.OutPath, dirPerm)
	if err != nil {
		return
	}

	err = filepath.WalkDir(i.opts.InPath, func(inPath string, d os.DirEntry, err error) error {
		dest := strings.ReplaceAll(inPath, i.opts.InPath, i.opts.OutPath)
		if d.IsDir() {
			if dest == "" {
				return nil
			}
			err = os.MkdirAll(dest, dirPerm)
			if err != nil {
				return err
			}
			return nil
		}

		out, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0700)
		if err != nil {
			return err
		}
		defer func() {
			cErr := out.Close()
			if err == nil {
				err = cErr
			}
		}()

		isRaw, err := i.rawCopyOnListMatch(inPath, out)
		if err != nil {
			return err
		}
		if isRaw {
			return nil
		}

		// Write to the buffer to ensure that files don't get partially written.
		err = i.interpretFile(inPath, nil)
		if err != nil {
			return err
		}

		// Write buffer to the file and cut last new line.
		_, err = out.Write(i.state.buf.Bytes()[:i.state.buf.Len()-1])
		if err != nil {
			return err
		}
		i.state.buf.Reset()
		return err
	})
	return
}

// runFileMode runs the import with the targeted Options.OutPath.
func (i *Interpreter) runFileMode() (err error) {
	out, err := os.OpenFile(i.opts.OutPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0700)
	if err != nil {
		return err
	}
	defer func() {
		cErr := out.Close()
		if err == nil {
			err = cErr
		}
	}()

	// Write to the buffer to ensure that files don't get partially written.
	err = i.interpretFile(i.opts.InPath, nil)
	if err != nil {
		return
	}

	// Write buffer to the file and cut last new line.
	_, err = out.Write(i.state.buf.Bytes()[:i.state.buf.Len()-1])
	return
}

func (i *Interpreter) interpretFile(file string, indent []byte) (err error) {
	cont, err := os.ReadFile(file)
	if err != nil {
		log.Warn().Err(err).Str("file", file).Msg("unable to read file")
		return
	}

	// Append indention to all linebreaks, prepend to the first line.
	cutSet := i.lineEnding
	if len(indent) > 0 {
		cont = bytes.ReplaceAll(cont, cutSet, append(cutSet, indent...))
		cont = append(indent, cont...)
	}

	lines := bytes.Split(cont, cutSet)
	for lineNum, l := range lines {
		lineNum++
		if i.opts.Indent {
			indent = leadingIndents(l)
		}

		callID := fmt.Sprintf("%s:%d", file, lineNum)

		// Skip the indents.
		linePart := l[len(indent):]
		prefix := i.matchedImportPrefix(linePart)
		if prefix == nil {
			// Line does not contain one of the required prefixes.
			if i.state.ignoreIndex[file] == 1 {
				// Still in an ignore block.
				continue
			}

			var fe foreach
			fe, err = i.state.foreachLoad(file)
			if err != nil && err != errMapLoadForeach {
				return
			}

			if err == nil && fe.buf.v != nil && fe.c.p >= 0 && len(fe.buf.v[fe.c.p].variables) > 0 {
				// Currently moving inside a foreach loop.
				err = i.appendForeachLine(file, l)
				if err != nil {
					return
				}
				continue
			}

			var ret []byte
			ret, err = i.resolve(file, l, nil)
			if err != nil {
				return
			}
			_, err = i.state.buf.Write(append(ret, cutSet...))
			if err != nil {
				return
			}
		} else {
			// Trim statement and check against internal commands.
			statement := i.TrimLine(linePart, prefix)
			split := bytes.Split(statement, []byte{' '})
			if len(split) > 0 && string(split[0]) != commandImport {
				err = i.executeCommand(string(split[0]), file, split[1:], lineNum, callID)
				if err != nil {
					return
				}
				continue
			}

			if len(split) < 2 {
				err = errors.New("no import path given")
				return
			}
			s := filepath.Clean(string(split[1]))
			file = filepath.Clean(file)
			if i.state.hasCyclicDependency(file, s) {
				err = fmt.Errorf("detected import cycle: %s -> %s", file, s)
				return
			}
			i.state.addDependency(file, s)
			err = i.interpretFile(s, indent)
			if err != nil {
				return err
			}
		}
	}
	return
}

func (i *Interpreter) matchedImportPrefix(line []byte) []byte {
	for _, pref := range i.prefixes {
		if bytes.HasPrefix(line, []byte(pref)) {
			return []byte(pref)
		}
	}
	return nil
}

func (i *Interpreter) rawCopyOnListMatch(inPath string, out io.Writer) (isRaw bool, err error) {
	if i.opts == nil {
		return
	}

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

	if i.matchedWhitelist(inPath) {
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

func leadingIndents(line []byte) (s []byte) {
	for _, r := range line {
		if r != ' ' && r != '\t' {
			break
		}
		s = append(s, r)
	}
	return
}
