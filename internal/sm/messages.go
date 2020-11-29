package sm

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/api/trace"
)

// Envelope is the standard header interface to messages to a state machine.
type Envelope interface {

	// Initialize sets up any linked span information, and the channel to use
	// for any completion responses.
	Initialize(ctx context.Context, tag int, ch chan *Response)

	// GetCh returns the response channel.
	GetCh() chan *Response

	// GetSpanContext returns the source span used in the source's add-link
	// event, and used as the linked-from span when processing this message.
	GetSpanContext() trace.SpanContext

	// GetLinkID returns the unique link id that further decorates the span
	// context above.
	GetLinkID() string

	// GetTag returns the type ID to use when matching this message in the
	// action state table.
	GetTag() int
}

// Response holds the completion response for a processed request, whether it
// was successful or not.
type Response struct {
	// Err holds any completion error code, or nil if the request was
	// successful.
	Err error

	// At contains the simulated time tick when the request completed its
	// processing.
	At int64

	// Msg holds any extended results information, or nil if there either is
	// none, or if an error is returned.
	Msg interface{}
}

// ErrUnexpectedMessage is the standard error when an incoming request arrives in
// a state that is not expecting it.
type ErrUnexpectedMessage struct {
	Msg   string
	State string
}

func (um *ErrUnexpectedMessage) Error() string {
	return fmt.Sprintf("unexpected message %q while in state %q", um.Msg, um.State)
}

// UnexpectedMessageResponse constructs a failure response for the case where
// the incoming request arrives when it is unexpected by the state machine.
func UnexpectedMessageResponse(machine *SimpleSM, occursAt int64, body interface{}) *Response {
	return &Response{
		Err: &ErrUnexpectedMessage{
			Msg:   fmt.Sprintf("%v", body),
			State: machine.CurrentIndex,
		},
		At:  occursAt,
		Msg: nil,
	}
}
