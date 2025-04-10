package core

import (
	"errors"

	"github.com/xiroxasx/fastplate/internal/interpreter_new/foreach"
)

func (c *Core) foreachStart(pd *PreprocessorDirective) (err error) {
	if len(pd.args) < 1 {
		return errors.New("at least 1 arg expected")
	}

	febArgs := make([]foreach.Arg, len(pd.args))
	for i, arg := range pd.args {
		febArgs[i] = foreach.Arg(unwrapVar(arg))
	}
	c.feb.AppendState(pd.fileName, febArgs)
	return
}

func (c *Core) foreachEnd(pd *PreprocessorDirective) (err error) {
	c.feb.MoveToPreviousState()

	if c.feb.IsActive() {
		return
	}

	const startLine = 0
	return c.feb.Evaluate(startLine, pd.buf, c)
}
