package importer

import (
	"bytes"
	"fmt"
	"io"
)

const (
	foreachValue        = "value"
	foreachIndex        = "index"
	foreachUnscopedVars = "UNSCOPED_VARS"
)

func (i *Interpreter) setForeachVar(file string, name string) {
	i.state.Lock()
	defer i.state.Unlock()

	// We only need the name, since the value is stored inside the scopedVars / globalVars.
	fe := i.state.foreach[file]
	fe.variables = append(fe.variables, variable{name: name})
	i.state.foreach[file] = fe
}

func (i *Interpreter) appendLine(file string, l []byte) {
	i.state.Lock()
	defer i.state.Unlock()

	fe := i.state.foreach[file]
	fe.lines = append(fe.lines, l)
	i.state.foreach[file] = fe
}

func (i *Interpreter) evaluateForeach(file string, out io.Writer) (err error) {
	resolve := func(idx int, v variable, file string, out io.Writer) {
		for _, l := range i.state.foreach[file].lines {
			var mod []byte
			mod, err = i.resolveForeach(idx, v, file, l)
			if err != nil {
				return
			}
			_, err = out.Write(append(mod, i.lineEnding...))
			if err != nil {
				return
			}
		}
	}

	for j, v := range i.state.foreach[file].variables {
		if v.name == foreachUnscopedVars {
			for idx, unscopedVar := range i.state.unscopedVars {
				resolve(idx, unscopedVar, file, out)
			}
			continue
		}
		resolve(j, variable{}, file, out)
	}
	return
}

// resolveForeach resolves an import variable to its corresponding value.
// If the variable could not be found, the placeholders will not get replaced!
func (i *Interpreter) resolveForeach(index int, v variable, file string, line []byte) (ret []byte, err error) {
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

		// Resolve foreach variables.
		for _, m := range match {
			var lookupVar variable
			fncName, vars := i.unpackFuncName(m)

			lookupVars := make([]variable, 0)
			if v != (variable{}) {
				lookupVars = append(lookupVars, v)
			}
			for _, varName := range vars {
				switch string(varName) {
				case foreachValue:
					if v == (variable{}) {
						lookupVars = append(lookupVars, i.state.varLookup(file, i.state.foreach[file].variables[index].name))
					}
				case foreachIndex:
					lookupVars = append(lookupVars, variable{value: fmt.Sprint(index)})
				default:
					lookupVars = append(lookupVars, variable{value: string(varName)})
				}
			}

			if len(fncName) > 0 {
				values := make([][]byte, len(lookupVars))
				for j := range lookupVars {
					values[j] = []byte(lookupVars[j].value)
				}

				var mod []byte
				mod, err = i.executeFunction(string(fncName), values)
				if err != nil {
					return
				}
				lookupVar.value = string(mod)
			}

			for _, v := range lookupVars {
				matched := bytes.Join([][]byte{templateStart, m, templateEnd}, []byte{})
				ret = bytes.ReplaceAll(ret, matched, []byte(v.value))
			}
		}
	}
	return
}
