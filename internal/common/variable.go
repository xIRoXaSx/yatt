package common

type variable struct {
	name  string
	value string
}

type Variable interface {
	Name() string
	Value() string
}

func TemplateStart() []byte {
	return []byte("{{")
}

func TemplateEnd() []byte {
	return []byte("}}")
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
