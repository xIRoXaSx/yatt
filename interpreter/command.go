package importer

import (
	"bytes"
	"fmt"
)

const (
	commandIgnore     = "ignore"
	commandVar        = "var"
	commandForeach    = "foreach"
	commandForeachEnd = "foreachend"
	commandImport     = "import"
	commandIf         = "if"
	commandElse       = "else"
	commandIfEnd      = "ifend"
)

func (i *Interpreter) executeCommand(command, file string, args [][]byte) (err error) {
	switch command {
	case commandIgnore:
		if len(args) != 1 {
			return fmt.Errorf("command %s: exactly 1 arg expected (start / end)", command)
		}
		i.ignore(file, string(args[0]))

	case commandVar:
		if len(args) < 2 {
			return fmt.Errorf("command %s: var statement needs a name and value", command)
		}
		i.setScopedVar(file, args)

	case commandForeach:
		if len(args) < 1 {
			return fmt.Errorf("command %s: at least 1 arg expected", command)
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
		err = i.evaluateForeach(file)
		if err != nil {
			return
		}
		i.state.foreach[file] = foreach{}

	case commandIf:
		if len(args) != 3 {
			return fmt.Errorf("command %s: exactly 3 args expected", command)
		}

		is := i.state.ifStatements[file]
		setStatementFlag := func(b bool) {
			if b {
				is.res = IF
			} else {
				is.res = ELSE
			}
		}

		arg0 := args[0]
		op := string(args[1])
		arg1 := args[2]
		arg0, err = i.resolve(file, arg0, nil)
		if err != nil {
			return
		}
		arg1, err = i.resolve(file, arg1, nil)
		if err != nil {
			return
		}
		switch op {
		case "=", "==":
			setStatementFlag(bytes.Equal(arg0, arg1))

		case "!=", "<>":
			setStatementFlag(!bytes.Equal(arg0, arg1))

		case ">":
			var floats []float64
			floats, err = parseFloats([][]byte{arg0, arg1})
			if err != nil {
				return
			}
			setStatementFlag(floats[0] > floats[1])

		case ">=":
			var floats []float64
			floats, err = parseFloats([][]byte{arg0, arg1})
			if err != nil {
				return
			}
			setStatementFlag(floats[0] >= floats[1])

		case "<":
			var floats []float64
			floats, err = parseFloats([][]byte{arg0, arg1})
			if err != nil {
				return
			}
			setStatementFlag(floats[0] < floats[1])

		case "<=":
			var floats []float64
			floats, err = parseFloats([][]byte{arg0, arg1})
			if err != nil {
				return
			}
			setStatementFlag(floats[0] <= floats[1])

		default:
			return fmt.Errorf("command %s: %s operator unknown", command, op)

		}
		is.active = true
		i.state.ifStatements[file] = is

	case commandElse:
		is := i.state.ifStatements[file]
		is.resPointer = ELSE
		i.state.ifStatements[file] = is

	case commandIfEnd:
		is := i.state.ifStatements[file]
		is.active = false
		i.state.ifStatements[file] = is

		var lines [][]byte
		if is.res == IF {
			lines = i.state.ifStatements[file].ifLines
		} else {
			lines = i.state.ifStatements[file].elseLines
		}

		for _, l := range lines {
			var mod []byte
			mod, err = i.resolve(file, l, nil)
			if err != nil {
				return
			}
			_, err = i.state.buf.Write(append(mod, i.lineEnding...))
			if err != nil {
				return
			}
		}
	}
	return
}

func (i *Interpreter) appendIfStatementLine(file string, b []byte) {
	is := i.state.ifStatements[file]
	if is.resPointer == IF {
		is.ifLines = append(i.state.ifStatements[file].ifLines, b)
	} else {
		is.elseLines = append(i.state.ifStatements[file].elseLines, b)
	}
	i.state.ifStatements[file] = is
}
