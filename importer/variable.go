package importer

import (
	"strings"
)

// variable parses variable names and values and puts them into state.vars.
func (i *Importer) variable(filename string, args []string) (err error) {
	i.state.Lock()
	defer i.state.Unlock()

	cmd := command{name: args[0]}
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
	i.state.vars[filename] = append(i.state.vars[filename], cmd)
	return
}
