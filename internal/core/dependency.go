package core

import (
	"sync"
)

type dependencies map[string][]string

// dependencyResolver stores information about a single import file.
type dependencyResolver struct {
	deps    dependencies
	visited []string
	mx      *sync.Mutex
}

func newDependencyResolver() dependencyResolver {
	return dependencyResolver{
		deps: make(dependencies, 0),
		mx:   &sync.Mutex{},
	}
}

func (d *dependencyResolver) addDependency(origin, destination string) {
	d.mx.Lock()
	defer d.mx.Unlock()

	d.deps[origin] = append(d.deps[origin], destination)
}

func (d *dependencyResolver) CheckForCyclicDependencies(start, destination string) (cyclic bool) {
	d.visited = make([]string, 0)

	return d.dependenciesAreCyclic(start, destination)
}

// dependenciesAreCyclic traverses the whole dependency tree to look if it contains a cycle.
func (d *dependencyResolver) dependenciesAreCyclic(start, destination string) (cyclic bool) {
	// Same files cannot be imported.
	if start == destination {
		return true
	}

	d.mx.Lock()
	deps := d.deps[destination]
	d.mx.Unlock()
	for _, dep := range deps {
		if dep == start || dep == destination {
			return true
		}
		for _, v := range d.visited {
			if dep == v {
				// Dependency may not be part of the already visited route.
				return true
			}
		}

		d.visited = append(d.visited, destination)
		cyclic = d.dependenciesAreCyclic(destination, dep)
		if cyclic {
			return
		}
	}

	return
}
