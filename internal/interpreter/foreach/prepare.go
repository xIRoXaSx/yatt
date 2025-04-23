package foreach

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
