package interpreter

func (i *Interpreter) ignoreStart(pd *preprocessorDirective) (err error) {
	i.state.ignoreIndex[pd.fileName] = ignoreStateOpen
	return
}

func (i *Interpreter) ignoreEnd(pd *preprocessorDirective) (err error) {
	i.state.ignoreIndex[pd.fileName] = ignoreStateClose
	return
}
