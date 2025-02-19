package interpreter

import "fmt"

// runFileMode runs the import with the targeted Options.OutPath.
func (i *Interpreter) runFileMode(inPath, outPath string) (err error) {
	// First check if we have cyclic dependencies.
	err = i.importPathCheckCyclicDependencies(inPath)
	if err != nil {
		return fmt.Errorf("dependency check: %v", err)
	}

	return i.writeInterpretedFile(inPath, outPath)
}
