package common

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/sdk/export/trace"

	"github.com/Jim3Things/CloudChamber/internal/tracing"
	"github.com/Jim3Things/CloudChamber/pkg/protos/log"
)

func ExtractEntry(_ context.Context, data *trace.SpanData) log.Entry {
	spanId := data.SpanContext.SpanIDString()
	parentId := fmt.Sprintf("%x", data.ParentSpanID)

	entry := log.Entry{
		Name:     data.Name,
		SpanID:   spanId,
		ParentID: parentId,
		Status:   data.Status.String(),
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
