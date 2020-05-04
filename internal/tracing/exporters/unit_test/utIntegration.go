package unit_test

import (
	"context"
	"testing"

	export "go.opentelemetry.io/otel/sdk/export/trace"

	"github.com/Jim3Things/CloudChamber/internal/tracing/exporters/common"
)

// Options are the options to be used when initializing a unit test export.
type Options struct {
}

// Exporter is an implementation of trace.Exporter that writes spans to the unit test logger.
type Exporter struct {
}

var testContext *testing.T

func NewExporter(_ Options) (*Exporter, error) {
	return &Exporter{}, nil
}

func SetTesting(item *testing.T) {
	testContext = item
}

func (e *Exporter) ExportSpan(ctx context.Context, data *export.SpanData) {
	entry := common.ExtractEntry(ctx, data)

	spanID := entry.GetSpanID()
	parentID := entry.GetParentID()
	status := entry.GetStatus()
	name := entry.GetName()
	stack := entry.GetStackTrace()

	testContext.Logf("[%s:%s] %s %s:%s", spanID, parentID, status, name, stack)
	if entry.Event != nil {
		for _, event := range entry.Event {
			tick := event.GetTick()
			if tick >= 0 {
				testContext.Logf("  @%d: %s (%s)%s", event.GetTick(), event.GetText(), event.GetReason(), event.GetStackTrace())
			} else {
				testContext.Logf("     : %s (%s)%s", event.GetText(), event.GetReason(), event.GetStackTrace())
			}
		}
	}
}
