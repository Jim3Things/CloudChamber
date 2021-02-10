package common

// MaxInt64 is a helper function to return the maximum of two int64 values
func MaxInt64(a int64, b int64) int64 {
	if a < b {
		return b
	}

	return a
}
