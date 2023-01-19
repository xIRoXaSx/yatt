package importer

import (
	"bytes"
	"strings"
)

var (
	templateStart = []byte("{{")
	templateEnd   = []byte("}}")
)

// variable parses variable names and values and returns it.
func (i *Importer) variable(args []string) (v variable) {
	i.state.Lock()
	defer i.state.Unlock()

	v = variable{name: args[0]}
	val := strings.SplitN(args[0], "=", 2)
	if len(val) == 1 {
		// Syntax: x = y, x =y
		val = strings.Split(args[1], "=")
		v.value = val[1]
	} else {
		// Syntax: x=y, x= y
		v.name = val[0]
		v.value = val[1]
	}
	if len(args) > 2 {
		// Skip equal sign.
		ind := 1
		if args[ind] == "=" {
			// Any case except syntax x= y
			if len(args) > ind+1 {
				// Skip next space.
				ind++
			}
		}
		v.value += strings.Join(args[ind:], " ")
	}
	return
}

// setScopedVar parses and sets a scoped variable from the given args.
func (i *Importer) setScopedVar(scope string, args []string) {
	i.state.scopedVars[scope] = append(i.state.scopedVars[scope], i.variable(args))
}

// setUnScopedVar parses and sets an unscoped variable from the given args.
func (i *Importer) setUnScopedVar(args []string) {
	i.state.unscopedVars = append(i.state.unscopedVars, i.variable(args))
}

// resolve resolves an import variable to its corresponding value.
// If the variable could not be found, the placeholders will not get replaced!
func (i *Importer) resolve(fileName string, line []byte) (ret []byte) {
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

		for _, m := range match {
			varName := string(m)
			v := i.state.lookupScoped(fileName, varName)
			if v.name == "" {
				v = i.state.lookupUnScoped(varName)
				if v.name == "" {
					continue
				}
			}
			matched := bytes.Join([][]byte{templateStart, []byte(v.name), templateEnd}, []byte{})
			ret = bytes.ReplaceAll(ret, matched, []byte(v.value))
		}
	}
	return
}
