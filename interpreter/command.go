package importer

import "errors"

const (
	commandIgnore = "ignore"
	commandVar    = "var"
)

func (i *Interpreter) executeCommand(command, file string, args [][]byte) (err error) {
	switch command {
	case commandIgnore:
		if len(args) == 0 {
			return errors.New("ignore statement needs either start or end flag")
		}
		i.ignore(file, string(args[0]))

	case commandVar:
		if len(args) == 0 {
			return errors.New("var statement needs a name and value")
		}
		i.setScopedVar(file, args)
	}
	return
}
