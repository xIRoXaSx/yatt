package common

import (
	"bytes"
)

type variable struct {
	name  string
	value string
}

type Variable interface {
	Name() string
	Value() string
}

func (v variable) Name() string {
	return v.name
}

func (v variable) Value() string {
	return v.value
}

func NewVar(name, value string) Variable {
	return variable{name: name, value: value}
}

func VarFromArg(arg []byte) (_ variable) {
	if len(arg) == 0 {
		return
	}

	tokens := bytes.SplitN(arg, []byte{'='}, 2)
	if len(tokens) == 1 {
		return
	}

	return variable{
		name:  string(bytes.TrimSpace(tokens[0])),
		value: string(bytes.TrimSpace(TrimQuotes(tokens[1]))),
	}
}

// TODO: Add VarFromArgs as []byte instead.
//		 This way we don't need to take care of token splits.

// VarFromArgs parses a variable from args and returns it as [Variable].
func VarFromArgs(args [][]byte) (v variable) {
	if len(args) == 0 {
		return
	}

	for idx, token := range args {
		// Skip empty tokens.
		if len(token) == 0 {
			continue
		}

		// Syntax: "x = y".
		if bytes.Equal(token, []byte{'='}) {
			varNameIdx := 0
			if idx > 0 {
				varNameIdx = idx - 1
			}
			var remainder []byte
			if idx < len(args)-1 {
				// Skip the equal sign.
				remainder = bytes.Join(args[idx+1:], []byte{' '})
			}
			// remainder could potentially be empty if we are at last loop index.
			return variable{
				name:  string(bytes.Join(args[:varNameIdx+1], []byte{})),
				value: string(remainder),
			}
		}

		tok := bytes.SplitN(token, []byte{'='}, 2)
		if len(tok) != 2 {
			continue
		}

		name := tok[0]
		value := tok[1]

		// Syntax: "x= y".
		if len(value) == 0 {
			var remainder []byte
			if idx+1 < len(args) {
				remainder = bytes.Join(args[idx+1:], []byte{})
			}
			return variable{
				name:  string(tok[0]),
				value: string(remainder),
			}
		}

		// Syntax: "x =y".
		if len(name) == 0 {
			return variable{
				name:  string(bytes.Join(args[:idx], []byte{})),
				value: string(value),
			}
		}

		// Syntax: "x=y".
		return variable{
			name:  string(name),
			value: string(value),
		}
	}

	return
}
