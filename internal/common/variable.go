package common

type Variable struct {
	name  string
	value string
}

type Var interface {
	Name() string
	Value() string
}

func TemplateStart() []byte {
	return []byte("{{")
}

func TemplateEnd() []byte {
	return []byte("}}")
}

func (v Variable) Name() string {
	return v.name
}

func (v Variable) Value() string {
	return v.value
}

func NewVar(name, value string) Var {
	return Variable{name: name, value: value}
}
