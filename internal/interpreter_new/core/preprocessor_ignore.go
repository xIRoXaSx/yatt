package core

type ignoreState uint8

const (
	ignoreStateClose ignoreState = iota
	ignoreStateOpen

	variableRegistryGlobalRegister = "global"
)

func (c *Core) ignoreStart(pd *PreprocessorDirective) (err error) {
	c.ignoreIndex[pd.fileName] = ignoreStateOpen
	return
}

func (c *Core) ignoreEnd(pd *PreprocessorDirective) (err error) {
	c.ignoreIndex[pd.fileName] = ignoreStateClose
	return
}
