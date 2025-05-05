package core

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/rs/zerolog"
	"github.com/xiroxasx/yatt/internal/common"
	"github.com/xiroxasx/yatt/internal/foreach"
)

const (
	preprocessorImportName = "import"

	variableGlobalKey     = "YATT_GLOBAL"
	variableGlobalKeyFile = variableGlobalKey + "_"
)

var (
	lineEnding         = common.LineEnding()
	templateStartBytes = common.TemplateStart()
	templateEndBytes   = common.TemplateEnd()

	errEmptyVariableParameter  = errors.New("variable name or value must not be empty")
	errDependencyCyclic        = errors.New("cyclic dependency detected")
	errDependencyUnknownSyntax = fmt.Errorf("unknown syntax: %s <file path>", preprocessorImportName)
)

// Core must implement the foreach.TokenResolver interface.
var _ foreach.TokenResolver = &Core{}

type Core struct {
	l        zerolog.Logger
	prefixes [][]byte
	opts     Options

	ignoreIndex  ignoreIndexes
	depsResolver dependencyResolver
	feb          foreach.Buffer

	registries

	*sync.Mutex
}

type Options struct {
	PreserveIndent bool
}

type ignoreIndexes map[string]ignoreState

type registries struct {
	varRegistryForeach variableRegistry
	varRegistryLocal   variableRegistry
	varRegistryGlobal  variableRegistry
}

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
	Name string
	Buf  io.Writer
	RC   io.ReadCloser
}

func New(l zerolog.Logger, prefixes []string, opts Options) *Core {
	ps := make([][]byte, len(prefixes))
	for i := range prefixes {
		ps[i] = []byte(prefixes[i])
	}

	newVarReg := func() variableRegistry {
		return variableRegistry{
			entries: make(map[string]vars, 0),
			Mutex:   &sync.Mutex{},
		}
	}

	return &Core{
		l:            l.With().Str("mod", "core").Logger(),
		opts:         opts,
		prefixes:     ps,
		ignoreIndex:  make(ignoreIndexes, 0),
		feb:          foreach.NewForeachBuffer(lineEnding),
		Mutex:        &sync.Mutex{},
		depsResolver: newDependencyResolver(),
		registries: registries{
			varRegistryForeach: newVarReg(),
			varRegistryLocal:   newVarReg(),
			varRegistryGlobal:  newVarReg(),
		},
	}
}

func (c *Core) VarsLookupGlobalFile(name string) []common.Variable {
	return c.varsLookupGlobalFile(name)
}

func (c *Core) VarsLookupGlobal() []common.Variable {
	return varsLookupRegistry(&c.varRegistryGlobal)
}

func (c *Core) Interpret(file InterpreterFile) (err error) {
	return c.interpret(file, nil)
}

// Implement foreach.TokenResolver interface.
func (c *Core) Resolve(fileName string, l []byte, vars ...common.Variable) (ret []byte, err error) {
	return c.resolve(resolveArgs{
		fileName:       fileName,
		line:           l,
		additionalVars: vars,
	})
}

// interpret tries to interpret the scanned content of file.rc.
// If the ReadCloser content contains available tokens, it tries to resolve them and writes it,
// along with the prepended indentParent, to buf.
func (c *Core) interpret(file InterpreterFile, additionalIndent []byte) (err error) {
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

		lineNum++
		line := scanner.Bytes()
		currentLineIndent := make([]byte, 0)
		if c.opts.PreserveIndent {
			// Line indents are required, check current line indents.
			lineIndet := common.GetLeadingWhitespace(line)
			currentLineIndent = append(lineIndet, additionalIndent...)
			line = line[len(lineIndet):]
		}

		err = c.searchTokensAndExecute(file.Name, line, currentLineIndent, file.Buf, lineNum)
		if err != nil {
			return
		}
	}

	return
}

//
// Helper functions.
//

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
