package core

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/rs/zerolog"
	"github.com/xiroxasx/fastplate/internal/common"
)

const (
	preprocessorImportName = "import"

	variableGlobalKey     = "FASTPLATE_GLOBAL"
	variableGlobalKeyFile = variableGlobalKey + "_"
)

var (
	lineEnding         = common.LineEnding()
	templateStartBytes = common.TemplateStart()
	templateEndBytes   = common.TemplateEnd()

	errEmptyVariableParameter  = errors.New("variable name or value must not be empty")
	errDependencyUnknownSyntax = fmt.Errorf("unknown syntax: %s <file path>", preprocessorImportName)
)

type Core struct {
	l        zerolog.Logger
	prefixes [][]byte
	opts     Options

	ignoreIndex  ignoreIndexes
	depsResolver dependencyResolver
	foreachBuff  foreachBufferStack
	// foreachIndex       foreachIndexes
	varRegistryLocal      variableRegistry
	varRegistryGlobal     variableRegistry // TODO: Currently merging unscopedVarIndexes into this as well!
	varRegistryGlobalFile variableRegistry // TODO: Currently merging unscopedVarIndexes into this as well!

	*sync.Mutex
}

type Options struct {
	PreserveIndent bool
}

type ignoreIndexes map[string]ignoreState

type variableRegistry struct {
	entries map[string]vars
	*sync.Mutex
}

type vars []common.Variable

type parserFunc []byte

func (i parserFunc) string() string {
	return string(i)
}

type InterpreterFile struct {
	Name   string
	Writer io.Writer
	RC     io.ReadCloser
}

func New(l zerolog.Logger, prefixes []string, opts Options) *Core {
	ps := make([][]byte, len(prefixes))
	for i := range prefixes {
		ps[i] = []byte(prefixes[i])
	}

	return &Core{
		l:            l.With().Str("mod", "core").Logger(),
		opts:         opts,
		prefixes:     ps,
		ignoreIndex:  make(ignoreIndexes, 0),
		foreachBuff:  newForeachBufferStack(""),
		Mutex:        &sync.Mutex{},
		depsResolver: newDependencyResolver(),
		varRegistryLocal: variableRegistry{
			entries: make(map[string]vars, 0),
			Mutex:   &sync.Mutex{},
		},
		varRegistryGlobal: variableRegistry{
			entries: make(map[string]vars, 0),
			Mutex:   &sync.Mutex{},
		},
		varRegistryGlobalFile: variableRegistry{
			entries: make(map[string]vars, 0),
			Mutex:   &sync.Mutex{},
		},
	}
}

func (c *Core) Interpret(parentLineIndent []byte, file InterpreterFile) (err error) {
	return c.interpret(file, parentLineIndent)
}

// interpret tries to interpret the scanned content of file.rc.
// If the ReadCloser content contains available tokens, it tries to resolve them and writes it,
// along with the prepended indentParent, to buf.
func (c *Core) interpret(file InterpreterFile, parentLineIndent []byte) (err error) {
	// Always ensure to close the file's rc.
	defer func() {
		cErr := file.RC.Close()
		if cErr != nil {
			if err == nil {
				err = cErr
				return
			}
			c.l.Err(cErr).Str("file", file.Name).Msg("closing file reader")
		}
	}()

	var (
		// Currently read line.
		lineNum int
		// Limits reads to 65536 bytes per line.
		scanner = bufio.NewScanner(file.RC)
	)
	for scanner.Scan() {
		err = scanner.Err()
		if err != nil {
			return
		}

		line := scanner.Bytes()
		currentLineIndent := make([]byte, 0)
		if c.opts.PreserveIndent {
			// Line indents are required, check current line indents.
			currentLineIndent = common.GetLeadingWhitespace(line)
		}

		err = c.searchTokensAndExecute(file.Name, line, currentLineIndent, parentLineIndent, file.Writer, lineNum+1)
		if err != nil {
			return
		}
	}

	return
}

//
// Helper functions.
//

func replaceVar(line, varName, replacement []byte) []byte {
	matched := bytes.Join([][]byte{templateStartBytes, varName, templateEndBytes}, nil)
	return bytes.ReplaceAll(line, matched, replacement)
}

func remapArgsWithVariables(fncNameStr string, varsFromArgs, additionalVars []common.Variable) (values [][]byte, err error) {
	values = make([][]byte, len(varsFromArgs))

additionalVar:
	for idx := range varsFromArgs {
		for _, av := range additionalVars {
			// Overwrite variable value if the names match.
			// This may be the case for "foreach"-variables.
			if varsFromArgs[idx].Name() == av.Name() {
				values[idx] = []byte(av.Value())
				continue additionalVar
			}
		}

		// Keep variable name intact so the function call can retrieve the var's value.
		if fncNameStr == "var" {
			values[idx] = []byte(varsFromArgs[idx].Name())
			continue
		}
		values[idx] = []byte(varsFromArgs[idx].Value())
	}

	return
}

func (c *Core) cutPrefix(b []byte) (ret []byte) {
	prefix := c.matchedPrefixToken(b)
	if prefix == nil {
		return
	}
	return trimLine(b, prefix)
}

func trimLine(b, prefix []byte) []byte {
	return bytes.TrimPrefix(bytes.TrimSpace(b), append(prefix, ' '))
}

func (c *Core) matchedPrefixToken(line []byte) (prefix []byte) {
	for _, pref := range c.prefixes {
		if bytes.HasPrefix(bytes.TrimSpace(line), []byte(pref)) {
			return []byte(pref)
		}
	}
	return
}
