package importer

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

const (
	foreachValue        = "value"
	foreachIndex        = "index"
	foreachName         = "name"
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

	var (
		iterator int
		id       int
	)
	vars := i.state.foreach[file].variables
	if len(vars) == 1 {
		iterator, err = strconv.Atoi(vars[0].name)
		if err != nil {
			// The given arg is not an integer, check if variable contains an integer value.
			iterator, err = strconv.Atoi(i.state.varLookup(file, vars[0].name).value)
			if err != nil && !strings.HasPrefix(vars[0].name, foreachUnscopedVars) {
				err = errors.New("foreach: single value provided but does not match integer value")
				return
			}
		}
	}

	for j, v := range i.state.foreach[file].variables {
		// Check if loop should run as for-loop (0 < n).
		if iterator > 0 {
			for it := 0; it < iterator; it++ {
				val := fmt.Sprint(it)
				resolve(it, variable{name: val, value: val}, file, out)
			}
			continue
		}

		// Check if loop should iterate over all unscoped vars.
		if v.name == foreachUnscopedVars {
			for idx, unscopedVar := range i.state.unscopedVars {
				resolve(idx, unscopedVar, file, out)
			}
			continue
		}

		// Check if loop should iterate over specific unscoped vars.
		varFile := strings.TrimPrefix(v.name, foreachUnscopedVars+"_")
		if varFile != v.name {
			idx := i.state.unscopedVarIndexes[strings.ToLower(varFile)]
			for _, unscopedVar := range i.state.unscopedVars[idx.start : idx.start+idx.len] {
				resolve(id, unscopedVar, file, out)
				id++
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
		{name: foreachName, value: v.name},
	}
	if index < len(i.state.foreach[file].variables) {
		feVars[0].value = i.state.varLookup(file, i.state.foreach[file].variables[index].name).value
	}
	if v != (variable{}) {
		feVars[0].value = v.value
	}

	ret, err = i.resolve(file, line, feVars)
	return
}
