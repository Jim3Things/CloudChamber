package utilities

import (
	"time"
)

// WaitForStateChange is a unit test utility function that checks for a
// comparison succeeding within the specified number of seconds.
func WaitForStateChange(maxDelay int, compare func() bool) bool {
	for i := 0; i < maxDelay * 100; i++ {
		if compare() {
			return true
		}
		time.Sleep(time.Duration(10) * time.Millisecond)
	}

	return compare()
}

