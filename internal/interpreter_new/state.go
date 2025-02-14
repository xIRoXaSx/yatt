package interpreter

import (
	"bytes"
	"sync"

	"github.com/xiroxasx/fastplate/internal/common"
)

type state struct {
	ignoreIndex       ignoreIndexes
	depsResolver      dependencyResolver
	foreachIndex      foreachIndexes
	varRegistryLocal  variableRegistry
	varRegistryGlobal variableRegistry // TODO: Currently merging unscopedVarIndexes into this as well!
	foreach           sync.Map
	buf               *bytes.Buffer
	*sync.Mutex
}

type ignoreIndexes map[string]ignoreState

type foreachIndexes map[string]int

type dependencies map[string][]string

type variableRegistry struct {
	entries map[string]vars
	*sync.Mutex
}

type vars []common.Variable

type ignoreState uint8

const (
	ignoreStateOpen ignoreState = iota
	ignoreStateClose

	variableRegistryGlobalRegisterGlobal = "global"
)

func (s *state) registerGlobalVar(register string, newVar common.Variable) {
	s.varRegistryGlobal.Lock()
	defer s.varRegistryGlobal.Unlock()

	for register, vars := range s.varRegistryGlobal.entries {
		for i, v := range vars {
			if newVar.Name() == v.Name() {
				// Update existing variable.
				s.varRegistryGlobal.entries[register][i] = common.NewVar(v.Name(), newVar.Value())
				return
			}
		}
	}

	s.varRegistryGlobal.entries[register] = append(s.varRegistryGlobal.entries[register], newVar)
}

func (s *state) varLookup(file, name string) (v variable) {
	v = s.varLookupLocal(file, name)
	if v.Name() == "" {
		v = s.varLookupGlobal(name)
	}
	return
}

func (s *state) varLookupGlobal(name string) (v variable) {
	for _, v := range s.varRegistryGlobal.entries[variableRegistryGlobalRegisterGlobal] {
		if v.Name() == name {
			return v.(variable)
		}
	}
	return variable{}
}

func (s *state) varLookupLocal(register, name string) (v variable) {
	for _, v := range s.varRegistryLocal.entries[register] {
		if v.Name() == name {
			return v.(variable)
		}
	}
	return variable{}
}
