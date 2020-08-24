package common

import (
	"context"
	"fmt"

	trace2 "go.opentelemetry.io/otel/api/trace"
	"go.opentelemetry.io/otel/sdk/export/trace"

	"github.com/Jim3Things/CloudChamber/internal/tracing"
	"github.com/Jim3Things/CloudChamber/pkg/protos/log"
)

// ExtractEntry transforms the incoming information about an OpenTelemetry span
// into the entry / event structure is understood by Cloud Chamber's tracing
// consumers, including the UI.
func ExtractEntry(_ context.Context, data *trace.SpanData) *log.Entry {
	spanID := data.SpanContext.SpanID.String()
	parentID := data.ParentSpanID.String()

	entry := &log.Entry{
		Name:       data.Name,
		SpanID:     spanID,
		ParentID:   parentID,
		Infrastructure: data.SpanKind == trace2.SpanKindInternal,
		Status:     fmt.Sprintf("%v: %v", data.StatusCode, data.StatusMessage),
		StackTrace: "",
	}

	for _, attr := range data.Attributes {
		switch attr.Key {
		case tracing.StackTraceKey:
			entry.StackTrace = attr.Value.AsString()
		}
	}

	for _, event := range data.MessageEvents {
		item := log.Event{
			Text: event.Name,
			Tick: -1,
		}

		for _, attr := range event.Attributes {
			switch attr.Key {
			case tracing.StepperTicksKey:
				item.Tick = attr.Value.AsInt64()

			case tracing.SeverityKey:
				item.Severity = log.Severity(attr.Value.AsInt64())

			case tracing.StackTraceKey:
				item.StackTrace = attr.Value.AsString()

			case tracing.MessageTextKey:
				item.Text = attr.Value.AsString()
			}

			if attr.Key == tracing.StackTraceKey {
				item.StackTrace = attr.Value.AsString()
			}
		}

		entry.Event = append(entry.Event, &item)
	}

	return entry
}

// FormatEntry produces a string containing the information in the span-level
// data.  This is used by exporters that emit the trace to a text-based stream.
func FormatEntry(entry *log.Entry, deferred bool) string {
	return fmt.Sprintf(
		"[%s:%s]%s%s %s %s:\n%s",
		entry.GetSpanID(),
		entry.GetParentID(),
		deferredFlag(deferred),
		infraFlag(entry.GetInfrastructure()),
		entry.GetStatus(),
		entry.GetName(),
		entry.GetStackTrace())
}

// FormatEvent produces a string for a single event in a span that contains the
// formatted information about that event.  Also used by exporters that
func FormatEvent(event *log.Event) string {
	if event.GetTick() < 0 {
		return fmt.Sprintf(
			"       : [%s] (%s) %s\n%s",
			severityFlag(event.GetSeverity()),
			event.GetName(),
			event.GetText(),
			event.GetStackTrace())
	}

	return fmt.Sprintf(
		"  @%4d: [%s] (%s) %s\n%s",
		event.GetTick(),
		severityFlag(event.GetSeverity()),
		event.GetName(),
		event.GetText(),
		event.GetStackTrace())
}

// +++ helper functions that format specific fields

func severityFlag(severity log.Severity) string {
	var severityToText = map[log.Severity]string{
		log.Severity_Debug:   "D",
		log.Severity_Reason:  "R",
		log.Severity_Info:    "I",
		log.Severity_Warning: "W",
		log.Severity_Error:   "E",
		log.Severity_Fatal:   "F",
	}

	t, ok := severityToText[severity]
	if !ok {
		t = "X"
	}

	return t
}

func infraFlag(value bool) string {
	if value {
		return " (Infra)"
	}

	return ""
}

func deferredFlag(value bool) string {
	if value {
		return " (deferred)"
	}

	return ""
}

// --- helper functions
