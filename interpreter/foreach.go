package importer

import (
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
	feVars := []variable{
		{name: foreachValue, value: ""},
		{name: foreachIndex, value: fmt.Sprint(index)},
	}
	if index < len(i.state.foreach[file].variables) {
		feVars[0].value = i.state.varLookup(file, i.state.foreach[file].variables[index].name).value
	} else if v != (variable{}) {
		feVars[0].value = v.value
	}

	ret, err = i.resolve(file, line, feVars)
	return
}
