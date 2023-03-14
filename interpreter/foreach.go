package importer

import (
	"bytes"
	"fmt"
	"io"
)

const (
	foreachValue = "value"
	foreachIndex = "index"
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
	for j := range i.state.foreach[file].variables {
		for _, l := range i.state.foreach[file].lines {
			var mod []byte
			mod, err = i.resolveForeach(j, file, l)
			if err != nil {
				return
			}
			_, err = out.Write(append(mod, i.lineEnding...))
			if err != nil {
				return
			}
		}
	}
	return
}

// resolveForeach resolves an import variable to its corresponding value.
// If the variable could not be found, the placeholders will not get replaced!
func (i *Interpreter) resolveForeach(index int, file string, line []byte) (ret []byte, err error) {
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
			var v variable
			varName, fncName := i.unpackFuncName(m)

			switch string(varName) {
			case foreachValue:
				v = i.state.varLookup(file, i.state.foreach[file].variables[index].name)
				if len(fncName) > 0 {
					var mod []byte
					mod, err = i.executeFunction(string(fncName), []byte(v.value))
					if err != nil {
						return
					}
					v.value = string(mod)
					matched := bytes.Join([][]byte{templateStart, fncName, []byte("("), []byte(v.name), []byte(")"), templateEnd}, []byte{})
					ret = bytes.ReplaceAll(ret, matched, []byte(v.value))
				}
			case foreachIndex:
				v = variable{value: fmt.Sprint(index)}
			default:
				continue
			}
			group := bytes.Join([][]byte{templateStart, m, templateEnd}, nil)
			ret = bytes.ReplaceAll(ret, group, []byte(v.value))
		}
	}
	return
}
