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

type ignoreIndexes map[string]bool

type foreachIndexes map[string]int

type dependencies map[string][]string

type variableRegistry struct {
	entries map[string]vars
	*sync.Mutex
}

type vars []common.Variable

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
