package interpreter

import (
	"sync"

	"github.com/xiroxasx/fastplate/internal/common"
)

type state struct {
	ignoreIndex  ignoreIndexes
	depsResolver dependencyResolver
	foreachBuff  foreachBufferStack
	// foreachIndex       foreachIndexes
	varRegistryLocal      variableRegistry
	varRegistryGlobal     variableRegistry // TODO: Currently merging unscopedVarIndexes into this as well!
	varRegistryGlobalFile variableRegistry // TODO: Currently merging unscopedVarIndexes into this as well!

	*sync.Mutex
}

// type foreachIndexes map[string]int

type dependencies map[string][]string

// TODO: Implement "lookup(register, name string)".
type variableRegistry struct {
	entries map[string]vars
	*sync.Mutex
}

type vars []common.Variable

type ignoreState uint8

const (
	ignoreStateClose ignoreState = iota
	ignoreStateOpen

	variableRegistryGlobalRegister = "global"
)

//
// Variable setter.
//

func (s *state) setGlobalVar(newVar common.Variable) {
	setRegistryVar(&s.varRegistryGlobal, variableRegistryGlobalRegister, newVar)
}

func (s *state) setLocalVar(register string, newVar common.Variable) {
	setRegistryVar(&s.varRegistryLocal, register, newVar)
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

func (s *state) varLookup(file, name string) (v common.Variable) {
	v = s.varLookupLocal(file, name)
	if v.Name() == "" {
		v = s.varLookupGlobal(name)
	}
	return
}

func (s *state) varLookupGlobal(name string) (v common.Variable) {
	return varLookupRegistry(&s.varRegistryGlobal, variableRegistryGlobalRegister, name)
}

func (s *state) varLookupLocal(register, name string) (v common.Variable) {
	return varLookupRegistry(&s.varRegistryLocal, register, name)
}

func (s *state) varsLookupGlobal() (v []common.Variable) {
	return s.varRegistryGlobal.entries[variableRegistryGlobalRegister]
}

func (s *state) varsLookupGlobalFile(register string) (v []common.Variable) {
	return s.varRegistryGlobalFile.entries[register]
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

//
// Helper.
//

// setLocalVar parses and sets a local variable from the given args.
func (s *state) setLocalVarByArgs(scope string, args [][]byte) (err error) {
	v := variableFromArgs(args)
	if v.Name() == "" || v.Value() == "" {
		return errEmptyVariableParameter
	}

	s.setLocalVar(scope, v)
	return
}
