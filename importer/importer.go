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
	unscopedVars []variable
	*sync.Mutex
}

func (s state) lookupScoped(fileName, name string) variable {
	for _, v := range s.scopedVars[fileName] {
		if v.name == name {
			return v
		}
	}
	return variable{}
}

func (s state) lookupUnScoped(name string) variable {
	for _, v := range s.unscopedVars {
		if v.name == name {
			return v
		}
	}
	return variable{}
}

type variable struct {
	name  string
	value string
}

func defaultImportPrefixes() []string {
	return []string{"#import", "# import"}
}

func New(opts *Options) (i Importer) {
	i = Importer{
		opts:     opts,
		prefixes: defaultImportPrefixes(),
		state: state{
			ignoreIndex: map[string]int8{},
			scopedVars:  map[string][]variable{},
			Mutex:       &sync.Mutex{},
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
		split := bytes.Split(i.StripPrefix(l), []byte{' '})
		if !bytes.Equal(split[0], []byte(commandVar)) {
			continue
		}

		// Skip the var declaration keyword and line end.
		sub := split[1:]
		str := make([]string, len(sub))
		for j, s := range sub {
			str[j] = string(s)
		}
		i.setUnScopedVar(str)
	}
	return
}

func (i *Importer) StripPrefix(b []byte) (ret []byte) {
	cutSet := []byte{'\n'}
	for _, p := range i.prefixes {
		noPrefix := bytes.TrimPrefix(b, []byte(p))
		if bytes.Equal(noPrefix, b) {
			continue
		}
		ret = bytes.Trim(noPrefix, string(append(cutSet, ' ')))
		return
	}
	return
}

func (i *Importer) Start() (err error) {
	stat, err := os.Stat(i.opts.InPath)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to stat input path")
	}

	start := time.Now()
	if stat.IsDir() {
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

		err = i.interpretFile(inPath, nil, out)
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

	buf := &bytes.Buffer{}
	err = i.interpretFile(i.opts.InPath, nil, buf)
	if err != nil {
		return
	}
	_, err = out.Write(buf.Bytes())
	return
}
