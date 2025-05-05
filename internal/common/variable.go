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
		value: string(TrimQuotes(bytes.TrimSpace(tokens[1]))),
	}
}
