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
			varName, fncName := i.unpackFuncName(m)

			switch string(varName) {
			case foreachValue:
				if v != (variable{}) {
					lookupVar = v
				} else {
					lookupVar = i.state.varLookup(file, i.state.foreach[file].variables[index].name)
				}
				if len(fncName) > 0 {
					var mod []byte
					mod, err = i.executeFunction(string(fncName), []byte(lookupVar.value))
					if err != nil {
						return
					}
					lookupVar.value = string(mod)
					matched := bytes.Join([][]byte{templateStart, fncName, []byte("("), []byte(lookupVar.name), []byte(")"), templateEnd}, []byte{})
					ret = bytes.ReplaceAll(ret, matched, []byte(lookupVar.value))
				}
			case foreachIndex:
				lookupVar = variable{value: fmt.Sprint(index)}
			default:
				continue
			}
			group := bytes.Join([][]byte{templateStart, m, templateEnd}, nil)
			ret = bytes.ReplaceAll(ret, group, []byte(lookupVar.value))
		}
	}
	return
}
