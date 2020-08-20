// This module contains support functions for tracking trace spans associated
// with an actor.
//
// The core issue is that the spans are created by the incoming middleware, but
// neither the span or the context is passed directly to the actor when it is
// invoked.  Therefore, this module maintains the spans to actor mapping so that
// the outer span can be retrieved by the actor itself.

package server

import (
	"sync"

	"github.com/AsynkronIT/protoactor-go/actor"
	"go.opentelemetry.io/otel/api/trace"
)

var spans = sync.Map{}

// Get a span based on an actor ID
func GetSpan(pid *actor.PID) trace.Span {
	value, ok := spans.Load(pid)
	if !ok {
		return nil
	}
	return value.(trace.Span)
}

// Remove an actor ID to span association
func ClearSpan(pid *actor.PID) {
	spans.Delete(pid)
}

// Establish an actor ID to span association
func SetSpan(pid *actor.PID, span trace.Span) {
	spans.Store(pid, span)
}
