package core

import (
	"bytes"
	"io"
	"path/filepath"

	"github.com/xiroxasx/fastplate/internal/common"
)

type resolveArgs struct {
	fileName       string
	line           []byte
	additionalVars []common.Variable
}

func (c *Core) searchTokensAndExecute(fileName string, line, currentLineIndent []byte, buf io.Writer, lineNum int) (err error) {
	prefix := c.matchedPrefixToken(line)
	if len(prefix) > 0 {
		// Trim the prefix and check against internal commands.
		statement := trimLine(line, prefix)
		split := bytes.Split(statement, []byte{' '})
		if len(split) == 0 {
			return
		}

		args := make([][]byte, len(split[1:]))
		for i := range args {
			args[i] = bytes.TrimRight(split[i+1], ",")
		}

		pd := newPreprocessorDirective(
			string(split[0]),
			filepath.Clean(fileName),
			lineNum,
			args,
			currentLineIndent,
		)
		err = c.Preprocess(pd, func(pd *PreprocessorDirective) error {
			return c.importPath(pd)
		})
		if err != nil {
			return
		}

		_, err = pd.WriteTo(buf)
		return
	}

	switch c.preprocessorState(fileName) {
	case RecurringTokenIgnore:
		// Currently moving inside a ignore block, skipping line...
		return

	case RecurringTokenForeach:
		if c.opts.PreserveIndent {
			line = append(currentLineIndent, line...)
		}
		c.feb.WriteLineToBuffer(line)
		return

	default:
		break
	}

	// No prefix found, try to resolve variables and functions if there are any.
	var ret []byte
	ret, err = c.resolve(resolveArgs{
		fileName: fileName,
		line:     line,
	})
	if err != nil {
		return
	}

	// Only prepend indents if line is not empty.
	if len(bytes.TrimSpace(ret)) != 0 {
		ret = append(currentLineIndent, ret...)
	}
	_, err = buf.Write(append(ret, lineEnding...))
	if err != nil {
		return
	}

	return
}

func (c *Core) resolve(rArgs resolveArgs) (ret []byte, err error) {
	ret = rArgs.line
	partials := bytes.Split(rArgs.line, templateStartBytes)
	if len(partials) == 1 {
		// Nothing needs to be resolved.
		return
	}
	for _, partial := range partials {
		tokens := bytes.Split(partial, templateEndBytes)
		if len(tokens) == 1 {
			continue
		}

		// Resolve functions and variables.
		// If no matched variable is found, try to find an global var.
		for _, token := range tokens {
			if len(token) == 0 {
				continue
			}

			fnc, args := unwrapFunc(token)
			if len(fnc) == 0 {
				// No function found, try to lookup and replace variable.
				ret = c.resolveVariable(rArgs.fileName, ret, token, rArgs.additionalVars)
				continue
			}

			// Try to resolve function.
			ret, err = c.resolveFunction(rArgs.fileName, ret, token, rArgs.additionalVars, fnc, args)
			if err != nil {
				return
			}
		}
	}

	if len(bytes.Split(ret, templateStartBytes)) > 1 && !bytes.Equal(ret, rArgs.line) {
		ret, err = c.resolve(resolveArgs{
			fileName:       rArgs.fileName,
			line:           ret,
			additionalVars: rArgs.additionalVars,
		})
		if err != nil {
			return
		}
	}

	return
}

func (c *Core) resolveVariable(fileName string, line, token []byte, additionalVars []common.Variable) (ret []byte) {
	// No function found, try to lookup and replace variable.
	ret = line
	tokenString := string(token)
	v := c.varLookup(fileName, tokenString)
	if v.Value() != "" {
		return replaceVar(ret, token, []byte(v.Value()))
	}

	for _, av := range additionalVars {
		if av.Name() == tokenString {
			v = av
			break
		}
	}
	if v.Name() == tokenString {
		ret = replaceVar(ret, token, []byte(v.Value()))
	}
	return
}

func (c *Core) resolveFunction(fileName string, line, token []byte, additionalVars []common.Variable, fnc parserFunc, args [][]byte) (ret []byte, err error) {
	ret = line
	fncName := fnc.string()

	// Check function's args for variables.
	varsFromArgs := make([]common.Variable, 0)
	for j := range args {
		v := c.varLookup(fileName, string(args[j]))
		if v.Name() == "" {
			// For some functions, numbers are also used. Add them.
			val := string(args[j])
			v = variable{name: val, value: val}
		}
		varsFromArgs = append(varsFromArgs, v)
	}
	if len(varsFromArgs) == 0 {
		return
	}

	var remappedArgs [][]byte
	remappedArgs, err = remapArgsWithVariables(fncName, varsFromArgs, additionalVars)
	if err != nil {
		return
	}

	var mod []byte
	mod, err = c.executeFunction(fnc, fileName, remappedArgs, additionalVars)
	if err != nil {
		return
	}
	ret = replaceVar(ret, token, mod)
	return
}

// unwrapFunc gets the function's name and its args from the given byte slice.
func unwrapFunc(b []byte) (fncName parserFunc, args [][]byte) {
	args = make([][]byte, 0)
	fnc := bytes.Split(bytes.TrimSpace(b), []byte("("))
	if len(fnc) == 1 {
		return
	}

	fncName = bytes.TrimSpace(fnc[0])
	fnc[1] = bytes.Split(fnc[1], []byte(")"))[0]
	args = bytes.Split(fnc[1], []byte(","))
	for j := range args {
		args[j] = bytes.TrimSpace(args[j])
	}
	return
}

func unwrapVar(token []byte) (t []byte) {
	tokens := bytes.Split(token, templateStartBytes)
	if len(tokens) == 1 {
		return token
	}
	match := bytes.SplitN(tokens[1], templateEndBytes, 2)
	if len(match) == 2 {
		// Token contains variable name, return it.
		return match[0]
	}
	return token
}
