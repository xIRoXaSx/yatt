package importer

import (
	"bytes"
	"strings"
)

var (
	templateStart = []byte("{{")
	templateEnd   = []byte("}}")
)

// variable parses a variable name and value and returns it.
func (i *Interpreter) variable(args [][]byte) (v variable) {
	i.state.Lock()
	defer i.state.Unlock()

	v = variable{name: string(args[0])}
	val := bytes.SplitN(args[0], []byte{'='}, 2)
	if len(val) == 1 {
		// Syntax: x = y, x =y
		val = bytes.Split(args[1], []byte{'='})
		v.value = string(val[1])
	} else {
		// Syntax: x=y, x= y
		v.name = string(val[0])
		v.value = string(val[1])
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
		v.value += strings.Join(str, " ")
	}
	return
}

// setScopedVar parses and sets a scoped variable from the given args.
func (i *Interpreter) setScopedVar(scope string, args [][]byte) {
	i.state.scopedVars[scope] = append(i.state.scopedVars[scope], i.variable(args))
}

// setUnscopedVar parses and sets an unscoped variable from the given args.
func (i *Interpreter) setUnscopedVar(v [][]byte) {
	for j, uv := range i.state.unscopedVars {
		if string(v[0]) == uv.name {
			// Update existing variable.
			i.state.unscopedVars[j].value = i.variable(v).value
			return
		}
	}
	i.state.unscopedVars = append(i.state.unscopedVars, i.variable(v))
}

// resolve resolves an import variable to its corresponding value.
// If the variable could not be found, the placeholders will not get replaced!
func (i *Interpreter) resolve(fileName string, line []byte, additionalVars []variable) (ret []byte, err error) {
	ret = line
	begin := bytes.Split(line, templateStart)
	if len(begin) == 1 {
		return
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
				v := i.state.varLookup(fileName, string(m))
				if v.value == "" {
					continue
				}

				matched := bytes.Join([][]byte{templateStart, m, templateEnd}, []byte{})
				ret = bytes.ReplaceAll(ret, matched, []byte(v.value))
				continue
			}

			lookupVars := make([]variable, 0)
			for j := range vars {
				v := i.state.varLookup(fileName, string(vars[j]))
				if v.name == "" {
					// For some functions, numbers are also used.
					val := string(vars[j])
					v = variable{name: val, value: val}
				}
				lookupVars = append(lookupVars, v)
			}
			if len(lookupVars) == 0 {
				continue
			}

			values := make([][]byte, len(lookupVars))
		lv:
			for j := range lookupVars {
				for _, av := range additionalVars {
					if lookupVars[j].name == av.name {
						values[j] = []byte(av.value)
						continue lv
					}
				}

				values[j] = []byte(lookupVars[j].value)
			}
			var mod []byte
			mod, err = i.executeFunction(string(fncName), values)
			if err != nil {
				return
			}
			matched := bytes.Join([][]byte{templateStart, m, templateEnd}, []byte{})
			ret = bytes.ReplaceAll(ret, matched, mod)
		}
	}

	for _, v := range additionalVars {
		matched := bytes.Join([][]byte{templateStart, []byte(v.name), templateEnd}, []byte{})
		ret = bytes.ReplaceAll(ret, matched, []byte(v.value))
	}
	return
}

func (i *Interpreter) unpackFuncName(b []byte) (fncName []byte, args [][]byte) {
	args = make([][]byte, 0)
	fnc := bytes.Split(b, []byte("("))
	if len(fnc) > 1 {
		fncName = fnc[0]
		fnc[1] = bytes.Trim(fnc[1], ")")

		args = bytes.Split(fnc[1], []byte(","))
		for j := range args {
			args[j] = bytes.TrimSpace(args[j])
		}
	}
	return
}
