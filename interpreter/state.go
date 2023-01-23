package importer

func (s state) lookupUnscoped(name string) variable {
	for _, v := range s.unscopedVars {
		if v.name == name {
			return v
		}
	}
	return variable{}
}

func (s state) lookupScoped(fileName, name string) variable {
	for _, v := range s.scopedVars[fileName] {
		if v.name == name {
			return v
		}
	}
	return variable{}
}

func (s state) addDependency(fileName, dependency string) {
	s.dependencies[fileName] = append(s.dependencies[fileName], dependency)
}

// hasCyclicDependency walks down the dependencies to check whether the given dependency has creates a loop.
// Returns true if a cycle has been detected.
func (s state) hasCyclicDependency(fileName, dependency string) bool {
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

func (s state) followDependency(dependency, target string) bool {
	for _, d := range s.dependencies[dependency] {
		if d == target {
			return true
		} else if d == "" {
			return false
		}
		return s.followDependency(d, target)
	}
	return false
}
