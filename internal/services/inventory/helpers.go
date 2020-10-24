package inventory

// aOrB is a simple helper function that returns either the 'a' string if the
// supplied flag is true, or the 'b' string.  This simplifies trace format
// calls by removing inline if-then sequences.
func aOrB(flag bool, a string, b string) string {
	if flag {
		return a
	}

	return b
}
