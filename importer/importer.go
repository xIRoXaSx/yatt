package importer

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Importer struct {
	opts     *Options
	prefixes []string
	state    state
}

type Options struct {
	InPath  string
	OutPath string
	Indent  bool
	NoStats bool
}

type state struct {
	ignoreIndex map[string]int8
	vars        map[string][]command
	*sync.Mutex
}

type command struct {
	name  string
	value string
}

func New(opts *Options) Importer {
	return Importer{
		opts:     opts,
		prefixes: []string{"#import", "# import"},
		state: state{
			ignoreIndex: map[string]int8{},
			vars:        map[string][]command{},
			Mutex:       &sync.Mutex{},
		},
	}
}

func (i *Importer) Start() (err error) {
	stat, err := os.Stat(i.opts.InPath)
	if err != nil {
		log.Fatalln(err)
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
	if err != nil {
		log.Fatalf("unable to walk dirs: %v\n", err)
	}
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
