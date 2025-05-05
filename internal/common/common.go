package common

import (
	"bytes"
)

func TemplateStart() []byte {
	return []byte("{{")
}

func TemplateEnd() []byte {
	return []byte("}}")
}

func LineEnding() []byte {
	return []byte(lineEnding)
}

// GetLeadingWhitespace returns all leading whitespace characters.
func GetLeadingWhitespace(line []byte) (s []byte) {
	for _, r := range line {
		if r != ' ' && r != '\t' {
			return
		}
		s = append(s, r)
	}
	return
}

func TrimQuotes(val []byte) (ret []byte) {
	quoteDouble := '"'
	quoteSingle := '\''
	ret = bytes.TrimFunc(val, func(r rune) bool {
		return r == quoteDouble || r == quoteSingle
	})
	return
}
