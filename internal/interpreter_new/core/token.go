package core

import (
	"bytes"
	"io"
	"path/filepath"

	"github.com/xiroxasx/fastplate/internal/common"
)

func (c *Core) searchTokensAndExecute(fileName string, line, currentLineIndent, parentLineIndent []byte, buf io.Writer, lineNum int) (err error) {
	lineDisplayNum := lineNum + 1
	lineNoIndent := line[len(currentLineIndent):]
	prefix := c.matchedPrefixToken(lineNoIndent)
	if len(prefix) > 0 {
		// Trim the prefix and check against internal commands.
		statement := trimLine(lineNoIndent, prefix)
		split := bytes.Split(statement, []byte{' '})
		if len(split) == 0 {
			return
		}

		pd := newPreprocessorDirective(
			string(split[0]),
			filepath.Clean(fileName),
			split[1:],
			append(currentLineIndent, parentLineIndent...),
		)
		err = c.Preprocess(pd, lineDisplayNum, func(pd *PreprocessorDirective) error {
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

	default:
		// No prefix found, try to resolve variables and functions if there are any.
		var ret []byte
		ret, err = c.resolve(fileName, line, nil)
		if err != nil {
			return
		}
		indents := append(parentLineIndent, currentLineIndent...)
		ret = append(indents, ret...)
		_, err = buf.Write(append(ret, lineEnding...))
		if err != nil {
			return
		}
	}

	return
}

func (c *Core) resolve(fileName string, line []byte, additionalVars []common.Variable) (ret []byte, err error) {
	templateStart := templateStartBytes
	templateEnd := templateEndBytes
	ret = line
	tokens := bytes.Split(line, templateStart)
	if len(tokens) == 1 {
		// Nothing needs to be resolved.
		return
	}

	for _, token := range tokens {
		match := bytes.Split(token, templateEnd)
		if len(match) == 1 {
			continue
		}

		// Resolve functions and variables.
		// If no matched variable is found, try to find an global var.
		for _, m := range match {
			if len(m) == 0 {
				continue
			}

			fncName, args := unwrapFunc(m)
			if len(fncName) == 0 {
				// No function found, try to lookup and replace variable.
				v := c.varLookup(fileName, string(m))
				if v.Value() == "" {
					continue
				}

				ret = replaceVar(ret, m, []byte(v.Value()))
				continue
			}

			// Check function's args for variables.
			// TODO: IMPROVE.
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
				continue
			}

			var remappedArgs [][]byte
			fncNameStr := fncName.string()
			remappedArgs, err = remapArgsWithVariables(fncNameStr, varsFromArgs, additionalVars)
			if err != nil {
				return
			}

			var mod []byte
			mod, err = c.executeFunction(fncName, fileName, remappedArgs, additionalVars)
			if err != nil {
				return
			}
			ret = replaceVar(ret, m, mod)
		}
	}

	if len(bytes.Split(ret, templateStart)) > 1 && !bytes.Equal(ret, line) {
		ret, err = c.resolve(fileName, ret, additionalVars)
		if err != nil {
			return
		}
	}

	for _, v := range additionalVars {
		ret = replaceVar(line, []byte(v.Name()), []byte(v.Value()))
	}
	return
}

// unwrapFunc gets the function's name and its args from the given byte slice.
func unwrapFunc(b []byte) (fncName parserFunc, args [][]byte) {
	args = make([][]byte, 0)
	fnc := bytes.SplitN(bytes.TrimSpace(b), []byte("("), 2)
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
