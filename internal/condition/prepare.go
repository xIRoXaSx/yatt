package condition

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var (
	ErrNoOpenCondition = errors.New("no open condition")
	ErrElseAlreadySeen = errors.New("else already seen")
	ErrElseIfAfterElse = errors.New("elseif after else")
)

func (b *Buffer) IsTrue(fileName string, args []Arg, tr TokenResolver) (eval bool, err error) {
	expr := bytes.TrimSpace(bytes.Join(argsToBytes(args), []byte{' '}))
	if len(expr) == 0 {
		return false, errors.New("empty condition")
	}

	operators := [][]byte{
		[]byte("=="),
		[]byte("!="),
		[]byte(">="),
		[]byte("<="),
		[]byte(">"),
		[]byte("<"),
	}

	for _, op := range operators {
		before, after, ok := bytes.Cut(expr, op)
		if !ok {
			continue
		}

		left, lErr := resolveOperand(fileName, before, tr)
		if lErr != nil {
			return false, lErr
		}
		right, rErr := resolveOperand(fileName, after, tr)
		if rErr != nil {
			return false, rErr
		}
		return compare(left, right, string(op))
	}

	value, err := resolveOperand(fileName, expr, tr)
	if err != nil {
		return false, err
	}
	return isTruthy(value), nil
}

func argsToBytes(args []Arg) [][]byte {
	ret := make([][]byte, len(args))
	for i := range args {
		ret[i] = args[i]
	}
	return ret
}

func resolveOperand(fileName string, raw []byte, tr TokenResolver) (string, error) {
	raw = bytes.TrimSpace(raw)
	resolved, err := tr.Resolve(fileName, raw)
	if err != nil {
		return "", err
	}
	return trimOperand(string(resolved)), nil
}

func trimOperand(s string) string {
	s = strings.TrimSpace(s)
	s = strings.Trim(s, `"'`)
	return strings.TrimSpace(s)
}

func compare(left, right, op string) (bool, error) {
	switch op {
	case "==":
		return left == right, nil
	case "!=":
		return left != right, nil
	}

	leftNum, err := strconv.ParseFloat(left, 64)
	if err != nil {
		return false, fmt.Errorf("parse left operand %q as number: %w", left, err)
	}
	rightNum, err := strconv.ParseFloat(right, 64)
	if err != nil {
		return false, fmt.Errorf("parse right operand %q as number: %w", right, err)
	}

	switch op {
	case ">":
		return leftNum > rightNum, nil
	case ">=":
		return leftNum >= rightNum, nil
	case "<":
		return leftNum < rightNum, nil
	case "<=":
		return leftNum <= rightNum, nil
	default:
		return false, fmt.Errorf("unknown condition operator %q", op)
	}
}

func isTruthy(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "false", "0", "no", "off":
		return false
	default:
		return true
	}
}
