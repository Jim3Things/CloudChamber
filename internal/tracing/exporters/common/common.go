package common

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/sdk/export/trace"

	"github.com/Jim3Things/CloudChamber/internal/tracing"
	"github.com/Jim3Things/CloudChamber/pkg/protos/log"
)

func ExtractEntry(_ context.Context, data *trace.SpanData) log.Entry {
	spanID := data.SpanContext.SpanID.String()
	parentID := fmt.Sprintf("%x", data.ParentSpanID)

	entry := log.Entry{
		Name:     data.Name,
		SpanID:   spanID,
		ParentID: parentID,
		Status:   fmt.Sprintf("%v: %v", data.StatusCode, data.StatusMessage),
	}

	for _, event := range data.MessageEvents {
		item := log.Event{
			Text: event.Name,
		}

		for _, attr := range event.Attributes {
			if attr.Key == tracing.StepperTicksKey {
				item.Tick = attr.Value.AsInt64()
			}

			if attr.Key == tracing.Reason {
				item.Reason = attr.Value.AsString()
			}
		}

		entry.Event = append(entry.Event, &item)
	}

	return entry
}
