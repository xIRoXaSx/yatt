package importer

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

func (i *Importer) interpretFile(filePath string, indent []byte, out io.Writer) (err error) {
	cont, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("warn: unable to read file %s: %v\n", filePath, err)
		return
	}

	// Append indention to all linebreaks, prepend to the first line.
	cutSet := []byte{'\n'}
	if len(indent) > 0 {
		cont = bytes.ReplaceAll(cont, cutSet, append(cutSet, indent...))
		cont = append(indent, cont...)
	}

	lines := bytes.Split(cont, cutSet)
	for _, l := range lines {
		if i.opts.Indent {
			indent = pushLeadingIndent(l)
		}

		// Skip the indents.
		linePart := l[len(indent):]
		prefix := i.matchedImportPrefix(linePart)
		if prefix == nil {
			// Line does not contain one of the required prefixes.
			if i.state.ignoreIndex[filePath] == 1 {
				// Still in an ignore block.
				continue
			}
			_, err = out.Write(append(i.resolve(filePath, l), cutSet...))
			if err != nil {
				return
			}
		} else {
			// Trim statement and check against internal commands.
			statement := i.TrimLine(linePart, prefix)
			split := bytes.Split(statement, []byte{' '})
			if len(split) > 1 {
				err = i.executeCommand(string(split[0]), filePath, split[1:])
				if err != nil {
					return
				}
				continue
			}

			stmnt := filepath.Clean(string(statement))
			filePath = filepath.Clean(filePath)
			if i.state.hasCyclicDependency(filePath, stmnt) {
				err = fmt.Errorf("detected import cycle: %s -> %s", filePath, stmnt)
				return
			}
			i.state.addDependency(filePath, stmnt)
			err = i.interpretFile(stmnt, indent, out)
			if err != nil {
				return err
			}
		}
	}
	return
}

func (i *Importer) matchedImportPrefix(line []byte) []byte {
	for _, pref := range i.prefixes {
		if bytes.HasPrefix(line, []byte(pref)) {
			return []byte(pref)
		}
	}
	return nil
}

func pushLeadingIndent(line []byte) (s []byte) {
	for _, r := range line {
		if r != ' ' && r != '\t' {
			break
		}
		s = append(s, r)
	}
	return
}
