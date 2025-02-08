package interpreter

import (
	"bytes"
)

// runFileMode runs the import with the targeted Options.OutPath.
func (i *Interpreter) runFileMode(inPath, outPath string) (err error) {
	buf := &bytes.Buffer{}
	return i.writeInterpretedFile(inPath, outPath, buf)
}
