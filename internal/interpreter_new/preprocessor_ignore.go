package interpreter

type ignoreIndexes map[string]ignoreState

func (s *state) ignoreStart(pd *preprocessorDirective) (err error) {
	s.ignoreIndex[pd.fileName] = ignoreStateOpen
	return
}

func (s *state) ignoreEnd(pd *preprocessorDirective) (err error) {
	s.ignoreIndex[pd.fileName] = ignoreStateClose
	return
}
