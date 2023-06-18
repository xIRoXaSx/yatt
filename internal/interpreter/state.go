package interpreter

import (
	"errors"
	"fmt"

	"github.com/xiroxasx/fastplate/internal/common"
)

var errMapLoadForeach = errors.New("unable to load foreach map values")

func (s *state) varLookup(file, name string) (v common.Var) {
	v = s.lookupScoped(file, name)
	if v.Name() == "" {
		v = s.lookupUnscoped(name)
	}
	return
}

func (s *state) lookupUnscoped(name string) common.Var {
	for _, v := range s.unscopedVars {
		if v.Name() == name {
			return v
		}
	}
	return variable{}
}

func (s *state) lookupScoped(fileName, name string) common.Var {
	for _, v := range s.scopedRegistry.scopedVars[fileName] {
		if v.Name() == name {
			return v
		}
	}
	return variable{}
}

func (s *state) addDependency(fileName, dependency string) {
	s.dependencies[fileName] = append(s.dependencies[fileName], dependency)
}

// hasCyclicDependency walks down the dependencies to check whether the given dependency creates a loop.
// Returns true if a cycle has been detected.
func (s *state) hasCyclicDependency(fileName, dependency string) bool {
	for _, d := range s.dependencies[dependency] {
		if d == fileName {
			return true
		} else if d == "" {
			return false
		}
		return s.hasCyclicDependency(fileName, d)
	}
	return false
}

// foreachLoad loads the value with key and casts it to foreach.
func (s *state) foreachLoad(key string) (fe foreach, err error) {
	ife, ok := s.foreach.Load(key)
	if !ok {
		err = errMapLoadForeach
		return
	}
	fe, ok = ife.(foreach)
	if !ok {
		err = fmt.Errorf("unable to cast foreach's value")
		return
	}
	return
}
