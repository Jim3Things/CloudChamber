package common

import (
	"time"
)

// CompleteWithin ensures that a channel either produces a result within the
// specified time, or it times out.
func CompleteWithin(ch <-chan bool, delay time.Duration) bool {
	select {
	case <-ch:
		return true
	case <-time.After(delay):
		return false
	}
}

// DoNotCompleteWithin ensures that the channel does not produce a result
// before the specified time.  When this function returns with success no
// value will have been read from the supplied channel.
func DoNotCompleteWithin(ch <-chan bool, delay time.Duration) bool {
	select {
	case <-ch:
		return false

	case <-time.After(delay):
		return true
	}
}
