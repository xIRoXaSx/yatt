package foreach

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"sync"

	"github.com/xiroxasx/fastplate/internal/common"
)

type VariableGetter interface {
	VarLookupForeach(register, name string, stateIndx int) common.Variable
	VarLookupLocal(register, name string) common.Variable
	VarsLookupGlobalFile(name string) []common.Variable
	VarsLookupGlobal() []common.Variable
}

type ResolverFunc func(fileName string, l []byte, vars ...common.Variable) (ret []byte, err error)
type VarLookupRecursiveFunc func(fileName, name string, vg VariableGetter, untilForeachIdx int) (_ []common.Variable)
type UnwrapperFunc func([]byte) []byte

type Buffer struct {
	stateEvalIdx  int
	states        []state
	lineEnding    []byte
	linesBuffered int
	stateMx       *sync.Mutex
}

type state struct {
	fileName string
	args     []Arg
	jumps    []jump
	closed   bool
	buf      *bytes.Buffer
}

type Arg []byte

type jump struct {
	lineNum  int
	stateIdx int
}

func NewForeachBuffer(lineEnding []byte) Buffer {
	return Buffer{
		stateEvalIdx: -1,
		states:       make([]state, 0),
		lineEnding:   lineEnding,
		stateMx:      &sync.Mutex{},
	}
}

func (f *Buffer) AppendState(fileName string, args []Arg) {
	f.stateMx.Lock()
	defer f.stateMx.Unlock()

	// Check if we need to set the state jump of the previous state.
	var (
		idx      int
		stateLen = len(f.states)
	)
	if stateLen > 0 {
		idx = stateLen
		state := &f.states[f.stateEvalIdx]
		state.jumps = append(state.jumps, jump{lineNum: f.linesBuffered, stateIdx: stateLen})
	}

	f.stateEvalIdx = idx
	f.states = append(f.states, state{
		fileName: fileName,
		args:     args,
		jumps:    make([]jump, 0),
		buf:      &bytes.Buffer{},
	})
}

func (f *Buffer) IsActive() bool {
	return len(f.states) > 0 && !f.states[0].closed
}

func (f *Buffer) MoveToPreviousState() {
	f.stateMx.Lock()
	defer f.stateMx.Unlock()

	// Close the current state and find the next evaluation index to move the curor to.
	f.states[f.stateEvalIdx].closed = true

	// Find next last opened state to attach to the corresponding buffer.
	var idx int
	for i := len(f.states) - 1; i > 0; i-- {
		if !f.states[i].closed {
			idx = i
			break
		}
	}
	f.stateEvalIdx = idx
}

func (f *Buffer) WriteLineToBuffer(v []byte) (err error) {
	// If the last state is closed, we need to write to the latest state.
	idx := f.stateEvalIdx
	v = append(v, f.lineEnding...)
	_, err = f.states[idx].buf.Write(v)
	if err != nil {
		return
	}

	f.linesBuffered++
	return
}

func (f *Buffer) Evaluate(
	lineNum int,
	dst io.Writer,
	vg VariableGetter,
	unwrapperFunc UnwrapperFunc,
	vlr VarLookupRecursiveFunc,
	resolverFunc ResolverFunc,
) (err error) {
	if len(f.states) == 0 {
		return errors.New("no states")
	}
	if !f.states[0].closed {
		return errors.New("unclosed states")
	}

	return f.eval(0, lineNum, vg, unwrapperFunc, resolverFunc, vlr, dst)
}

func (f *Buffer) eval(
	stateIdx int,
	lineNum int,
	vg VariableGetter,
	uw UnwrapperFunc,
	rf ResolverFunc,
	vlr VarLookupRecursiveFunc,
	dst io.Writer,
) (err error) {
	vars, rangeNum := f.evaluationVars(f.states[stateIdx].fileName, vg, uw, vlr, stateIdx)
	if rangeNum > -1 {
		for i := 0; i < rangeNum; i++ {
			err = f.evalLines(stateIdx, lineNum, vg, uw, rf, vlr, dst, []common.Variable{
				common.NewVar("index", strconv.Itoa(i)),
			}...)
			if err != nil {
				return
			}
		}
		return
	}

	for i, v := range vars {
		err = f.evalLines(stateIdx, lineNum, vg, uw, rf, vlr, dst, []common.Variable{
			common.NewVar("index", strconv.Itoa(i)),
			common.NewVar("value", v.Value()),
		}...)
		if err != nil {
			return
		}
	}
	return
}

func (b *Buffer) evalLines(
	stateIdx int,
	lineNum int,
	vg VariableGetter,
	uf UnwrapperFunc,
	rf ResolverFunc,
	vlr VarLookupRecursiveFunc,
	dst io.Writer,
	vars ...common.Variable,
) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("foreach evaluation: %v (line %d)", err, lineNum)
		}
	}()

	state := b.states[stateIdx]
	jumps := state.jumps
	buf := bytes.Split(state.buf.Bytes(), b.lineEnding)
	var bufLn int

line:
	for ln := lineNum; ln < b.linesBuffered; ln++ {
		// Check for nested foreach loops.
		for _, j := range jumps {
			if j.lineNum != ln {
				continue
			}

			err = b.eval(j.stateIdx, j.lineNum, vg, uf, rf, vlr, dst)
			if err != nil {
				return
			}
			continue line
		}

		if bufLn >= len(buf) {
			return
		}

		line := buf[bufLn]
		var resolved []byte
		resolved, err = rf(state.fileName, line, append(vars, common.NewVar("line", strconv.Itoa(ln)))...)
		if err != nil {
			return
		}
		_, err = dst.Write(append(resolved, b.lineEnding...))
		if err != nil {
			return
		}
		bufLn++
	}

	return
}

func (b *Buffer) evaluationVars(
	fileName string,
	vg VariableGetter,
	uf UnwrapperFunc,
	vlr VarLookupRecursiveFunc,
	stateIdx int,
) (vs []common.Variable, rangeNum int) {
	stack := b.states[stateIdx]
	variables := make([]common.Variable, 0)
	argsLen := len(stack.args)
	rangeNum = -1
	for _, arg := range stack.args {
		argStr := string(uf(arg))

		if argsLen == 1 {
			// Looks like the user wants to range over the amount specified in the arg.
			rangeNum = 0
			rng, err := strconv.Atoi(argStr)
			if err == nil {
				rangeNum = rng
			}
			return
		}

		vars := vlr(fileName, argStr, vg, stateIdx)
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
