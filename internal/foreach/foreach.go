package foreach

import (
	"sync"

	"github.com/xiroxasx/yatt/internal/common"
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

func (b *Buffer) IsActive() bool {
	return len(b.states) > 0 && !b.states[0].closed
}

func (b *Buffer) StateIndex() int {
	b.stateMx.Lock()
	defer b.stateMx.Unlock()

	return b.evalStateIdx
}

func (b *Buffer) ReverseLoopOrder(stateIdx int) (idxs []int) {
	idx := stateIdx
	for idx > 0 {
		idx = b.states[idx].previousStateIdx
		idxs = append(idxs, idx)
	}

	return
}
