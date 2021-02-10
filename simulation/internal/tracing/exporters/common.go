package exporters

import (
	"context"
	"fmt"
	"strings"

	trace2 "go.opentelemetry.io/otel/api/trace"
	"go.opentelemetry.io/otel/sdk/export/trace"

	"github.com/Jim3Things/CloudChamber/simulation/internal/tracing"
	"github.com/Jim3Things/CloudChamber/simulation/pkg/protos/log"
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
		StartingLink:   "",
		LinkSpanID:     "",
		LinkTraceID:    "",
	}

	for _, attr := range data.Attributes {
		switch attr.Key {
		case tracing.StackTraceKey:
			entry.StackTrace = attr.Value.AsString()

		case tracing.ReasonKey:
			entry.Reason = attr.Value.AsString()

		case tracing.LinkTagKey:
			entry.StartingLink = attr.Value.AsString()
			if len(data.Links) > 0 {
				entry.LinkSpanID = data.Links[0].SpanID.String()
				entry.LinkTraceID = data.Links[0].TraceID.String()
			}
		}
	}

	for _, event := range data.MessageEvents {
		item := log.Event{
			Text:        event.Name,
			Tick:        -1,
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

			case tracing.LinkTagKey:
				item.LinkId = attr.Value.AsString()

			case tracing.ActionKey:
				item.EventAction = log.Action(attr.Value.AsInt64())

			case tracing.ImpactKey:
				item.Impacted = processImpacts(attr.Value.AsArray())
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
		"[%s:%s]%s%s%s %s %s (%s):\n%s\n",
		entry.GetSpanID(),
		entry.GetParentID(),
		formatLink(entry.GetStartingLink(), entry.GetLinkSpanID(), entry.GetLinkTraceID()),
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
	switch event.EventAction {
	case log.Action_AddLink:
		return strings.TrimSuffix(formatAddLink(event, leader), leader)

	case log.Action_SpanStart:
		return strings.TrimSuffix(formatSpanStart(event, leader), leader)

	default:
		return strings.TrimSuffix(formatNormalEvent(event, leader), leader)
	}
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

func formatAddLink(event *log.Event, leader string) string {
	stack := tab + strings.ReplaceAll(event.GetStackTrace(), "\n", "\n"+tab)

	if event.GetTick() < 0 {
		return doIndent(fmt.Sprintf(
			"       : Add link: %s\n%s\n",
			event.GetLinkId(),
			stack), leader)
	}

	return doIndent(fmt.Sprintf(
		"  @%4d: Add link: %s\n%s\n",
		event.GetTick(),
		event.GetLinkId(),
		stack), leader)
}

// formatNormalEvent produces a string for all other events
func formatNormalEvent(event *log.Event, leader string) string {
	stack := tab + strings.ReplaceAll(event.GetStackTrace(), "\n", "\n"+tab)

	if event.GetTick() < 0 {
		return doIndent(fmt.Sprintf(
			"       : [%s] (%s) %s\n%s%s\n",
			severityFlag(event.GetSeverity()),
			event.GetName(),
			event.GetText(),
			formatModules(event.Impacted),
			stack), leader)
	}

	return doIndent(fmt.Sprintf(
		"  @%4d: [%s] (%s) %s\n%s%s\n",
		event.GetTick(),
		severityFlag(event.GetSeverity()),
		event.GetName(),
		event.GetText(),
		formatModules(event.Impacted),
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

func formatLink(tag string, spanID string, traceID string) string {
	if len(tag) > 0 {
		return fmt.Sprintf("<-[%s:%s@%s] ", spanID, traceID, tag)
	}

	return ""
}

func formatModules(modules []*log.Module) string {
	if len(modules) == 0 {
		return ""
	}

	res := tab + "Impacts:\n"
	for i, module := range modules {
		res = fmt.Sprintf("%s%s%s[%d] %s: %s\n", res, tab, tab, i, module.Impact.String(), module.Name)
	}

	return res
}

func processImpacts(attrs interface{}) []*log.Module {
	var modules []*log.Module
	values := attrs.([]string)

	for _, value := range values {
		tags := strings.Split(value, ":")

		modules = append(modules, &log.Module{
			Impact: decodeImpact(tags[0]),
			Name:   tags[1],
		})
	}

	return modules
}

func decodeImpact(tag string) log.Impact {
	switch tag {
	case tracing.ImpactCreate:
		return log.Impact_Create

	case tracing.ImpactRead:
		return log.Impact_Read

	case tracing.ImpactModify:
		return log.Impact_Modify

	case tracing.ImpactDelete:
		return log.Impact_Delete

	case tracing.ImpactUse:
		return log.Impact_Execute

	default:
		return log.Impact_Invalid
	}
}

// --- helper functions
