package importer

const (
	ignoreStart = "start"
	ignoreEnd   = "end"
)

func (i *Importer) ignore(filename, args string) {
	i.state.Lock()
	defer i.state.Unlock()

	switch args {
	case ignoreStart:
		i.state.ignoreIndex[filename] = 1
	case ignoreEnd:
		i.state.ignoreIndex[filename] = 0
	}
	return
}
