package core

import (
	"errors"
	"strings"

	"github.com/xiroxasx/fastplate/internal/common"
	"github.com/xiroxasx/fastplate/internal/interpreter_new/foreach"
)

func (c *Core) foreachStart(pd *PreprocessorDirective) (err error) {
	if len(pd.args) < 1 {
		return errors.New("at least 1 arg expected")
	}

	febArgs := make([]foreach.Arg, len(pd.args))
	for i, arg := range pd.args {
		febArgs[i] = foreach.Arg(arg)
	}
	c.feb.AppendState(pd.fileName, febArgs, pd.indent)
	return
}

func (c *Core) foreachEnd(pd *PreprocessorDirective) (err error) {
	c.feb.MoveToPreviousState()

	if c.feb.IsActive() {
		return
	}

	const startLine = 0
	err = c.feb.Evaluate(startLine, c, unwrapVar, func(fileName string, l []byte, vars ...common.Variable) (ret []byte, err error) {
		return c.resolve(fileName, l, vars)
	}, c.varLookupRecursive, pd.buf)
	if err != nil {
		return
	}

	return
}

func (c *Core) varLookupRecursive(fileName, name string, vg foreach.VariableGetter, untilForeachIdx int) (_ []common.Variable) {
	v := c.varLookupForeach(fileName, name, untilForeachIdx)
	if v != nil {
		return []common.Variable{v}
	}

	// Foreach variable not found, try getting a local variable.
	vs := vg.VarLookupLocal(fileName, name)
	if vs.Name() != "" {
		return []common.Variable{vs}
	}

	// Try to find it against a global var file name.
	vArgs := strings.Split(name, variableGlobalKeyFile)
	if len(vArgs) > 1 {
		vs := vg.VarsLookupGlobalFile(vArgs[1])
		return vs
	}

	// Last resort, try global vars.
	if name == variableGlobalKey {
		return vg.VarsLookupGlobal()
	}
	gVars := vg.VarsLookupGlobal()
	for _, v := range gVars {
		if v.Name() == name {
			return []common.Variable{v}
		}
	}

	return
}

func (c *Core) varLookupForeach(fileName, name string, stateIdx int) (_ common.Variable) {
	if stateIdx > len(c.varRegistryForeach)-1 {
		return nil
	}

	for _, reg := range c.varRegistryForeach[:stateIdx] {
		v := varLookupRegistry(&reg, fileName, name)
		if v.Name() == "" {
			continue
		}

		return v
	}

	return
}
