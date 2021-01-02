package sm

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/api/trace"

	"github.com/Jim3Things/CloudChamber/pkg/errors"
)

// Envelope is the standard header interface to messages to a state machine.
type Envelope interface {

	// Initialize sets up any linked span information, and the channel to use
	// for any completion responses.
	Initialize(ctx context.Context, tag int, ch chan *Response)

	// Ch returns the response channel.
	Ch() chan *Response

	// SpanContext returns the source span used in the source's add-link
	// event, and used as the linked-from span when processing this message.
	SpanContext() trace.SpanContext

	// LinkID returns the unique link id that further decorates the span
	// context above.
	LinkID() string

	// Tag returns the type ID to use when matching this message in the
	// action state table.
	Tag() int
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

// UnexpectedMessageResponse constructs a failure response for the case where
// the incoming request arrives when it is unexpected by the state machine.
func UnexpectedMessageResponse(machine *SM, occursAt int64, body interface{}) *Response {
	return &Response{
		Err: &errors.ErrUnexpectedMessage{
			Msg:   fmt.Sprintf("%v", body),
			State: machine.CurrentIndex,
		},
		At:  occursAt,
		Msg: nil,
	}
}

// FailedResponse constructs a failure response message with the correct time,
// target, and reason.
func FailedResponse(occursAt int64, err error) *Response {
	return &Response{
		Err: err,
		At:  occursAt,
		Msg: nil,
	}
}

// SuccessResponse constructs a success response message with the correct time
// and target.
func SuccessResponse(occursAt int64) *Response {
	return &Response{
		Err: nil,
		At:  occursAt,
		Msg: nil,
	}
}
