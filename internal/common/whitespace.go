package common

// GetLeadingWhitespace returns all leading whitespace characters.
func GetLeadingWhitespace(line []byte) (s []byte) {
	for _, r := range line {
		if r != ' ' && r != '\t' {
			return
		}
		s = append(s, r)
	}
	return
}
