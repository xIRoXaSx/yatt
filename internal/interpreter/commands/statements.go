package commands

import (
	"bytes"
	"fmt"
	"io"
	"math"

	"github.com/xiroxasx/fastplate/internal/common"
	"github.com/xiroxasx/fastplate/internal/interpreter/functions"
)

type Statements struct {
	statementBuf []statementBuffer
	c            cursor
	open         int
}

type statementBuffer struct {
	lines         [][][]byte
	c             cursor
	initC         cursor
	refFrom       pointer
	nextRef       int
	nextStatement statementPointer
}

type statementPointer struct {
	index     int
	statement int
}

type pointer struct {
	index int
	c     cursor
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
	stm.open++
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

	lastBufIdx := int(math.Max(0, float64(len(stm.statementBuf)-1)))
	if len(stm.statementBuf)-1 < stm.c.p {
		stm.statementBuf = append(stm.statementBuf, statementBuffer{lines: make([][][]byte, 2)})
	}
	if lastBufIdx >= len(stm.statementBuf) {
		lastBufIdx--
	}
	lastBuf := &stm.statementBuf[lastBufIdx]

	latest := &stm.statementBuf[stm.c.p]
	latest.refFrom = pointer{index: lastBufIdx, c: cursor{p: int(lastBuf.c.p)}}
	defer func() {
		lastBuf.nextStatement = statementPointer{
			index:     stm.open - 1,
			statement: lastBuf.initC.p,
		}
	}()

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
	stm.statementBuf[stm.c.p-1].initC.p = 1
}

// TODO: Currently all if statements get evaluated.
// Add option to "follow" the "evaluation path".
func Resolve(stm *Statements, file string, lineEnding []byte, w io.Writer, resolve ResolveFn) (err error) {
	stm.open--
	if stm.c.p-1 > 0 {
		stm.c.p--
		return
	}

	idx := 0
	for {
		buf := stm.statementBuf[idx]
		for _, l := range buf.lines[buf.c.p] {
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

		if idx+1 == len(stm.statementBuf) || buf.c.p != buf.nextStatement.statement {
			break
		}
		idx++
	}

	return
}

func AppendStatementLine(stm *Statements, b []byte) {
	stmP := stm.c.p - 1
	buf := stm.statementBuf[stmP]
	stm.statementBuf[stmP].lines[buf.initC.p] = append(stm.statementBuf[stmP].lines[buf.initC.p], b)
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
