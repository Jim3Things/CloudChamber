package sm

import (
	"context"

	"go.opentelemetry.io/otel/api/trace"

	"github.com/Jim3Things/CloudChamber/internal/tracing"
)

// EnvelopeState holds the outer context wrapper state for an incoming request to
// the state machine.  It should be embedded in a message that the state
// machine will actually process.
type EnvelopeState struct {
	// Response is the channel that the completion response is to be sent over.
	ch chan *Response

	// span is the tracing span context that is logically associated with this
	// request.
	span trace.SpanContext

	// link is the locally unique ID to associate the caller's span with the
	// execution span.
	link string

	// tag is an identifying 'type' marker for the body of the message.
	tag int
}

func (e *EnvelopeState) Ch() chan *Response {
	return e.ch
}

func (e *EnvelopeState) SpanContext() trace.SpanContext {
	return e.span
}

func (e *EnvelopeState) LinkID() string {
	return e.link
}

func (e *EnvelopeState) Tag() int {
	return e.tag
}

func (e *EnvelopeState) Initialize(ctx context.Context, tag int, ch chan *Response) {
	span := trace.SpanFromContext(ctx)
	linkID, ok := tracing.GetAndMarkLink(span)
	if ok {
		tracing.AddLink(ctx, linkID)
	}

	e.span = span.SpanContext()
	e.link = linkID

	e.tag = tag

	e.ch = ch
}

func (e *EnvelopeState) InitializeNoLink(tag int, ch chan *Response) {
	e.span = trace.SpanContext{}
	e.link = ""

	e.tag = tag

	e.ch = ch
}
