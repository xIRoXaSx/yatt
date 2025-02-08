package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/xiroxasx/fastplate/internal/interpreter"
)

type MultiString []string

func (vp *MultiString) String() string {
	return strings.Join(*vp, " ")
}

func (vp *MultiString) Set(v string) (err error) {
	*vp = append(*vp, v)
	return
}

func parseFlags() (a interpreter.Options) {
	fileBlackList := make(MultiString, 0)
	fileWhiteList := make(MultiString, 0)
	varFilePaths := make(MultiString, 0)

	flag.BoolVar(&a.Indent, "indent", false, "whether to retain indention or not")
	flag.Var(&fileBlackList, "blacklist", "regex to describe which files should not be interpreted")
	flag.Var(&fileWhiteList, "whitelist", "regex to describe which files should be interpreted")
	flag.BoolVar(&a.NoStats, "no-stats", false, "do not print stats at the end of the execution")
	flag.BoolVar(&a.Verbose, "verbose", false, "print verbosely")
	flag.StringVar(&a.InPath, "in", "", "the root path")
	flag.StringVar(&a.OutPath, "out", "", "the output path. If not used, in will be overwritten")
	flag.BoolVar(&a.UseCRLF, "crlf", false, "whether to split lines by \\r\\n or \\n")
	flag.Var(&varFilePaths, "var", "the optional var file path.")
	flag.Parse()

	a.FileBlacklist = fileBlackList
	a.FileWhitelist = fileWhiteList
	a.VarFilePaths = varFilePaths
	return
}

func main() {
	l := log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	if len(os.Args) == 1 {
		l.Error().Msg("invalid syntax: fastplate <path> [options]")
		return
	}

	opts := parseFlags()
	if opts.OutPath == "" {
		r := bufio.NewReader(os.Stdin)
		fmt.Printf("Are you sure that you want to overwrite %s? [y/N] ", opts.InPath)
		b, err := r.ReadByte()
		if err != nil {
			l.Fatal().Err(err).Msg("unable to read input")
		}
		if bytes.ToLower([]byte{b})[0] != 'y' {
			l.Fatal().Err(err).Msg("canceled")
		}
		opts.OutPath = opts.InPath
	}

	if opts.InPath == "" {
		l.Fatal().Msg("in path needs to be defined")
	}

	opts.InPath = filepath.Clean(opts.InPath)
	opts.OutPath = filepath.Clean(opts.OutPath)

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if opts.Verbose {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	ip := interpreter.New(&opts)
	err := ip.Start()
	if err != nil {
		l.Fatal().Err(err).Msg("error upon execution")
	}
}
