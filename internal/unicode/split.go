package unicode

// SplitString splits a string after `n` unicode chartacters
func SplitString(s string, n int) (start, end string) {
	i := 0
	for j := range s {
		if i == n {
			return s[:j], s[j:]
		}
		i++
	}
	return s, ""
}
