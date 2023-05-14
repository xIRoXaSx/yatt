package commands

import (
	"bytes"
	"fmt"
	"io"

	"github.com/xiroxasx/fastplate/internal/common"
	"github.com/xiroxasx/fastplate/internal/interpreter/functions"
)

type Statements struct {
	statementBuf []statementBuffer
	c            cursor
}

type statementBuffer struct {
	lines [][][]byte
	c     cursor
	condC cursor
}

type cursor struct {
	p int
	j int
}

type ResolveFn func(string, []byte) ([]byte, error)

func (stm *Statements) Active() bool {
	return stm.c.p > 0
}

func EvaluateStatement(stm *Statements, file, command string, args [][]byte, resolve ResolveFn) (err error) {
	templateStart := common.TemplateStart()
	templateEnd := common.TemplateEnd()
	setStatementFlag := func(b bool) {
		if b {
			stm.statementBuf[stm.c.p].c.p = 0
		} else {
			stm.statementBuf[stm.c.p].c.p = 1
		}
		stm.c.p++
	}

	if len(stm.statementBuf)-1 < stm.c.p {
		stm.statementBuf = append(stm.statementBuf, statementBuffer{lines: make([][][]byte, 2)})
	}

	// Get the first (left) part of the if statement.
	r, idx := matchUntilSymbol(args, templateStart, templateEnd)
	arg0 := bytes.Join(r, []byte(""))
	if idx+1 > len(args) {
		return fmt.Errorf("command %s: unable to find matching end", command)
	}

	// Get the operator of the if statement.
	op, idOp := findSymbols(args[idx+1:], statementOperators())
	idx += idOp

	// Get the last (right) part of the if statement.
	r, _ = matchUntilSymbol(args[idx+1:], templateStart, templateEnd)
	arg1 := bytes.Join(r, []byte(""))
	arg0, err = resolve(file, arg0)
	if err != nil {
		return
	}
	arg1, err = resolve(file, arg1)
	if err != nil {
		return
	}
	switch string(op) {
	case "=", "==":
		setStatementFlag(bytes.Equal(arg0, arg1))

	case "!=", "<>":
		setStatementFlag(!bytes.Equal(arg0, arg1))

	case ">":
		var floats []float64
		floats, err = functions.ParseFloats([][]byte{arg0, arg1})
		if err != nil {
			return
		}
		setStatementFlag(floats[0] > floats[1])

	case ">=":
		var floats []float64
		floats, err = functions.ParseFloats([][]byte{arg0, arg1})
		if err != nil {
			return
		}
		setStatementFlag(floats[0] >= floats[1])

	case "<":
		var floats []float64
		floats, err = functions.ParseFloats([][]byte{arg0, arg1})
		if err != nil {
			return
		}
		setStatementFlag(floats[0] < floats[1])

	case "<=":
		var floats []float64
		floats, err = functions.ParseFloats([][]byte{arg0, arg1})
		if err != nil {
			return
		}
		setStatementFlag(floats[0] <= floats[1])

	default:
		return fmt.Errorf("command %s: %s operator unknown", command, op)

	}
	return
}

func MvToElse(stm *Statements) {
	stm.statementBuf[stm.c.p-1].condC.p = 1
}

// TODO: Currently all if statements get evaluated.
// Add option to "follow" the "evaluation path".
func Resolve(stm *Statements, file string, lineEnding []byte, w io.Writer, resolve ResolveFn) (err error) {
	if stm.c.p-1 > 0 {
		stm.c.p--
		return
	}

	for _, sb := range stm.statementBuf {
		for _, l := range sb.lines[sb.c.p] {
			fmt.Println(string(l))
			var mod []byte
			mod, err = resolve(file, l)
			if err != nil {
				return
			}
			_, err = w.Write(append(mod, lineEnding...))
			if err != nil {
				return
			}
		}
	}
	return
}

func AppendStatementLine(stm *Statements, b []byte) {
	stmP := stm.c.p - 1
	buf := stm.statementBuf[stmP]
	stm.statementBuf[stmP].lines[buf.condC.p] = append(stm.statementBuf[stmP].lines[buf.condC.p], b)
	return
}

func matchUntilSymbol(val [][]byte, matchSymbol, untilSymbol []byte) (ret [][]byte, idx int) {
	var (
		openCount int
		b         []byte
	)
	idx = -1
	for idx, b = range val {
		if len(bytes.TrimSpace(b)) == 0 {
			continue
		}

		ret = append(ret, b)
		openCount += bytes.Count(b, matchSymbol)
		openCount -= bytes.Count(b, untilSymbol)
		if openCount == 0 {
			return
		}
	}
	return
}

func findSymbols(val [][]byte, operators [][]byte) (ret []byte, idx int) {
	var b []byte
	idx = -1
	for idx, b = range val {
		v := bytes.TrimSpace(b)
		if len(v) == 0 {
			continue
		}
		for _, op := range operators {
			if bytes.Equal(v, op) {
				idx++
				ret = b
				return
			}
		}
	}
	return
}

func statementOperators() [][]byte {
	return [][]byte{
		[]byte("=="),
		[]byte("="),
		[]byte("!="),
		[]byte("<>"),
		[]byte(">="),
		[]byte(">"),
		[]byte("<="),
		[]byte("<"),
	}
}
