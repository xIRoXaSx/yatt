package interpreter

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/xiroxasx/fastplate/internal/interpreter/commands"
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

func (i *Interpreter) executeCommand(command, file string, args [][]byte, lineNum int, callID string) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("%v (%s)", err, callID)
		}
	}()

	switch command {
	case commandIgnore:
		if len(args) != 1 {
			return fmt.Errorf("command %s: exactly 1 arg expected (start / end)", command)
		}
		i.ignore(file, string(args[0]))

	case commandVar:
		if len(args) < 2 {
			return fmt.Errorf("command %s: expected a name and a value", command)
		}
		i.setScopedVar(file, args)

	case commandForeach:
		if len(args) < 1 {
			return fmt.Errorf("command %s: at least 1 arg expected", command)
		}

		var fe foreach
		ln := lineNum
		fe, err = i.state.foreachLoad(file)
		if err != nil {
			if err != errMapLoadForeach {
				return
			}

			err = nil
			fe = foreach{
				open: 1,
				buf:  queue{v: []foreachBuffer{{ln: ln}}},
				mx:   &sync.Mutex{},
			}

		} else {
			buf := foreachBuffer{ln: ln}
			bufLen := fe.buf.len()
			if bufLen-1 > 0 {
				lastBuf := fe.buf.last()
				lastBuf.nextRef = append(lastBuf.nextRef, fe.c.p)
			}

			var ref *foreachBuffer
			if fe.c.p >= 0 && len(fe.buf.v) > 0 {
				ref = fe.buf.firstN(fe.c.p)
				ref.startNext = append(ref.startNext, len(ref.lines))
			}

			// Check if line is directly nested.
			if (fe.c.p > 0 && buf.ln-fe.buf.v[bufLen-1].ln == 1) && len(fe.buf.v[bufLen-1].lines) == 0 {
				fe.buf.v[bufLen-1].lines = [][]byte{{}}
			}

			fe.buf.push(buf)
			fe.c.j = bufLen - fe.c.p
			fe.c.p = bufLen
			fe.open++
		}

		i.state.foreach.Store(file, fe)
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
				err = i.setForeachVar(file, string(trim))
				if err != nil {
					return
				}
			}
		}

	case commandForeachEnd:
		var fe foreach
		fe, err = i.state.foreachLoad(file)
		if err != nil {
			return
		}

		// Wait until each foreach loop is closed.
		fe.open--
		fe.c.p -= fe.c.j
		i.state.foreach.Store(file, fe)
		if fe.open > 0 {
			return
		}

		err = i.evaluateForeach(fe, file)
		if err != nil {
			return
		}
		fe.buf.v = nil
		fe.c = cursor{}
		i.state.foreach.Store(file, fe)
		return

	case commandIf:
		var stm commands.Statements
		stm, err = i.state.statementLoad(file)
		if err != nil {
			if err != errMapLoadStatements {
				return
			}
			err = nil
			stm = commands.Statements{}
		}

		commands.EvaluateStatement(&stm, file, command, args, func(s string, b []byte) ([]byte, error) {
			return i.resolve(s, b, nil)
		})
		i.state.statements.Store(file, stm)

	case commandElse:
		var stm commands.Statements
		stm, err = i.state.statementLoad(file)
		if err != nil {
			return
		}

		commands.MvToElse(&stm)
		i.state.statements.Store(file, stm)

	case commandIfEnd:
		var stm commands.Statements
		stm, err = i.state.statementLoad(file)
		if err != nil {
			return
		}

		err = commands.Resolve(&stm, file, i.lineEnding, i.state.buf, func(s string, b []byte) ([]byte, error) {
			return i.resolve(s, b, nil)
		})
		i.state.statements.Store(file, stm)
		if err != nil {
			return
		}
	}
	return
}

func (i *Interpreter) appendStatementLine(file string, b []byte) (err error) {
	stm, err := i.state.statementLoad(file)
	if err != nil {
		return
	}

	commands.AppendStatementLine(&stm, b)
	i.state.statements.Store(file, stm)
	return
}
