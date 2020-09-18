package log

import (
	"errors"
	"fmt"

	"go.opentelemetry.io/otel/api/trace"

	"github.com/Jim3Things/CloudChamber/pkg/protos/common"
)

// Validate verifies that the various IDs are correctly structured.
func (x *Entry) Validate(prefix string) error {
	// Span ID must be legal
	spanID, err := trace.SpanIDFromHex(x.SpanID)
	if err != nil || !spanID.IsValid() {
		return common.ErrInvalidID{
			Field: fmt.Sprintf("%sSpanID", prefix),
			Type: "span",
			ID:    x.SpanID,
		}
	}

	// Parent ID must be legal (it can be all zeroes, though)
	if _, err = trace.SpanIDFromHex(x.ParentID);
		err != nil && !errors.Is(err, trace.ErrNilSpanID) {
		return common.ErrInvalidID{
			Field: fmt.Sprintf("%sParentID", prefix),
			Type: "parent",
			ID:    x.ParentID,
		}
	}

	// Trace ID must be legal
	traceID, err := trace.IDFromHex(x.TraceID)
	if err != nil || !traceID.IsValid() {
		return common.ErrInvalidID{
			Field: fmt.Sprintf("%sTraceID", prefix),
			Type: "trace",
			ID:    x.TraceID,
		}
	}

	return nil
}
