package interpreter

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"

	"github.com/xiroxasx/fastplate/internal/common"
)

const (
	variableForeachRegister = "FOREACH"
)

type variableGetter interface {
	varsLookupGlobal() []common.Variable
	varsLookupGlobalFile(name string) []common.Variable
	varLookupLocal(register, name string) common.Variable
}

func (s *state) foreachStart(pd *preprocessorDirective) (err error) {
	if len(pd.args) < 1 {
		return errors.New("at least 1 arg expected")
	}

	s.foreachBuff.state = append(s.foreachBuff.state, foreachBufferState{
		args:   pd.args,
		indent: pd.indent,
		buf:    &bytes.Buffer{},
		varRegistry: variableRegistry{
			entries: make(map[string]vars, 0),
			Mutex:   &sync.Mutex{},
		},
	})
	s.foreachBuff.cursor++
	return
}

func (s *state) foreachEnd(pd *preprocessorDirective) (err error) {
	defer func() {
		if err == nil && s.foreachBuff.cursor > 0 {
			s.foreachBuff.cursor--
		}
	}()

	return s.foreachEvaluate(pd.buf, pd.indent)
}

//
// State.
//

type foreachBufferStack struct {
	// cursor is the index for the corresponding foreachBuffer state.
	cursor uint
	// Each "foreach start" directive creates a new state, since foreach statements can be nested.
	state     []foreachBufferState
	fileName  string
	varGetter variableGetter
}

func newForeachBufferStack(fileName string) foreachBufferStack {
	return foreachBufferStack{
		fileName: fileName,
		state:    make([]foreachBufferState, 0),
	}
}

type foreachBufferState struct {
	args        [][]byte
	indent      []byte
	buf         *bytes.Buffer
	varRegistry variableRegistry
}

func (f *foreachBufferStack) writeToBuffer(line []byte) (err error) {
	_, err = f.currentBuffer().Write(line)
	return
}

func (f *foreachBufferStack) currentBuffer() *bytes.Buffer {
	return f.state[f.cursor].buf
}

func (f *foreachBufferState) varLookupForeach(name string) (v common.Variable) {
	return varLookupRegistry(&f.varRegistry, variableForeachRegister, name)
}

func (f *foreachBufferStack) isActive() bool {
	return f.cursor > 0
}

func (s *state) foreachEvaluate(dst io.Writer, indent []byte) (err error) {
	var foreachLineNum int
	defer func() {
		if err != nil {
			err = fmt.Errorf("foreach evaluation: %v (foreach line %d)", err, foreachLineNum)
		}
	}()

	buf := s.foreachBuff.currentBuffer()
	content := buf.Bytes()
	lines := bytes.Split(content, []byte(lineEnding))

	lineIterator := func(vars ...common.Variable) (err error) {
		for i, line := range lines {
			var resolved []byte
			resolved, err = s.resolve(s.foreachBuff.fileName, line, append(vars, common.NewVar("index", strconv.Itoa(i))))
			if err != nil {
				foreachLineNum = i
				return
			}
			_, err = dst.Write(append(indent, resolved...))
			if err != nil {
				foreachLineNum = i
				return
			}
		}

		return
	}

	vars, rangeNum := s.foreachBuff.evaluationVars()
	if rangeNum > -1 {
		for i := 0; i < rangeNum; i++ {
			err = lineIterator()
			if err != nil {
				return
			}
		}
		return
	}

	for _, v := range vars {
		err = lineIterator(common.NewVar("value", v.Value()))
		if err != nil {
			return
		}
	}

	return
}

// evaluationVars returns either the variables or range that should be used for the foreach process.
// rangeNum is greater than -1 if the current stack has exact one arg (value or variable), which we then try to parse to an integer.
// rangeNum is 0 if the current stack has exact one arg (value or variable), which we tried to parse but failed to do so.
// Otherwise we try to get any variable that can be found in eihter the foreach, local or the global registry.
func (f *foreachBufferStack) evaluationVars() (vs []common.Variable, rangeNum int) {
	stack := f.state[f.cursor]
	variables := make([]common.Variable, 0)
	argsLen := len(stack.args)
	rangeNum = -1
	for _, arg := range stack.args {
		argStr := string(arg)

		if argsLen == 1 {
			// Looks like the user wants to range over the amount specified in the arg.
			rangeNum = 0
			rng, err := strconv.Atoi(argStr)
			if err == nil {
				rangeNum = rng
			}
			return
		}

		vars := f.varLookupRecursive(argStr)
		if len(vars) == 1 && argsLen == 1 {
			// Looks like the user wants to range over the amount specified in a variable.
			rangeNum = 0
			rng, err := strconv.Atoi(argStr)
			if err == nil && argsLen == 1 {
				// User wants to range over the amount specified.
				rangeNum = rng
			}
			return
		}

		if len(vars) > 0 {
			variables = append(variables, vars...)
		}
	}

	return variables, -1
}

func (f *foreachBufferStack) varLookupRecursive(name string) (_ []common.Variable) {
	for _, state := range f.state {
		vs := state.varLookupForeach(name)
		if vs.Name() == "" {
			continue
		}

		return []common.Variable{vs}
	}

	// Foreach variable not found, try getting a local variable.
	vs := f.varGetter.varLookupLocal(f.fileName, name)
	if vs.Name() != "" {
		return []common.Variable{vs}
	}

	// Try to find it against a global var file name.
	vArgs := strings.Split(name, variableGlobalKeyFile)
	if len(vArgs) > 1 {
		vs := f.varGetter.varsLookupGlobalFile(vArgs[1])
		return vs
	}

	// Last resort, try global vars.
	if name == variableGlobalKey {
		return f.varGetter.varsLookupGlobal()
	}
	gVars := f.varGetter.varsLookupGlobal()
	for _, v := range gVars {
		if v.Name() == name {
			return []common.Variable{v}
		}
	}

	return
}

func (f *foreachBufferStack) varLookupForeach(name string) (_ common.Variable, stateIndex int) {
	for i, state := range f.state {
		vs := state.varLookupForeach(name)
		if vs.Name() == "" {
			continue
		}

		return vs, i
	}

	return
}
