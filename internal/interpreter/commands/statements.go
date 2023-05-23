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
	lines     [][][]byte
	c         cursor
	initC     cursor
	ln        int
	lnPad     int
	startNext []int
	next      statementPointer
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

func EvaluateStatement(stm *Statements, file, command string, args [][]byte, lineNum int, resolve ResolveFn) (err error) {
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
	if len(stm.statementBuf)-1 < stm.c.p+stm.c.j {
		stm.statementBuf = append(stm.statementBuf, statementBuffer{ln: lineNum, lines: make([][][]byte, 2)})
	}
	if lastBufIdx >= len(stm.statementBuf) {
		lastBufIdx--
	}
	lastBuf := &stm.statementBuf[lastBufIdx]
	stm.statementBuf[stm.c.p].startNext = append(stm.statementBuf[stm.c.p].startNext, lineNum-lastBuf.ln)
	defer func() {
		lastBuf.next = statementPointer{
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
	buf := &stm.statementBuf[stm.c.p-1]
	buf.lnPad++
	buf.initC.p = 1
}

// Currently only non-same-levelled if-statements are read correctly...
func Resolve(stm *Statements, file string, lineEnding []byte, w io.Writer, resolve ResolveFn) (err error) {
	stm.open--
	//stm.statementBuf[stm.c.p+stm.c.j].lnPad++
	stm.c.j = len(stm.statementBuf) - stm.c.p
	if stm.c.p-1 > 0 {
		stm.c.p--
		return
	}

	var idx int
	for {
		buf := stm.statementBuf[idx]
		var last int
		for _, ln := range buf.startNext {
			last, err = test(stm, idx, last, ln, file, lineEnding, w, resolve)
			if err != nil {
				return
			}

			if buf.c.p != buf.next.statement {
				return
			}

			idx++
			buf = stm.statementBuf[idx]
		}
		if idx == len(stm.statementBuf) {
			break
		}
	}

	return
}

func test(stm *Statements, idx, start, until int, file string, lineEnding []byte, w io.Writer, resolve ResolveFn) (iter int, err error) {
	buf := stm.statementBuf[idx]
	ln := buf.ln
	//ln += buf.lnPad
	next := stm.statementBuf[idx]
	if idx+1 < len(stm.statementBuf) {
		next = stm.statementBuf[idx+1]
	}

	for i, l := range buf.lines[buf.c.p][start:] {
		if i == until {
			ln = i
		}

		for _, nextIdx := range buf.startNext {
			if i+1 != nextIdx {
				continue
			}

			_, err = test(stm, idx+1, 0, until, file, lineEnding, w, resolve)
			if err != nil {
				return
			}
			continue
		}

		ln++
		if ln == next.ln {
			iter = i
			_, err = test(stm, idx+1, i, until, file, lineEnding, w, resolve)
			if err != nil {
				return
			}
			continue
		}

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
