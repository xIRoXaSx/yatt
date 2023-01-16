package importer

import (
	"bytes"
	"io"
	"log"
	"os"
)

func (i *Importer) interpretFile(inPath string, indent []byte, out io.Writer) (err error) {
	cont, err := os.ReadFile(inPath)
	if err != nil {
		log.Printf("warn: unable to read file %s: %v\n", inPath, err)
		return err
	}
	// Prepend indention to all linebreaks.
	cutSet := []byte{'\n'}
	if len(indent) > 0 {
		cont = bytes.ReplaceAll(cont, cutSet, append(cutSet, indent...))
		cont = append(indent, cont...)
	}

	lines := bytes.SplitAfter(cont, []byte{'\n'})
	for _, l := range lines {
		if i.opts.Indent {
			indent = pushLeadingIndent(l)
		}

		// Skip the indents.
		linePart := l[len(indent):]
		prefix := i.matchedImportPrefix(linePart)
		if prefix == nil {
			_, err = out.Write(l)
		} else {
			// Check if file of statement exists and read it if it does.
			statement := string(bytes.Trim(bytes.TrimPrefix(linePart, prefix), string(append(cutSet, ' '))))
			err = i.interpretFile(statement, indent, out)
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
