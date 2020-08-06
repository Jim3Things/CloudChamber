package unit_test

import (
	"context"
	"sync"
	"testing"

	export "go.opentelemetry.io/otel/sdk/export/trace"

	"github.com/Jim3Things/CloudChamber/internal/tracing/exporters/common"
	"github.com/Jim3Things/CloudChamber/pkg/protos/log"
)

// Exporter is an implementation of trace.Exporter that writes spans to the
// unit test logger.
type Exporter struct {
}

var (
	// Control access to the common entries here
	mutex = sync.Mutex{}

	// Active test context, needed to emit any trace entries, or nil if not
	// currently in an active unit test
	testContext *testing.T = nil

	// Since Cloud Chamber code can produce async activity that occurs outside
	// of an active test context, we have to be able to handle these events.
	// When such events happen we save them here and process them as soon as we
	// see an active test context.
	queue = common.NewDeferrable(0)
)

// NewExporter creates a new unit test Exporter instance
func NewExporter() (*Exporter, error) {
	return &Exporter{}, nil
}

// SetTesting is a function that stores the active testing context, or nil if
// between unit tests.
func SetTesting(item *testing.T) {
	mutex.Lock()
	defer mutex.Unlock()

	testContext = item

	if testContext != nil {
		flushSaved(context.Background())
	}
}

// Export a span to the output channel
func (e *Exporter) ExportSpan(ctx context.Context, data *export.SpanData) {
	entry := common.ExtractEntry(ctx, data)

	mutex.Lock()
	defer mutex.Unlock()

	if testContext != nil {
		flushSaved(ctx)
		processOneEntry(entry, false)
	} else {
		_ = queue.Defer(entry)
	}
}

// Flush all saved (out of band) entries into the trace log
func flushSaved(ctx context.Context) {
	_ = queue.Flush(ctx, func(ctx context.Context, item *log.Entry) error {
		processOneEntry(item, true)
		return nil
	})
}

// Send one entry to the output channel
func processOneEntry(entry *log.Entry, deferred bool) {
	if deferred {
		testContext.Logf("[%s:%s] %s (deferred) %s:\n%s", entry.GetSpanID(), entry.GetParentID(), entry.GetStatus(), entry.GetName(), entry.GetStackTrace())
	} else {
		testContext.Logf("[%s:%s] %s %s:\n%s", entry.GetSpanID(), entry.GetParentID(), entry.GetStatus(), entry.GetName(), entry.GetStackTrace())
	}

	for _, event := range entry.Event {
		if event.GetTick() < 0 {
			testContext.Logf("       : [%s] (%s) %s\n%s", common.SeverityFlag(event.GetSeverity()), event.GetName(), event.GetText(), event.GetStackTrace())
		} else {
			testContext.Logf("  @%4d: [%s] (%s) %s\n%s", event.GetTick(), common.SeverityFlag(event.GetSeverity()), event.GetName(), event.GetText(), event.GetStackTrace())
		}
	}
}
