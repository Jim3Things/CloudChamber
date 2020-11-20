package messages

import (
	"context"

	"go.opentelemetry.io/otel/api/trace"

	"github.com/Jim3Things/CloudChamber/internal/sm"
	"github.com/Jim3Things/CloudChamber/internal/tracing"
)

// EnvelopeState holds the outer context wrapper state for an incoming request to
// the state machine.  It should be embedded in a message that the state
// machine will actually process.
type EnvelopeState struct {
	// Ch is the channel that the completion response is to be sent over.
	Ch chan *sm.Response

	// Span is the tracing span context that is logically associated with this
	// request.
	Span trace.SpanContext

	// Link is the link tag to associate the caller's span with the execution
	// span.
	Link string

	Tag int
}

func (e *EnvelopeState) GetCh() chan *sm.Response {
	return e.Ch
}

func (e *EnvelopeState) GetSpanContext() trace.SpanContext {
	return e.Span
}

func (e *EnvelopeState) GetLinkID() string {
	return e.Link
}

func (e *EnvelopeState) GetTag() int {
	return e.Tag
}

func (e *EnvelopeState) Initialize(ctx context.Context, ch chan *sm.Response) {
	span := trace.SpanFromContext(ctx)
	tag, ok := tracing.GetAndMarkLink(span)
	if ok {
		tracing.AddLink(ctx, tag)
	}

	e.Span = span.SpanContext()
	e.Link = tag

	e.Ch = ch
}

func (e *EnvelopeState) InitializeNoLink(ch chan *sm.Response) {
	e.Span = trace.SpanContext{}
	e.Link = ""
	e.Ch = ch
}
