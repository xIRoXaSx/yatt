package importer

import (
	"errors"
	"fmt"
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

func (i *Interpreter) appendForeachLine(file string, l []byte) {
	i.state.Lock()
	defer i.state.Unlock()

	fe := i.state.foreach[file]
	fe.lines = append(fe.lines, l)
	i.state.foreach[file] = fe
}

func (i *Interpreter) evaluateForeach(file string) (err error) {
	resolve := func(varIdx, feIdx int, v variable, file string) {
		for _, l := range i.state.foreach[file].lines {
			var mod []byte
			mod, err = i.resolveForeach(varIdx, feIdx, v, file, l)
			if err != nil {
				return
			}
			_, err = i.state.buf.Write(append(mod, i.lineEnding...))
			if err != nil {
				return
			}
		}
	}

	vars := i.state.foreach[file].variables
	if len(vars) == 1 {
		var iterator int
		iterator, err = strconv.Atoi(vars[0].name)
		if err != nil {
			// The given arg is not an integer, check if variable contains an integer value.
			iterator, err = strconv.Atoi(i.state.varLookup(file, vars[0].name).value)
			if err != nil && !strings.HasPrefix(vars[0].name, foreachUnscopedVars) {
				err = errors.New("foreach: single value provided but does not match integer value")
				return
			}

			if iterator > 0 {
				// Loop should run as for-loop (0 < n).
				for it := 0; it < iterator; it++ {
					val := fmt.Sprint(it)
					resolve(it, it, variable{name: val, value: val}, file)
				}
				return
			}
		}
	}

	var id int
	for j, v := range i.state.foreach[file].variables {
		// Check if loop should iterate over all unscoped vars.
		if v.name == foreachUnscopedVars {
			for idx, unscopedVar := range i.state.unscopedVars {
				resolve(idx, id, unscopedVar, file)
				id++
			}
			continue
		}

		// Check if loop should iterate over specific unscoped var files.
		varFile := strings.TrimPrefix(v.name, foreachUnscopedVars+"_")
		if varFile != v.name {
			idx := i.state.unscopedVarIndexes[strings.ToLower(varFile)]
			for vid, unscopedVar := range i.state.unscopedVars[idx.start : idx.start+idx.len] {
				resolve(idx.start+vid, id, unscopedVar, file)
				id++
			}
			continue
		}

		resolve(j, id, variable{}, file)
		id++
	}
	return
}

// resolveForeach resolves an import variable to its corresponding value.
// If the variable could not be found, the placeholders will not get replaced!
func (i *Interpreter) resolveForeach(varIdx, feIdx int, v variable, file string, line []byte) (ret []byte, err error) {
	feVars := []variable{
		{name: foreachValue, value: ""},
		{name: foreachIndex, value: fmt.Sprint(feIdx)},
		{name: foreachName, value: v.name},
	}
	if varIdx < len(i.state.foreach[file].variables) {
		feVars[0].value = i.state.varLookup(file, i.state.foreach[file].variables[varIdx].name).value
	}
	if v != (variable{}) {
		feVars[0].value = v.value
	}

	ret, err = i.resolve(file, line, feVars)
	return
}
