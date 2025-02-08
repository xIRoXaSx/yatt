package interpreter

import (
	"bytes"
	"io"
	"path/filepath"

	"github.com/xiroxasx/fastplate/internal/common"
)

type interpreterFile struct {
	name   string
	rc     io.ReadCloser
	writer io.Writer
}

type recurringToken uint8

const (
	recurringTokenIgnore recurringToken = iota + 1
	recurringTokenForeach
)

func (i *Interpreter) preprocessorState(fileName string) (t recurringToken) {
	// Line does not contain one of the required prefixes.
	if i.state.ignoreIndex[fileName] {
		return recurringTokenIgnore
	} else if i.state.foreachIndex[fileName] != 0 {
		return recurringTokenForeach
	}

	return
}

func (i *Interpreter) searchTokensAndExecute(fileName string, line, indentParent, indentLine []byte, buf io.Writer, lineNum int) (err error) {
	lineDisplayNum := lineNum + 1
	lineNoIndent := line[len(indentLine):]
	prefix := i.matchedPrefixToken(lineNoIndent)
	if len(prefix) > 0 {
		// Trim the prefix and check against internal commands.
		statement := i.trimLine(lineNoIndent, prefix)
		split := bytes.Split(statement, []byte{' '})
		if len(split) == 0 {
			return
		}

		pd := &preprocessorDirective{
			name:     string(split[0]),
			fileName: filepath.Clean(fileName),
			args:     split[1:],
			indent:   indentParent,
			buf:      &bytes.Buffer{},
		}
		err = i.preprocess(pd, lineDisplayNum)
		if err != nil {
			return
		}

		_, err = pd.buf.WriteTo(buf)
		return
	}

	switch i.preprocessorState(fileName) {
	case recurringTokenIgnore:
		// Currently moving inside a ignore block, skipping line...
		return

	case recurringTokenForeach:
		// TODO: Implementation.
		return

	default:
		// No prefix found, try to resolve variables and functions if there are any.
		var ret []byte
		ret, err = i.resolve(fileName, line, nil)
		if err != nil {
			return
		}
		ret = append(ret, append(indentParent, indentLine...)...)
		_, err = buf.Write(append(ret, i.lineEnding...))
		if err != nil {
			return
		}
	}

	return
}

func (i *Interpreter) resolve(fileName string, line []byte, additionalVars []common.Variable) (ret []byte, err error) {
	// TODO: Implementation / refactor.
	return
}

// getLeadingWhitespace returns all leading whitespace characters.
func getLeadingWhitespace(line []byte) (s []byte) {
	for _, r := range line {
		if r != ' ' && r != '\t' {
			break
		}
		s = append(s, r)
	}
	return
}
