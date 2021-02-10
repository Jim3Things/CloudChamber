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

// CompleteWithinInterface ensures that a channel either produces a result
// within the specified time, or it times out.
func CompleteWithinInterface(ch <-chan interface{}, delay time.Duration) interface{} {
	select {
	case res := <-ch:
		return res
	case <-time.After(delay):
		return nil
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

// DoNotCompleteWithinInterface ensures that the channel does not produce a result
// before the specified time.  When this function returns with success no
// value will have been read from the supplied channel.
func DoNotCompleteWithinInterface(ch <-chan interface{}, delay time.Duration) interface{} {
	select {
	case res := <-ch:
		return res

	case <-time.After(delay):
		return nil
	}
}
