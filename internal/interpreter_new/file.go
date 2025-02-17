package interpreter

// runFileMode runs the import with the targeted Options.OutPath.
func (i *Interpreter) runFileMode(inPath, outPath string) (err error) {
	return i.writeInterpretedFile(inPath, outPath)
}
