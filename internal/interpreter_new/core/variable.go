package core

import (
	"bytes"
	"os"

	"github.com/xiroxasx/fastplate/internal/common"
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
// Variable setter.
//

func (c *Core) InitLocalVariablesByFiles(varFileNames ...string) {
	// Check if the global var files exist and read it into the memory.
	for _, vf := range varFileNames {
		cont, err := os.ReadFile(vf)
		if err != nil {
			c.l.Fatal().Err(err).Str("path", vf).Msg("unable to read variable file")
		}

		lines := bytes.Split(cont, lineEnding)
		for _, l := range lines {
			split := bytes.Split(c.cutPrefix(l), []byte{' '})
			if string(split[0]) != directiveNameVariable {
				continue
			}

			// Skip the var declaration keyword.
			v := common.VarFromArg(bytes.TrimPrefix(l, split[0]))
			c.setGlobalVar(v)
		}
	}
}

func (c *Core) setGlobalVar(newVar common.Variable) {
	setRegistryVar(&c.varRegistryGlobal, variableRegistryGlobalRegister, newVar)
}

func (c *Core) setLocalVar(register string, newVar common.Variable) {
	setRegistryVar(&c.varRegistryLocal, register, newVar)
}

func setRegistryVar(reg *variableRegistry, register string, newVar common.Variable) {
	reg.Lock()
	defer reg.Unlock()

	for register, vars := range reg.entries {
		for i, v := range vars {
			if newVar.Name() == v.Name() {
				// Update existing variable.
				reg.entries[register][i] = common.NewVar(v.Name(), newVar.Value())
				return
			}
		}
	}

	reg.entries[register] = append(reg.entries[register], newVar)
}

//
// Variable getter.
//

func (c *Core) varLookup(file, name string) (v common.Variable) {
	v = c.varLookupLocal(file, name)
	if v.Name() == "" {
		v = c.varLookupGlobal(name)
	}
	return
}

func (c *Core) varLookupGlobal(name string) (v common.Variable) {
	return varLookupRegistry(&c.varRegistryGlobal, variableRegistryGlobalRegister, name)
}

func (c *Core) varLookupLocal(register, name string) (v common.Variable) {
	return varLookupRegistry(&c.varRegistryLocal, register, name)
}

func (c *Core) varsLookupGlobal() (v []common.Variable) {
	return c.varRegistryGlobal.entries[variableRegistryGlobalRegister]
}

func (c *Core) varsLookupGlobalFile(register string) (v []common.Variable) {
	return c.varRegistryGlobalFile.entries[register]
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
