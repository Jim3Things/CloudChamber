package exporters

import (
	"context"
	"fmt"
	"strings"

	trace2 "go.opentelemetry.io/otel/api/trace"
	"go.opentelemetry.io/otel/sdk/export/trace"

	"github.com/Jim3Things/CloudChamber/internal/tracing"
	"github.com/Jim3Things/CloudChamber/pkg/protos/log"
)

const (
	tab = "    "
)

// extractEntry transforms the incoming information about an OpenTelemetry span
// into the entry / event structure is understood by Cloud Chamber's tracing
// consumers, including the UI.
func extractEntry(_ context.Context, data *trace.SpanData) *log.Entry {
	spanID := data.SpanContext.SpanID.String()
	traceID := data.SpanContext.TraceID.String()
	parentID := data.ParentSpanID.String()

	entry := &log.Entry{
		Name:           data.Name,
		SpanID:         spanID,
		ParentID:       parentID,
		TraceID:        traceID,
		Infrastructure: data.SpanKind == trace2.SpanKindInternal,
		Status:         fmt.Sprintf("%v: %v", data.StatusCode, data.StatusMessage),
		StackTrace:     "",
		Reason:         "",
	}

	for _, attr := range data.Attributes {
		switch attr.Key {
		case tracing.StackTraceKey:
			entry.StackTrace = attr.Value.AsString()

		case tracing.ReasonKey:
			entry.Reason = attr.Value.AsString()
		}
	}

	for _, event := range data.MessageEvents {
		item := log.Event{
			Text: event.Name,
			Tick: -1,
			EventAction: log.Action_Trace,
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

			case tracing.ChildSpanKey:
				item.SpanId = attr.Value.AsString()

			case tracing.ActionKey:
				item.EventAction = log.Action(attr.Value.AsInt64())
			}
		}

		switch item.EventAction {
		case log.Action_UpdateSpanName:
			entry.Name = item.Text

		case log.Action_UpdateReason:
			entry.Reason = item.Text

		default:
			entry.Event = append(entry.Event, &item)
		}
	}

	return entry
}

// formatEntry produces a string containing the information in the span-level
// data.  This is used by exporters that emit the trace to a text-based stream.
func formatEntry(entry *log.Entry, deferred bool, leader string) string {
	stack := doIndent(entry.GetStackTrace(), tab)

	return doIndent(fmt.Sprintf(
		"[%s:%s]%s%s %s %s(%s):\n%s\n",
		entry.GetSpanID(),
		entry.GetParentID(),
		deferredFlag(deferred),
		infraFlag(entry.GetInfrastructure()),
		entry.GetStatus(),
		entry.GetName(),
		entry.GetReason(),
		stack), leader)
}

// formatEvent produces a string for a single event in a span that contains the
// formatted information about that event.  Also used by exporters that emit
// the trace events to a text-based stream.
func formatEvent(event *log.Event, leader string) string {
	if event.EventAction == log.Action_SpanStart {
		return strings.TrimSuffix(formatSpanStart(event, leader), leader)
	}

	return strings.TrimSuffix(formatNormalEvent(event, leader), leader)
}

// formatSpanStart produces a string for a 'create child span' event
func formatSpanStart(event *log.Event, leader string) string {
	stack := tab + strings.ReplaceAll(event.GetStackTrace(), "\n", "\n"+tab)

	if event.GetTick() < 0 {
		return doIndent(fmt.Sprintf(
			"       : Start Child Span: %s\n%s\n",
			event.GetSpanId(),
			stack), leader)
	}

	return doIndent(fmt.Sprintf(
		"  @%4d: Start Child Span: %s\n%s\n",
		event.GetTick(),
		event.GetSpanId(),
		stack), leader)
}

// formatNormalEvent produces a string for all other events
func formatNormalEvent(event *log.Event, leader string) string {
	stack := tab + strings.ReplaceAll(event.GetStackTrace(), "\n", "\n"+tab)

	if event.GetTick() < 0 {
		return doIndent(fmt.Sprintf(
			"       : [%s] (%s) %s\n%s\n",
			severityFlag(event.GetSeverity()),
			event.GetName(),
			event.GetText(),
			stack), leader)
	}

	return doIndent(fmt.Sprintf(
		"  @%4d: [%s] (%s) %s\n%s\n",
		event.GetTick(),
		severityFlag(event.GetSeverity()),
		event.GetName(),
		event.GetText(),
		stack), leader)
}

// +++ helper functions

func doIndent(s string, indent string) string {
	return strings.TrimSuffix(
		strings.ReplaceAll(indent+s, "\n", "\n"+indent),
		indent)
}

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
