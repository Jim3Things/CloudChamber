package unit_test

import (
	"context"
	"testing"

	export "go.opentelemetry.io/otel/sdk/export/trace"

	"github.com/Jim3Things/CloudChamber/internal/tracing/exporters/common"
	"github.com/Jim3Things/CloudChamber/pkg/protos/log"
)

// Options are the options to be used when initializing a unit test export.
type Options struct {
}

// Exporter is an implementation of trace.Exporter that writes spans to the unit test logger.
type Exporter struct {
}

var testContext *testing.T

// Since the actor system can produce async activity that occurs outside of an
// active test context, we have to be able to handle these events.  When they
// happen we save them into this array and process them as soon as we see an
// active test context.
var savedEntries []*log.Entry

func NewExporter(_ Options) (*Exporter, error) {
	return &Exporter{}, nil
}

// Set the testing context hook, or clear it, if the reference is nil.
func SetTesting(item *testing.T) {
	testContext = item

	if testContext != nil {
		flushSaved()
	}
}

// Export a span to the output channel
func (e *Exporter) ExportSpan(ctx context.Context, data *export.SpanData) {
	if testContext != nil {
		flushSaved()
		processOneEntry(common.ExtractEntry(ctx, data))
	} else {
		savedEntries = append(savedEntries, common.ExtractEntry(ctx, data))
	}
}

// Flush all saved (out of band) entries into the trace log
func flushSaved() {
	for _, item := range savedEntries {
		processOneEntry(item)
	}

	savedEntries = []*log.Entry{}
}

// Send one entry to the output channel
func processOneEntry(entry *log.Entry) {
	testContext.Logf("[%s:%s] %s %s:\n%s", entry.GetSpanID(), entry.GetParentID(), entry.GetStatus(), entry.GetName(), entry.GetStackTrace())

	for _, event := range entry.Event {
		if event.GetTick() < 0 {
			testContext.Logf("       : [%s] (%s) %s\n%s", severityFlag(event.GetSeverity()), event.GetName(), event.GetText(), event.GetStackTrace())
		} else {
			testContext.Logf("  @%4d: [%s] (%s) %s\n%s", event.GetTick(), severityFlag(event.GetSeverity()), event.GetName(), event.GetText(), event.GetStackTrace())
		}
	}
}

func severityFlag(severity log.Severity) string {
	var severityToText = map[log.Severity]string {
		log.Severity_Debug: "D",
		log.Severity_Reason: "R",
		log.Severity_Info: "I",
		log.Severity_Warning: "W",
		log.Severity_Error: "E",
		log.Severity_Fatal: "F",
	}

	t, ok := severityToText[severity]
	if !ok {
		t = "X"
	}

	return t
}
