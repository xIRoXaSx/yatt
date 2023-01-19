package importer

import (
	"bytes"
	"strings"
)

// variable parses variable names and values and puts them into state.vars.
func (i *Importer) variable(fileName string, args []string) (err error) {
	i.state.Lock()
	defer i.state.Unlock()

	cmd := variable{name: args[0]}
	val := strings.SplitN(args[0], "=", 2)
	if len(val) == 1 {
		// Syntax: x = y, x =y
		val = strings.Split(args[1], "=")
		cmd.value = val[1]
	} else {
		// Syntax: x=y, x= y
		cmd.name = val[0]
		cmd.value = val[1]
	}
	if len(args) > 2 {
		// Skip equal sign.
		ind := 1
		if args[ind] == "=" {
			// Any case except syntax x= y
			if len(args) > ind+1 {
				// Skip next space.
				ind++
			}
		}
		cmd.value += strings.Join(args[ind:], " ")
	}
	i.state.vars[fileName] = append(i.state.vars[fileName], cmd)
	return
}

var (
	templateStart = []byte("{{")
	templateEnd   = []byte("}}")
)

func (i *Importer) resolve(fileName string, line []byte) (ret []byte) {
	ret = line
	begin := bytes.Split(line, templateStart)
	if len(begin) == 1 {
		return
	}

	for _, b := range begin {
		match := bytes.Split(b, templateEnd)
		if len(match) == 1 {
			continue
		}

		for _, m := range match {
			v := i.state.lookupVar(fileName, string(m))
			if v.name == "" {
				continue
			}
			matched := bytes.Join([][]byte{templateStart, []byte(v.name), templateEnd}, []byte{})
			ret = bytes.ReplaceAll(ret, matched, []byte(v.value))
		}
	}
	return
}
