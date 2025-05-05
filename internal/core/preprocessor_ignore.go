package core

type ignoreState int

func (i ignoreState) isActive() bool {
	return i > 0
}

func (c *Core) ignoreStart(pd *PreprocessorDirective) error {
	c.ignoreIndex[pd.fileName]++
	return nil
}

func (c *Core) ignoreEnd(pd *PreprocessorDirective) error {
	c.ignoreIndex[pd.fileName]--
	return nil
}
