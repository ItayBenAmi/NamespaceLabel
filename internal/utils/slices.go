package internal

// Contains returns true if a string is contained within a slice, and false otherwise.
func Contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}
