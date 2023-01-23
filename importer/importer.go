package importer

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

const varFileName = "gport.var"

type Importer struct {
	opts     *Options
	prefixes []string
	state    state
}

type Options struct {
	InPath      string
	OutPath     string
	VarFilePath string
	Indent      bool
	NoStats     bool
}

type state struct {
	ignoreIndex  map[string]int8
	scopedVars   map[string][]variable
	dependencies map[string][]string
	unscopedVars []variable
	dirMode      bool
	*sync.Mutex
}

type variable struct {
	name  string
	value string
}

func defaultImportPrefixes() []string {
	return []string{"#import", "# import"}
}

func (s state) lookupUnscoped(name string) variable {
	for _, v := range s.unscopedVars {
		if v.name == name {
			return v
		}
	}
	return variable{}
}

func (s state) lookupScoped(fileName, name string) variable {
	for _, v := range s.scopedVars[fileName] {
		if v.name == name {
			return v
		}
	}
	return variable{}
}

func (s state) addDependency(fileName, dependency string) {
	s.dependencies[fileName] = append(s.dependencies[fileName], dependency)
}

// hasCyclicDependency walks down the dependencies to check whether the given dependency has creates a loop.
// Returns true if a cycle has been detected.
func (s state) hasCyclicDependency(fileName, dependency string) bool {
	for _, d := range s.dependencies[dependency] {
		if d == fileName {
			return true
		} else if d == "" {
			return false
		}
		return s.hasCyclicDependency(fileName, d)
	}
	return false
}

func (s state) followDependency(dependency, target string) bool {
	for _, d := range s.dependencies[dependency] {
		if d == target {
			return true
		} else if d == "" {
			return false
		}
		return s.followDependency(d, target)
	}
	return false
}

func New(opts *Options) (i Importer) {
	i = Importer{
		opts:     opts,
		prefixes: defaultImportPrefixes(),
		state: state{
			ignoreIndex:  map[string]int8{},
			scopedVars:   map[string][]variable{},
			dependencies: map[string][]string{},
			Mutex:        &sync.Mutex{},
		},
	}

	// Check if the global var file exists and read it into the memory.
	vFile := opts.VarFilePath
	_, err := os.Stat(vFile)
	if err != nil {
		// Look in the current working directory.
		_, err = os.Stat(varFileName)
		if err != nil {
			return
		}
		vFile = varFileName
	}
	cont, err := os.ReadFile(vFile)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to read global import variable file")
	}
	cutSet := []byte{'\n'}
	lines := bytes.Split(cont, cutSet)
	for _, l := range lines {
		split := bytes.Split(i.CutPrefix(l), []byte{' '})
		if string(split[0]) != commandVar {
			continue
		}

		// Skip the var declaration keyword.
		i.setUnscopedVar(split[1:])
	}
	return
}

func (i *Importer) TrimLine(b, prefix []byte) []byte {
	return bytes.Trim(bytes.TrimPrefix(b, prefix), "\n ")
}

func (i *Importer) CutPrefix(b []byte) (ret []byte) {
	prefix := i.matchedImportPrefix(b)
	if prefix == nil {
		return
	}
	return i.TrimLine(b, prefix)
}

func (i *Importer) Start() (err error) {
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
		fmt.Printf("Execution took %v\n", el)
	}
	return
}

// runDirMode runs the import for each file inside the Options.InPath.
func (i *Importer) runDirMode() (err error) {
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

		// Write to the buffer to ensure that files don't get partially written.
		buf := &bytes.Buffer{}
		err = i.interpretFile(inPath, nil, buf)
		if err != nil {
			return err
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

		// Write buffer to the file and cut last new line.
		_, err = out.Write(buf.Bytes()[:buf.Len()-1])
		return err
	})
	return
}

// runFileMode runs the import with the targeted Options.OutPath.
func (i *Importer) runFileMode() (err error) {
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
	buf := &bytes.Buffer{}
	err = i.interpretFile(i.opts.InPath, nil, buf)
	if err != nil {
		return
	}

	// Write buffer to the file and cut last new line.
	_, err = out.Write(buf.Bytes()[:buf.Len()-1])
	return
}
