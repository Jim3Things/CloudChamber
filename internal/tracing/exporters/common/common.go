package common

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/sdk/export/trace"

	"github.com/Jim3Things/CloudChamber/internal/tracing"
	"github.com/Jim3Things/CloudChamber/pkg/protos/log"
)

func ExtractEntry(_ context.Context, data *trace.SpanData) *log.Entry {
	spanID := data.SpanContext.SpanID.String()
	parentID := data.ParentSpanID.String()

	entry := &log.Entry{
		Name:     data.Name,
		SpanID:   spanID,
		ParentID: parentID,
		Status:   fmt.Sprintf("%v: %v", data.StatusCode, data.StatusMessage),
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

func SeverityFlag(severity log.Severity) string {
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

