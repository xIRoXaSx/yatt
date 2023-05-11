package interpreter

import (
	"bytes"
	"strings"
	"sync"

	"github.com/xiroxasx/fastplate/internal/common"
)

var (
	templateStart = []byte("{{")
	templateEnd   = []byte("}}")
)

type variable struct {
	name  string
	value string
}

func (v variable) Name() string {
	return v.name
}

func (v variable) Value() string {
	return v.value
}

// variable parses a variable name and value and returns it.
func (i *Interpreter) variable(args [][]byte) (v common.Var) {
	i.state.Lock()
	defer i.state.Unlock()

	newVar := variable{name: string(args[0])}
	val := bytes.SplitN(args[0], []byte{'='}, 2)
	if len(val) == 1 {
		// Syntax: x = y, x =y
		val = bytes.Split(args[1], []byte{'='})
		newVar.value = string(val[1])
	} else {
		// Syntax: x=y, x= y
		newVar = variable{name: string(val[0]), value: string(val[1])}
	}
	if len(args) > 2 {
		// Skip equal sign.
		ind := 1
		if string(args[ind]) == "=" {
			// Any case except syntax x= y
			if len(args) > ind+1 {
				// Skip next space.
				ind++
			}
		}
		subArgs := args[ind:]
		str := make([]string, len(subArgs))
		for j, s := range subArgs {
			str[j] = string(s)
		}
		newVar.value += strings.Join(str, " ")
	}
	return newVar
}

// setScopedVar parses and sets a scoped variable from the given args.
func (i *Interpreter) setScopedVar(scope string, args [][]byte) {
	i.state.scopedRegistry.Lock()
	defer i.state.scopedRegistry.Unlock()

	for j, sv := range i.state.scopedRegistry.scopedVars[scope] {
		if sv.Name() == string(args[0]) {
			i.state.scopedRegistry.scopedVars[scope][j] = i.variable(args)
			break
		}
	}
	i.state.scopedRegistry.scopedVars[scope] = append(i.state.scopedRegistry.scopedVars[scope], i.variable(args))
}

// setUnscopedVar parses and sets an unscoped variable from the given args.
func (i *Interpreter) setUnscopedVar(varFile string, v [][]byte) {
	for j, uv := range i.state.unscopedVars {
		if string(v[0]) == uv.Name() {
			// Update existing variable.
			old := i.state.unscopedVars[j]
			i.state.unscopedVars[j] = common.NewVar(old.Name(), i.variable(v).Value())
			return
		}
	}
	i.state.unscopedVars = append(i.state.unscopedVars, i.variable(v))
	lowerVarFile := strings.ToLower(varFile)
	idx := i.state.unscopedVarIndexes[lowerVarFile]
	if idx.mx == nil {
		idx.mx = &sync.Mutex{}
		idx.start = len(i.state.unscopedVars) - 1
	}
	idx.mx.Lock()
	defer idx.mx.Unlock()

	idx.len++
	i.state.unscopedVarIndexes[lowerVarFile] = idx
}

// resolve resolves an import variable to its corresponding value.
// If the variable could not be found, the placeholders will not get replaced!
func (i *Interpreter) resolve(fileName string, line []byte, additionalVars []common.Var) (ret []byte, err error) {
	ret = line
	begin := bytes.Split(line, templateStart)
	if len(begin) == 1 {
		return
	}

	replaceVar := func(varName, replacement []byte) []byte {
		matched := bytes.Join([][]byte{templateStart, varName, templateEnd}, []byte{})
		return bytes.ReplaceAll(ret, matched, replacement)
	}

	for _, b := range begin {
		match := bytes.Split(b, templateEnd)
		if len(match) == 1 {
			continue
		}

		// Resolve scoped / local variables first.
		// If no matched variable is found, try to find an unscoped / global var.
		for _, m := range match {
			if len(m) == 0 {
				continue
			}

			fncName, vars := i.unpackFuncName(m)
			if len(fncName) == 0 {
				// No function found, try to lookup and replace variable.
				v := i.state.varLookup(fileName, string(m))
				if v.Value() == "" {
					continue
				}

				ret = replaceVar(m, []byte(v.Value()))
				continue
			}

			lookupVars := make([]common.Var, 0)
			for j := range vars {
				v := i.state.varLookup(fileName, string(vars[j]))
				if v.Name() == "" {
					// For some functions, numbers are also used. Add them.
					val := string(vars[j])
					v = variable{name: val, value: val}
				}
				lookupVars = append(lookupVars, v)
			}
			if len(lookupVars) == 0 {
				continue
			}

			fncNameStr := string(fncName)
			values := make([][]byte, len(lookupVars))
		lv:
			for j := range lookupVars {
				for _, av := range additionalVars {
					// Overwrite variable value if the names match.
					// This may be the case for "foreach"-variables.
					if lookupVars[j].Name() == av.Name() {
						values[j] = []byte(av.Value())
						continue lv
					}
				}
				if fncNameStr == "var" {
					values[j] = []byte(lookupVars[j].Name())
				} else {
					values[j] = []byte(lookupVars[j].Value())
				}
			}
			var mod []byte
			mod, err = i.executeFunction(fncNameStr, values, fileName, additionalVars)
			if err != nil {
				return
			}
			ret = replaceVar(m, mod)
		}
	}

	if len(bytes.Split(ret, templateStart)) > 1 && !bytes.Equal(ret, line) {
		ret, err = i.resolve(fileName, ret, additionalVars)
		if err != nil {
			return
		}
	}

	for _, v := range additionalVars {
		ret = replaceVar([]byte(v.Name()), []byte(v.Value()))
	}
	return
}

func (i *Interpreter) unpackFuncName(b []byte) (fncName []byte, args [][]byte) {
	args = make([][]byte, 0)
	fnc := bytes.Split(bytes.TrimSpace(b), []byte("("))
	if len(fnc) > 1 {
		fncName = bytes.TrimSpace(fnc[0])
		fnc[1] = bytes.Trim(fnc[1], ")")

		args = bytes.Split(fnc[1], []byte(","))
		for j := range args {
			args[j] = bytes.TrimSpace(args[j])
		}
	}
	return
}
