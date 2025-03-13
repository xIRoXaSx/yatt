package common

import (
	"bytes"
	"strings"
)

type variable struct {
	name  string
	value string
}

type Variable interface {
	Name() string
	Value() string
}

// func TemplateStart() []byte {
// 	return []byte("{{")
// }

// func TemplateEnd() []byte {
// 	return []byte("}}")
// }

func (v variable) Name() string {
	return v.name
}

func (v variable) Value() string {
	return v.value
}

func NewVar(name, value string) Variable {
	return variable{name: name, value: value}
}

// VarFromArgs parses a variable name and value from agrs and returns it.
func VarFromArgs(tokens [][]byte) (v Variable) {
	newVar := variable{name: string(tokens[0])}
	val := bytes.SplitN(tokens[0], []byte{'='}, 2)
	if len(val) == 1 {
		// Syntax: x = y, x =y
		val = bytes.Split(tokens[1], []byte{'='})
		newVar.value = string(val[1])
	} else {
		// Syntax: x=y, x= y
		newVar = variable{name: string(val[0]), value: string(val[1])}
	}
	if len(tokens) > 2 {
		// Skip equal sign.
		ind := 1
		if string(tokens[ind]) == "=" {
			// Any case except syntax x= y
			if len(tokens) > ind+1 {
				// Skip next space.
				ind++
			}
		}
		subArgs := tokens[ind:]
		str := make([]string, len(subArgs))
		for j, s := range subArgs {
			str[j] = string(s)
		}
		newVar.value += strings.Join(str, " ")
	}
	return newVar
}
