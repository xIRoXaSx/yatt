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
	resolveJump  int
}

type statementBuffer struct {
	lines     [][][]byte
	c         cursor
	initC     cursor
	ln        int
	startNext []int
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
	fmt.Println("\tnum:", lineNum, "open:", stm.open)
	if len(stm.statementBuf)-1 < statementPtr(stm)+1 || stm.resolveJump > 0 {
		stm.statementBuf = append(stm.statementBuf, statementBuffer{ln: lineNum, lines: make([][][]byte, 2)})
	}
	lastBufIdx := int(math.Max(0, float64(len(stm.statementBuf)-1)))
	if lastBufIdx > 0 {
		lastBufIdx--
	}
	lastBuf := &stm.statementBuf[lastBufIdx]
	stm.statementBuf[stm.c.p].startNext = append(stm.statementBuf[stm.c.p].startNext, lineNum-lastBuf.ln)

	// Get the first (left) part of the if statement.
	templateStart := common.TemplateStart()
	templateEnd := common.TemplateEnd()
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
	var eval bool
	switch string(op) {
	case "=", "==":
		eval = bytes.Equal(arg0, arg1)

	case "!=", "<>":
		eval = !bytes.Equal(arg0, arg1)

	case ">":
		var floats []float64
		floats, err = functions.ParseFloats([][]byte{arg0, arg1})
		if err != nil {
			return
		}
		eval = floats[0] > floats[1]

	case ">=":
		var floats []float64
		floats, err = functions.ParseFloats([][]byte{arg0, arg1})
		if err != nil {
			return
		}
		eval = floats[0] >= floats[1]

	case "<":
		var floats []float64
		floats, err = functions.ParseFloats([][]byte{arg0, arg1})
		if err != nil {
			return
		}
		eval = floats[0] < floats[1]

	case "<=":
		var floats []float64
		floats, err = functions.ParseFloats([][]byte{arg0, arg1})
		if err != nil {
			return
		}
		eval = floats[0] <= floats[1]

	default:
		return fmt.Errorf("command %s: %s operator unknown", command, op)

	}

	//buf := &stm.statementBuf[stm.c.p]
	buf := &stm.statementBuf[stm.c.p+stm.c.j]
	if eval {
		buf.c.p = 0
	} else {
		buf.c.p = 1
	}
	stm.c.p++

	stm.c.j = len(stm.statementBuf) - stm.open
	stm.resolveJump = 0
	return
}

func MvToElse(stm *Statements, lineNum int) {
	////stm.statementBuf[stm.c.p-1].initC.p = 1
	fmt.Println("Else:", lineNum, statementPtr(stm), stm.c.j)
	stm.statementBuf[statementPtr(stm)].initC.p = 1
}

func Resolve(stm *Statements, file string, idx int, lineEnding []byte, w io.Writer, resolve ResolveFn) (err error) {
	defer func() {
		jumped := len(stm.statementBuf) - stm.open
		if jumped > 0 {
			stm.resolveJump = jumped
			fmt.Println("jumped:", jumped, stm.c.j, stm.c.j+stm.c.p)
		}
	}()

	stm.open--
	if stm.open > 0 {
		stm.c.p--
		return
	}

	loopLines := func(idx, start int) (next int, err error) {
		buf := stm.statementBuf[idx]
		for i, l := range buf.lines[buf.c.p][start:] {
			for _, ln := range buf.startNext {
				if i+1 == ln {
					next = i + 1
					break
				}
			}
			// Check if next statement is required.
			if next != 0 {
				return
			}

			// Write back.
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

	// DEBUG ONLY!
	for i, buf := range stm.statementBuf {
		for s, statement := range buf.lines {
			for _, l := range statement {
				fmt.Println(i, s, string(l))
			}
		}
	}

	for {
		var next int
		next, err = loopLines(idx, next)
		if err != nil {
			return
		}

		// TODO: Result2 is one layer too deep!
		if next != 0 {
			for next != 0 && next-1 < len(stm.statementBuf[idx].lines[stm.c.p]) {
				var n int
				_ = n
				n, err = loopLines(idx+1, 0)
				if err != nil {
					return
				}

				next, err = loopLines(idx, next-1)
				if err != nil {
					return
				}
			}
			// This index has already been used + looped through, skip it.
			idx++
		}

		idx++
		if idx == len(stm.statementBuf) {
			break
		}
	}

	return
}

// TODO: Sibling statements cause shifting...
func AppendStatementLine(stm *Statements, b []byte) {
	stmP := statementPtr(stm)
	buf := stm.statementBuf[stmP]
	stm.statementBuf[stmP].lines[buf.initC.p] = append(stm.statementBuf[stmP].lines[buf.initC.p], b)
	return
}

func statementPtr(stm *Statements) int {
	return stm.c.p - 1 + stm.c.j
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
