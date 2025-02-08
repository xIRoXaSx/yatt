package interpreter

import (
	"fmt"
)

var (
	errDependencyUnknownSyntax = fmt.Errorf("unknown syntax: %s <file path>", preprocessorImportName)
)

// dependencyResolver stores information about a single import file.
type dependencyResolver struct {
	deps    dependencies
	visited map[string][]string
}

func newDependencyResolver() dependencyResolver {
	return dependencyResolver{
		deps:    make(dependencies, 0),
		visited: make(map[string][]string, 0),
	}
}

func (d *dependencyResolver) addDependency(origin, lookupPath string) {
	d.deps[origin] = append(d.deps[origin], lookupPath)
}

func (d *dependencyResolver) isAlreadyVisited(lookupPath, destination string) (visited bool) {
	visitedPaths := d.visited[lookupPath]
	if len(visitedPaths) == 0 {
		return false
	}

	for _, vp := range visitedPaths {
		if vp == destination {
			return true
		}

		visited = d.isAlreadyVisited(vp, destination)
		if visited {
			return
		}
	}

	return
}

// dependenciesAreCyclic traverses the whole dependency tree to look if it contains a cycle.
func (d *dependencyResolver) dependenciesAreCyclic(origin, lookupPath string) (cyclic bool) {
	d.visited[origin] = append(d.visited[origin], lookupPath)
	deps := d.deps[lookupPath]
	for _, dep := range deps {
		// Direct neighbor.
		if dep == origin {
			return true
		}

		v := d.isAlreadyVisited(dep, origin)
		if v {
			return true
		}

		cyclic = d.dependenciesAreCyclic(lookupPath, dep)
		if cyclic {
			return
		}
	}

	return
}
