package core

import (
	"bytes"
	"os"
	"strconv"
	"strings"

	"github.com/xiroxasx/yatt/internal/common"
)

const (
	variableRegistryGlobalRegister = "global"
)

type variable struct {
	name  string
	value string
}

// Implements the [common.Variable] interface.
func (v variable) Name() string {
	return v.name
}

// Implements the [common.Variable] interface.
func (v variable) Value() string {
	return v.value
}

//
// Interface implementation.
//

// Implement foreach.TokenResolver interface.
func (c *Core) VarLookupRecursive(fileName, name string, untilForeachIdx int) (_ []common.Variable) {
	return c.varLookupRecursive(fileName, name, untilForeachIdx)
}

//
// Variable setter.
//

func (c *Core) InitGlobalVariablesByFiles(varFileNames ...string) {
	// Check if the global var files exist and read it into the memory.
	for _, vf := range varFileNames {
		cont, err := os.ReadFile(vf)
		if err != nil {
			c.l.Fatal().Err(err).Str("path", vf).Msg("unable to read variable file")
		}

		lines := bytes.Split(cont, lineEnding)
		for _, l := range lines {
			split := bytes.Split(c.cutPrefix(l), []byte{' '})
			if len(split) < 3 || string(split[0]) != directiveNameVariable {
				continue
			}

			// Skip the var declaration keyword.
			v := common.VarFromArg(bytes.Join(split[1:], []byte(" ")))
			c.setGlobalVarWithReg(vf, v)
		}
	}
}

func (c *Core) setForeachVar(register string, newVar common.Variable) {
	setRegistryVar(&c.varRegistryForeach, register, newVar)
}

func (c *Core) setGlobalVar(newVar common.Variable) {
	setRegistryVar(&c.varRegistryGlobal, variableRegistryGlobalRegister, newVar)
}

func (c *Core) setGlobalVarWithReg(register string, newVar common.Variable) {
	setRegistryVar(&c.varRegistryGlobal, register, newVar)
}

func (c *Core) setLocalVar(register string, newVar common.Variable) {
	setRegistryVar(&c.varRegistryLocal, register, newVar)
}

func setRegistryVar(reg *variableRegistry, register string, newVar common.Variable) {
	reg.Lock()
	defer reg.Unlock()

	for i, v := range reg.entries[register] {
		if newVar.Name() == v.Name() {
			// Update existing variable.
			reg.entries[register][i] = newVar
			return
		}
	}

	reg.entries[register] = append(reg.entries[register], newVar)
}

//
// Variable getter.
//

func (c *Core) varLookup(file, name string) (v common.Variable) {
	if c.feb.StateIndex() > -1 {
		v = c.varLookupForeach(c.feb.StateIndex(), name)
		if v != nil {
			return
		}
	}

	v = c.varLookupLocal(file, name)
	if v.Name() != "" {
		return
	}

	// TODO: unify with core.varsLookupGlobalFile()
	v = c.varLookupGlobalWithRegister(file, name)
	if v.Name() != "" {
		return
	}

	v = c.varLookupGlobal(name)
	return
}

func (c *Core) varLookupGlobal(name string) (v common.Variable) {
	for k := range c.varRegistryGlobal.entries {
		for _, v := range c.varRegistryGlobal.entries[k] {
			if v.Name() == name {
				return v
			}
		}
	}

	return variable{}
}

func (c *Core) varLookupGlobalWithRegister(register, name string) (v common.Variable) {
	return varLookupRegistry(&c.varRegistryGlobal, register, name)
}

func (c *Core) varLookupLocal(register, name string) (v common.Variable) {
	return varLookupRegistry(&c.varRegistryLocal, register, name)
}

// TODO: Partials (e.g.: ", value=") are used to lookup variables,
// filter them out in the caller function.
func (c *Core) varLookupForeach(stateIdx int, name string) (_ common.Variable) {
	// Allowed to traverse from parent loops but limit to the current state index.
	idxs := c.feb.ReverseLoopOrder(stateIdx)
	idxs = append([]int{stateIdx}, idxs...)
	for i := 0; i < len(idxs); i++ {
		idx := idxs[i]
		reg := strconv.Itoa(idx)
		vars := c.varRegistryForeach.entries[reg]
		for _, v := range vars {
			if v.Name() != name {
				continue
			}

			return v
		}
	}

	return
}

func (c *Core) varLookupRecursive(fileName, name string, foreachStateIdx int) (_ []common.Variable) {
	v := c.varLookupForeach(foreachStateIdx, name)
	if v != nil {
		return []common.Variable{v}
	}

	// Foreach variable not found, try getting a local variable.
	vs := c.varLookupLocal(fileName, name)
	if vs.Name() != "" {
		return []common.Variable{vs}
	}

	// Try to find it against a global var file name.
	vArgs := strings.Split(name, variableGlobalKeyFile)
	if len(vArgs) > 1 {
		vs := c.varsLookupGlobalFile(vArgs[1])
		return vs
	}

	// Last resort, try global vars.
	if name == variableGlobalKey {
		return c.varsLookupGlobal()
	}
	gVars := c.VarsLookupGlobal()
	for _, v := range gVars {
		if v.Name() == name {
			return []common.Variable{v}
		}
	}

	return
}

func (c *Core) varsLookupGlobalFile(register string) (v []common.Variable) {
	return c.varRegistryGlobal.entries[register]
}

func (c *Core) varsLookupGlobal() (v []common.Variable) {
	v = make([]common.Variable, 0)
	for _, vars := range c.varRegistryGlobal.entries {
		v = append(v, vars...)
	}
	return
}

func varLookupRegistry(reg *variableRegistry, register, varName string) (v common.Variable) {
	reg.Lock()
	defer reg.Unlock()

	for _, v := range reg.entries[register] {
		if v.Name() == varName {
			return v
		}
	}

	// If variable is not found, return an empty one.
	return variable{}
}

func varsLookupRegistry(reg *variableRegistry) (v []common.Variable) {
	reg.Lock()
	defer reg.Unlock()

	for _, vs := range reg.entries {
		v = append(v, vs...)
	}

	return
}

//
// Helper.
//

// setLocalVar parses and sets a local variable from the given args.
func (c *Core) setLocalVarByArg(scope string, args []byte) (err error) {
	v := common.VarFromArg(args)
	if v.Name() == "" || v.Value() == "" {
		return errEmptyVariableParameter
	}

	c.setLocalVar(scope, v)
	return
}
