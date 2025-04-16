package foreach

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"sync"

	"github.com/xiroxasx/fastplate/internal/common"
)

// TokenResolver is an interface which is responsible for:
// looking up, resolving and replacing variable and function tokens with their corresponding value.
type TokenResolver interface {
	Resolve(fileName string, l []byte, vars ...common.Variable) (ret []byte, err error)
	VarLookupRecursive(fileName, name string, untilForeachIdx int) (_ []common.Variable)
}

type Buffer struct {
	preEvalIdx    int
	evalStateIdx  int
	states        []state
	lineEnding    []byte
	linesBuffered int
	stateMx       *sync.Mutex
}

type state struct {
	fileName         string
	args             []Arg
	jumps            []jump
	closed           bool
	lines            [][]byte
	previousStateIdx int
}

type Arg []byte

type jump struct {
	lineNum  int
	stateIdx int
}

func NewForeachBuffer(lineEnding []byte) Buffer {
	return Buffer{
		preEvalIdx:   -1,
		states:       make([]state, 0),
		lineEnding:   lineEnding,
		evalStateIdx: -1,
		stateMx:      &sync.Mutex{},
	}
}

func (b *Buffer) AppendState(fileName string, args []Arg) {
	b.stateMx.Lock()
	defer b.stateMx.Unlock()

	// Check if we need to set the state jump of the previous state.
	var (
		idx      int
		stateLen = len(b.states)
	)
	if stateLen > 0 {
		idx = stateLen
		state := &b.states[b.preEvalIdx]
		state.jumps = append(state.jumps, jump{
			lineNum:  b.linesBuffered,
			stateIdx: stateLen,
		})
	}

	b.preEvalIdx = idx
	b.states = append(b.states, state{
		fileName: fileName,
		args:     args,
		jumps:    make([]jump, 0),
		lines:    make([][]byte, 0),
	})
}

func (b *Buffer) IsActive() bool {
	return len(b.states) > 0 && !b.states[0].closed
}

func (b *Buffer) StateIndex() int {
	b.stateMx.Lock()
	defer b.stateMx.Unlock()

	return b.evalStateIdx
}

func (b *Buffer) MoveToPreviousState() (idx int) {
	b.stateMx.Lock()
	defer b.stateMx.Unlock()

	// Close the current state and find the next evaluation index to move the curor to.
	b.states[b.preEvalIdx].closed = true

	// Find next last opened state to attach to the corresponding buffer on next write.
	for i := len(b.states) - 1; i > 0; i-- {
		if !b.states[i].closed {
			idx = i
			break
		}
	}
	b.states[b.preEvalIdx].previousStateIdx = idx
	b.preEvalIdx = idx
	return
}

func (b *Buffer) WriteLineToBuffer(v []byte) {
	// If the last state is closed, we need to write to the latest state.
	idx := b.preEvalIdx
	v = append(v, b.lineEnding...)
	b.states[idx].lines = append(b.states[idx].lines, v)

	b.linesBuffered++
}

func (b *Buffer) Evaluate(lineNum int, dst io.Writer, tr TokenResolver) (err error) {
	if len(b.states) == 0 {
		return errors.New("no states")
	}
	if !b.states[0].closed {
		return errors.New("unclosed states")
	}

	defer func() {
		b.stateMx.Lock()
		b.evalStateIdx = -1
		b.stateMx.Unlock()
	}()

	return b.eval(0, lineNum, tr, dst)
}

func (b *Buffer) eval(stateIdx int, lineNum int, tr TokenResolver, dst io.Writer) (err error) {
	resetState := func(idx int) {
		b.stateMx.Lock()
		b.evalStateIdx = idx
		b.stateMx.Unlock()
	}

	vars, rangeNum := b.evaluationVars(b.states[stateIdx].fileName, tr, stateIdx)
	if rangeNum > -1 {
		for i := 0; i < rangeNum; i++ {
			err = b.evalLines(stateIdx, lineNum, tr, dst, []common.Variable{
				common.NewVar("index", strconv.Itoa(i)),
			}...)
			if err != nil {
				return
			}

			// Jump back to the provided state for the next loop.
			resetState(stateIdx)
		}
		return
	}

	for i, v := range vars {
		err = b.evalLines(stateIdx, lineNum, tr, dst, []common.Variable{
			common.NewVar("index", strconv.Itoa(i)),
			common.NewVar("value", v.Value()),
		}...)
		if err != nil {
			return
		}

		// Jump back to the provided state for the next loop.
		resetState(stateIdx)
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

			err = b.eval(j.stateIdx, j.lineNum, tr, dst)
			if err != nil {
				return
			}
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

func (b *Buffer) evaluationVars(fileName string, tr TokenResolver, stateIdx int) (vs []common.Variable, rangeNum int) {
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

func (b *Buffer) ReverseLoopOrder(stateIdx int) (idxs []int) {
	idx := stateIdx
	for idx > 0 {
		idx = b.states[idx].previousStateIdx
		idxs = append(idxs, idx)
	}

	return
}
