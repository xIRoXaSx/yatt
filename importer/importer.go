package importer

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Importer struct {
	opts     *Options
	prefixes []string
}

type Options struct {
	InPath  string
	OutPath string
	Indent  bool
	NoStats bool
}

func New(opts *Options) Importer {
	return Importer{
		opts:     opts,
		prefixes: []string{"#import", "# import"},
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
