package core

import (
	"errors"
	"fmt"

	"github.com/xiroxasx/yatt/internal/condition"
)

func (c *Core) conditionIf(pd *PreprocessorDirective) (err error) {
	if len(pd.args) < 1 {
		return errors.New("at least 1 arg expected")
	}

	condArgs := make([]condition.Arg, len(pd.args))
	for i, arg := range pd.args {
		condArgs[i] = arg
	}
	eval, err := c.cb.IsTrue(pd.fileName, condArgs, c)
	if err != nil {
		return fmt.Errorf("condition isTrue: %v", err)
	}
	c.cb.PushIf(pd.fileName, pd.lineNum, eval)
	return
}

func (c *Core) conditionElseIf(pd *PreprocessorDirective) (err error) {
	if len(pd.args) < 1 {
		return errors.New("at least 1 arg expected")
	}

	canEvaluate, err := c.cb.CanEvaluateElseIf()
	if err != nil {
		return err
	}
	if !canEvaluate {
		return c.cb.ElseIf(false)
	}

	condArgs := make([]condition.Arg, len(pd.args))
	for i, arg := range pd.args {
		condArgs[i] = arg
	}
	eval, err := c.cb.IsTrue(pd.fileName, condArgs, c)
	if err != nil {
		return fmt.Errorf("condition isTrue: %v", err)
	}
	return c.cb.ElseIf(eval)
}

func (c *Core) conditionElse(pd *PreprocessorDirective) (err error) {
	if len(pd.args) > 0 {
		return errors.New("no args expected")
	}
	return c.cb.Else()
}

func (c *Core) conditionEnd(pd *PreprocessorDirective) (err error) {
	if len(pd.args) > 0 {
		return errors.New("no args expected")
	}
	return c.cb.End()
}

func (c *Core) ensureNoOpenConditions(fileName string) error {
	if !c.cb.HasOpenFramesForFile(fileName) {
		return nil
	}
	return errors.New("unclosed condition")
}
