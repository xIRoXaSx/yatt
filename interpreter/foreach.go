package importer

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
)

const (
	foreachValue        = "value"
	foreachIndex        = "index"
	foreachName         = "name"
	foreachUnscopedVars = "UNSCOPED_VARS"
)

type foreach struct {
	c    cursor
	open int
	buf  queue
	mx   *sync.Mutex
}

type cursor struct {
	p int
	j int
}

type foreachBuffer struct {
	nextRef   []int
	startNext []int
	ln        int
	variables []variable
	lines     [][]byte
}

func (i *Interpreter) setForeachVar(file string, name string) (err error) {
	fe, err := i.state.foreachLoad(file)
	if err != nil {
		return
	}

	// We only need the name, since the value is stored inside the scopedVars / globalVars.
	fe.buf.v[fe.c.p].variables = append(fe.buf.v[fe.c.p].variables, variable{name: name})
	i.state.foreach.Store(file, fe)
	return
}

func (i *Interpreter) appendForeachLine(file string, l []byte) (err error) {
	fe, err := i.state.foreachLoad(file)
	if err != nil {
		return
	}

	buf := fe.buf.firstN(fe.c.p)
	buf.lines = append(buf.lines, l)
	i.state.foreach.Store(file, fe)
	return
}

func (i *Interpreter) evaluateForeach(fe foreach, file string) (err error) {
	newBufferedLoop := func(fe foreach, file string) {
		fe.buf.mv(1)
		err = i.evaluateForeach(fe, file)
		if err != nil {
			return
		}
		fe.buf.mv(-1)
	}

	loopLines := func(varIdx, feIdx int, v variable, buf foreachBuffer) (err error) {
		// Loops may be nested directly inside each other.
		// If this happens and no other lines have been given, the line's length is 0.
		if buf.lines == nil && len(buf.startNext) == 1 {
			newBufferedLoop(fe, file)
			return
		}
		modified := len(buf.lines[0]) == 0
		for lineNum, l := range buf.lines {
			// Only resolve foreach loop and write it to the buffer if line isn't empty.
			if len(l) > 0 {
				var replaced []byte
				replaced, err = i.resolveForeach(varIdx, feIdx, v, file, l)
				if err != nil {
					return
				}

				_, err = i.state.buf.Write(append(replaced, i.lineEnding...))
				if err != nil {
					return
				}
			}

			for _, next := range buf.startNext {
				if (!modified && lineNum+1 == next) || (modified && lineNum == next) {
					newBufferedLoop(fe, file)
					break
				}
			}
		}
		return
	}

	mvBuff := func(buf foreachBuffer) {
		if fe.buf.p > 0 && fe.buf.p > len(buf.nextRef) {
			fe.buf.mv(-1)
		}
	}

	buf := *fe.buf.load()
	if len(buf.variables) == 1 {
		var (
			iterator int
			var0     = buf.variables[0]
		)
		iterator, err = strconv.Atoi(var0.name)
		if err != nil {
			// The given arg is not an integer, check if variable holds an integer value.
			iterator, err = strconv.Atoi(i.state.varLookup(file, var0.name).value)
			if err != nil && !strings.HasPrefix(var0.name, foreachUnscopedVars) {
				err = errors.New("foreach: single value provided but does not match integer value")
				return
			}
		}
		// Loop should run as for-loop (0 < n).
		for it := 0; it < iterator; it++ {
			val := fmt.Sprint(it)
			err = loopLines(it, it, variable{name: val, value: val}, buf)
			if err != nil {
				return
			}
		}
		mvBuff(buf)
		return
	}

	id := -1
	for vIdx, v := range buf.variables {
		id++
		// Check if loop should iterate over all unscoped vars.
		if v.name == foreachUnscopedVars {
			for idx, unscopedVar := range i.state.unscopedVars {
				err = loopLines(idx, id, unscopedVar, buf)
				if err != nil {
					return
				}
			}
			continue
		}

		// Check if loop should iterate over specific unscoped var files.
		varFile := strings.TrimPrefix(v.name, foreachUnscopedVars+"_")
		if varFile != v.name {
			idx := i.state.unscopedVarIndexes[strings.ToLower(varFile)]
			for vId, unscopedVar := range i.state.unscopedVars[idx.start : idx.start+idx.len] {
				err = loopLines(idx.start+vId, id, unscopedVar, buf)
				if err != nil {
					return
				}
			}
			continue
		}

		err = loopLines(vIdx, id, v, buf)
		if err != nil {
			return
		}
	}
	mvBuff(buf)
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

	fe, err := i.state.foreachLoad(file)
	if err != nil {
		return
	}

	feState := fe.buf.load()
	if varIdx < len(feState.variables) {
		feVars[0].value = i.state.varLookup(fmt.Sprintf("%s_%d", file, feIdx), feState.variables[varIdx].name).value
	}
	if v != (variable{}) {
		feVars[0].value = v.value
	}
	ret, err = i.resolve(file, line, feVars)
	return
}
