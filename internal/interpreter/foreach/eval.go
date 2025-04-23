package foreach

import (
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/xiroxasx/fastplate/internal/common"
)

func (b *Buffer) Evaluate(lineNum int, dst io.Writer, tr TokenResolver) (err error) {
	if len(b.states) == 0 {
		return errors.New("no states")
	}
	if !b.states[0].closed {
		return errors.New("unclosed states")
	}

	b.evalStateIdx = 0
	defer func() {
		b.stateMx.Lock()
		b.evalStateIdx = -1
		b.stateMx.Unlock()
	}()

	return b.eval(0, lineNum, tr, dst)
}

func (b *Buffer) eval(stateIdx int, lineNum int, tr TokenResolver, dst io.Writer) (err error) {
	vars, rangeNum := b.loopEnumerator(b.states[stateIdx].fileName, tr, stateIdx)
	if rangeNum > -1 {
		for i := 0; i < rangeNum; i++ {
			// Evaluate each state (may be nested) accordingly.
			err = b.evalLines(stateIdx, lineNum, tr, dst, []common.Variable{
				common.NewVar("index", strconv.Itoa(i)),
			}...)
			if err != nil {
				return
			}

			// Jump back to the first state for the next loop.
			b.moveEvalState(stateIdx)
		}
		return
	}

	for i, v := range vars {
		// Evaluate each state (may be nested) accordingly.
		err = b.evalLines(stateIdx, lineNum, tr, dst, []common.Variable{
			common.NewVar("index", strconv.Itoa(i)),
			common.NewVar("value", v.Value()),
		}...)
		if err != nil {
			return
		}

		// Jump back to the provided state for the next loop.
		b.moveEvalState(stateIdx)
	}
	return
}

func (b *Buffer) evalLines(stateIdx int, lineNum int, tr TokenResolver, dst io.Writer, vars ...common.Variable) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("foreach evaluation: %v (line %d)", err, lineNum)
		}
	}()

	state := b.states[stateIdx]
	jumps := state.jumps
	var bufLn int

line:
	for ln := lineNum; ln < b.linesBuffered; ln++ {
		// Check for nested foreach loops.
		for _, j := range jumps {
			if j.lineNum != ln {
				continue
			}

			// Prepare for the nested foreach loop.
			b.moveEvalState(j.stateIdx)

			err = b.eval(j.stateIdx, j.lineNum, tr, dst)
			if err != nil {
				return
			}

			// Revert state index.
			b.moveEvalState(stateIdx)
			continue line
		}

		if bufLn >= len(state.lines) {
			return
		}

		line := state.lines[bufLn]
		var resolved []byte
		resolved, err = tr.Resolve(state.fileName, line, append(vars, common.NewVar("line", strconv.Itoa(ln)))...)
		if err != nil {
			return
		}
		_, err = dst.Write(resolved)
		if err != nil {
			return
		}
		bufLn++
	}

	return
}

//
// Helper
//

func (b *Buffer) moveEvalState(stateIdx int) {
	b.stateMx.Lock()
	b.evalStateIdx = stateIdx
	b.stateMx.Unlock()
}

func (b *Buffer) loopEnumerator(fileName string, tr TokenResolver, stateIdx int) (vs []common.Variable, rangeNum int) {
	state := b.states[stateIdx]
	variables := make([]common.Variable, 0)
	argsLen := len(state.args)
	rangeNum = -1
	for _, arg := range state.args {
		argStr := string(arg)

		if argsLen == 1 {
			// Looks like the user wants to range over the amount specified in the arg.
			rangeNum = 0
			rng, err := strconv.Atoi(argStr)
			if err == nil {
				rangeNum = rng
			}
			return
		}

		vars := tr.VarLookupRecursive(fileName, argStr, stateIdx)
		if len(vars) == 1 && argsLen == 1 {
			// Looks like the user wants to range over the amount specified in a variable.
			rangeNum = 0
			rng, err := strconv.Atoi(argStr)
			if err == nil && argsLen == 1 {
				// User wants to range over the amount specified.
				rangeNum = rng
			}
			return
		}

		if len(vars) > 0 {
			variables = append(variables, vars...)
		}
	}

	return variables, -1
}
