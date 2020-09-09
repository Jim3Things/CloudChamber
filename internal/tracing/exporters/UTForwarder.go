package exporters

import (
	"reflect"
	"testing"

	pbl "github.com/Jim3Things/CloudChamber/pkg/protos/log"
)

// UTForwarder implements span entry forwarding to the test context logger
type UTForwarder struct {

	// t is the active test context, or nil, if we're between tests.
	t *testing.T
}

// NewUTForwarder creates a new instance, with no active test context
func NewUTForwarder() *UTForwarder {
	return &UTForwarder{t: nil}
}

// Open establishes a connection to active test context
func (utf *UTForwarder) Open(attrs interface{}) error {
	if attrs == nil {
		return ErrOpenAttrsNil
	}

	tc, ok := attrs.(*testing.T)

	if !ok {
		return ErrInvalidOpenAttrsType{
			expected: reflect.TypeOf(utf.t).String(),
			actual:   reflect.TypeOf(attrs).String(),
		}
	}

	utf.t = tc
	return nil
}

// Close disconnects from the test context
func (utf *UTForwarder) Close() {
	utf.t = nil
}

// Forward emits the formatted log entry into the test context
func (utf *UTForwarder) Forward(entry *pbl.Entry, early bool) error {
	utf.t.Log(formatEntry(entry, early, ""))

	for _, event := range entry.Event {
		utf.t.Log(formatEvent(event, ""))
	}

	return nil
}
