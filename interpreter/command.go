package importer

import (
	"bytes"
	"errors"
	"io"
)

const (
	commandIgnore     = "ignore"
	commandVar        = "var"
	commandForeach    = "foreach"
	commandForeachEnd = "foreachend"
	commandImport     = "import"
)

func (i *Interpreter) executeCommand(command, file string, args [][]byte, out io.Writer) (err error) {
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

	case commandForeach:
		if len(args) == 0 {
			return errors.New("foreach statement needs at least one value")
		}

		for _, arg := range args {
			// Brackets are optional, trim them.
			arg = bytes.Trim(bytes.Trim(bytes.TrimSpace(arg), "["), "]")
			if arg == nil {
				continue
			}

			b := bytes.Split(arg, []byte(","))
			for _, trim := range b {
				if len(trim) == 0 {
					continue
				}

				// Trim braces to get variable name.
				trim = bytes.Trim(bytes.Trim(trim, "{{"), "}}")
				i.setForeachVar(file, string(trim))
			}
		}

	case commandForeachEnd:
		err = i.evaluateForeach(file, out)
		if err != nil {
			return
		}
		i.state.foreach[file] = foreach{}

	}
	return
}
