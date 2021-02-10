package log

import (
    "errors"
    "fmt"

    "go.opentelemetry.io/otel/api/trace"

    errors2 "github.com/Jim3Things/CloudChamber/simulation/pkg/errors"
)

// Validate verifies that the various IDs are correctly structured.
func (x *Entry) Validate(prefix string) error {
	// Span ID must be legal
	if spanID, err := trace.SpanIDFromHex(x.SpanID); err != nil || !spanID.IsValid() {
		return errors2.ErrInvalidID{
			Field: fmt.Sprintf("%sSpanID", prefix),
			Type:  "span",
			ID:    x.SpanID,
		}
	}

	// Parent ID must be legal (it can be all zeroes, though)
	if _, err := trace.SpanIDFromHex(x.ParentID);
		err != nil && !errors.Is(err, trace.ErrNilSpanID) {
		return errors2.ErrInvalidID{
			Field: fmt.Sprintf("%sParentID", prefix),
			Type:  "parent",
			ID:    x.ParentID,
		}
	}

	// Trace ID must be legal
	if traceID, err := trace.IDFromHex(x.TraceID); err != nil || !traceID.IsValid() {
		return errors2.ErrInvalidID{
			Field: fmt.Sprintf("%sTraceID", prefix),
			Type:  "trace",
			ID:    x.TraceID,
		}
	}

	if len(x.StartingLink) != 0 {
		if linkSpanID, err := trace.SpanIDFromHex(x.LinkSpanID); err != nil || !linkSpanID.IsValid() {
			return errors2.ErrInvalidID{
				Field: fmt.Sprintf("%sLinkSpanID", prefix),
				Type:  "span",
				ID:    x.SpanID,
			}
		}

		if  linkTraceID, err := trace.IDFromHex(x.LinkTraceID); err != nil || !linkTraceID.IsValid() {
			return errors2.ErrInvalidID{
				Field: fmt.Sprintf("%sLinkTraceID", prefix),
				Type:  "span",
				ID:    x.SpanID,
			}
		}
	}

	for i, event := range x.Event {
		if err := event.Validate(fmt.Sprintf("%sEvent[%d].", prefix, i)); err != nil {
			return err
		}
	}

	return nil
}

// Validate ensures that the fields in a trace event are self-consistent, that
// tracing events do not have link or child span IDs, that span data rewrites
// have some text to replace the original with, and that the child and link
// events have the target information included.
func (x *Event) Validate(prefix string) error {
	switch x.Severity {
	case Severity_Debug,
		Severity_Reason,
		Severity_Info,
		Severity_Warning,
		Severity_Error,
		Severity_Fatal:
		break

	default:
		return &errors2.ErrInvalidEnum{
			Field:  fmt.Sprintf("%sSeverity", prefix),
			Actual: int64(x.Severity),
		}
	}

	switch x.EventAction {
	case Action_Trace, Action_UpdateSpanName, Action_UpdateReason:
		textLen := int64(len(x.Text))
		if textLen == 0 {
			return &errors2.ErrMinLenString{
				Field:    fmt.Sprintf("%sText", prefix),
				Actual:   textLen,
				Required: 1,
			}
		}

		if err := x.ensureSpanIdEmpty(prefix); err != nil {
			return err
		}

		if err := x.ensureLinkIdEmpty(prefix); err != nil {
			return err
		}

	case Action_SpanStart:
		if spanId, err := trace.SpanIDFromHex(x.SpanId); err != nil || !spanId.IsValid() {
			return errors2.ErrInvalidID{
				Field: fmt.Sprintf("%sSpanID", prefix),
				Type:  "span",
				ID:    x.SpanId,
			}
		}

		if err := x.ensureLinkIdEmpty(prefix); err != nil {
			return err
		}

	case Action_AddLink:
		if err := x.ensureSpanIdEmpty(prefix); err != nil {
			return err
		}

		if len(x.LinkId) == 0 {
			return errors2.ErrMinLenString{
				Field: fmt.Sprintf("%sLinkID", prefix),
				Required: 1,
				Actual:   int64(len(x.LinkId)),
			}
		}

	default:
		return &errors2.ErrInvalidEnum{
			Field:  fmt.Sprintf("%sEventAction", prefix),
			Actual: int64(x.Severity),
		}
	}

	return nil
}

func (x *Event) ensureLinkIdEmpty(prefix string) error {
	if len(x.LinkId) != 0 {
		return &errors2.ErrIDMustBeEmpty{
			Field:  fmt.Sprintf("%sLinkId", prefix),
			Actual: x.LinkId,
		}
	}

	return nil
}

func (x *Event) ensureSpanIdEmpty(prefix string) error {
	if len(x.SpanId) != 0 {
		return &errors2.ErrIDMustBeEmpty{
			Field:  fmt.Sprintf("%sSpanId", prefix),
			Actual: x.SpanId,
		}
	}

	return nil
}
