package common

type Interpreter interface {
	Ignore(filename string, arg string)
	SetScopedVar(scope string, args [][]byte)
}
