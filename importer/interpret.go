package importer

import (
	"bytes"
	"io"
	"log"
	"os"
)

func (i *Importer) interpretFile(stmnt string, indent []byte, out io.Writer) (err error) {
	cont, err := os.ReadFile(stmnt)
	if err != nil {
		log.Printf("warn: unable to read file %s: %v\n", stmnt, err)
		return err
	}
	// Prepend indention to all linebreaks.
	cutSet := []byte{'\n'}
	if len(indent) > 0 {
		cont = bytes.ReplaceAll(cont, cutSet, append(cutSet, indent...))
		cont = append(indent, cont...)
	}

	lines := bytes.SplitAfter(cont, cutSet)
	for _, l := range lines {
		if i.opts.Indent {
			indent = pushLeadingIndent(l)
		}

		// Skip the indents.
		linePart := l[len(indent):]
		prefix := i.matchedImportPrefix(linePart)
		if prefix == nil {
			// Still in an ignore block.
			if i.state.ignoreIndex[stmnt] == 1 {
				continue
			}

			l = i.resolve(stmnt, l)
			_, err = out.Write(l)
		} else {
			// Trim statement and check against internal commands.
			statement := bytes.Trim(bytes.TrimPrefix(linePart, prefix), string(append(cutSet, ' ')))
			split := bytes.Split(statement, []byte{' '})
			if len(split) > 1 {
				err = i.executeCommand(string(split[0]), stmnt, split[1:])
				if err != nil {
					return
				}
				continue
			}

			err = i.interpretFile(string(statement), indent, out)
			if err != nil {
				return err
			}
			// Append new line after the imported content since the statement contains one too.
			_, err = out.Write(cutSet)
			if err != nil {
				return
			}
		}
		if err != nil {
			return
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
