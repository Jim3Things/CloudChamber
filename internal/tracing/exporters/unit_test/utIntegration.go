package unit_test

import (
	"context"
	"fmt"
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

	testContext.Log(fmt.Sprintf("[%s:%s] %s %s:", entry.GetSpanID(), entry.GetParentID(), entry.GetStatus(), entry.GetName()))
	if entry.Event != nil {
		for _, event := range entry.Event {
			tick := event.GetTick()
			if tick >= 0 {
				testContext.Log(fmt.Sprintf("  @%d: %s (%s)", event.GetTick(), event.GetText(), event.GetReason()))
			} else {
				testContext.Log(fmt.Sprintf("     : %s (%s)", event.GetText(), event.GetReason()))
			}
		}
	}
}
