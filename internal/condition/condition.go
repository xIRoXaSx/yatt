package condition

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
	frames  []frame
	history map[int]frame
	nextID  int
	stateMx *sync.Mutex
}

type frame struct {
	id            int
	parentID      int
	fileName      string
	lineNum       int
	parentActive  bool
	branchMatched bool
	active        bool
	elseSeen      bool
}

type Arg []byte

func NewConditionBuffer() Buffer {
	return Buffer{
		frames:  make([]frame, 0),
		history: make(map[int]frame),
		nextID:  -1,
		stateMx: &sync.Mutex{},
	}
}

func (b *Buffer) IsActive() bool {
	b.stateMx.Lock()
	defer b.stateMx.Unlock()

	return b.isActiveLocked()
}

func (b *Buffer) HasOpenFrames() bool {
	b.stateMx.Lock()
	defer b.stateMx.Unlock()

	return len(b.frames) > 0
}

func (b *Buffer) HasOpenFramesForFile(fileName string) bool {
	b.stateMx.Lock()
	defer b.stateMx.Unlock()

	for _, f := range b.frames {
		if f.fileName == fileName {
			return true
		}
	}
	return false
}

func (b *Buffer) CanEvaluateElseIf() (bool, error) {
	b.stateMx.Lock()
	defer b.stateMx.Unlock()

	if len(b.frames) == 0 {
		return false, ErrNoOpenCondition
	}

	f := b.frames[len(b.frames)-1]
	if f.elseSeen {
		return false, ErrElseIfAfterElse
	}
	return f.parentActive && !f.branchMatched, nil
}

func (b *Buffer) StateIndex() int {
	b.stateMx.Lock()
	defer b.stateMx.Unlock()

	if len(b.frames) == 0 {
		return -1
	}
	return b.frames[len(b.frames)-1].id
}

func (b *Buffer) ReverseLoopOrder(stateIdx int) (idxs []int) {
	b.stateMx.Lock()
	defer b.stateMx.Unlock()

	f, ok := b.history[stateIdx]
	for ok && f.parentID > -1 {
		idxs = append(idxs, f.parentID)
		f, ok = b.history[f.parentID]
	}

	return
}

func (b *Buffer) PushIf(fileName string, lineNum int, eval bool) {
	b.stateMx.Lock()
	defer b.stateMx.Unlock()

	parentActive := b.isActiveLocked()
	parentID := -1
	if len(b.frames) > 0 {
		parentID = b.frames[len(b.frames)-1].id
	}

	b.nextID++
	f := frame{
		id:            b.nextID,
		parentID:      parentID,
		fileName:      fileName,
		lineNum:       lineNum,
		parentActive:  parentActive,
		branchMatched: eval,
		active:        parentActive && eval,
	}
	b.frames = append(b.frames, f)
	b.history[f.id] = f
}

func (b *Buffer) ElseIf(eval bool) error {
	b.stateMx.Lock()
	defer b.stateMx.Unlock()

	if len(b.frames) == 0 {
		return ErrNoOpenCondition
	}

	f := &b.frames[len(b.frames)-1]
	if f.elseSeen {
		return ErrElseIfAfterElse
	}
	if f.branchMatched {
		f.active = false
		b.history[f.id] = *f
		return nil
	}

	f.active = f.parentActive && eval
	f.branchMatched = eval
	b.history[f.id] = *f
	return nil
}

func (b *Buffer) Else() error {
	b.stateMx.Lock()
	defer b.stateMx.Unlock()

	if len(b.frames) == 0 {
		return ErrNoOpenCondition
	}

	f := &b.frames[len(b.frames)-1]
	if f.elseSeen {
		return ErrElseAlreadySeen
	}

	f.active = f.parentActive && !f.branchMatched
	f.branchMatched = true
	f.elseSeen = true
	b.history[f.id] = *f
	return nil
}

func (b *Buffer) End() error {
	b.stateMx.Lock()
	defer b.stateMx.Unlock()

	if len(b.frames) == 0 {
		return ErrNoOpenCondition
	}

	b.frames = b.frames[:len(b.frames)-1]
	return nil
}

func (b *Buffer) isActiveLocked() bool {
	return len(b.frames) == 0 || b.frames[len(b.frames)-1].active
}
