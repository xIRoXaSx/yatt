package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/xiroxasx/fastplate/interpreter"
)

func parseFlags() (a importer.Options) {
	flag.BoolVar(&a.Indent, "indent", false, "whether to retain indention or not")
	flag.BoolVar(&a.NoStats, "no-stats", false, "do not print stats at the end of the execution")
	flag.StringVar(&a.InPath, "in", "", "the root path")
	flag.StringVar(&a.OutPath, "out", "", "the output path. If not used, in will be overwritten")
	flag.StringVar(&a.VarFilePath, "var", "", "the optional var file path.")
	flag.Parse()
	return
}

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

func main() {
	if len(os.Args) == 1 {
		return
	}

	opts := parseFlags()
	if opts.OutPath == "" {
		r := bufio.NewReader(os.Stdin)
		fmt.Printf("Are you sure that you want to overwrite %s? [y/N] ", opts.InPath)
		b, err := r.ReadBytes('\n')
		if err != nil {
			log.Fatal().Err(err).Msg("unable to read input")
		}
		if bytes.ToLower([]byte{b[0]})[0] != 'y' {
			log.Fatal().Err(err).Msg("canceled")
		}
		opts.OutPath = opts.InPath
	}

	if opts.InPath == "" {
		log.Fatal().Msg("in path needs to be defined")
	}

	opts.InPath = filepath.Clean(opts.InPath)
	opts.OutPath = filepath.Clean(opts.OutPath)

	imp := importer.New(&opts)
	err := imp.Start()
	if err != nil {
		log.Fatal().Err(err).Msg("error upon execution")
	}
}
