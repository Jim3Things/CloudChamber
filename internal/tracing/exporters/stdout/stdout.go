package stdout

import (
	"context"
	"fmt"
	"io"
	"os"

	export "go.opentelemetry.io/otel/sdk/export/trace"

	"github.com/Jim3Things/CloudChamber/internal/tracing/exporters/common"
)

// Options are the options to be used when initializing a stdout export.
type Options struct {
	// PrettyPrint will pretty the json representation of the span,
	// making it print "pretty". Default is false.
	PrettyPrint bool
}

// Exporter is an implementation of trace.Exporter that writes spans to stdout.
type Exporter struct {
	pretty       bool
	outputWriter io.Writer
}

func NewExporter(o Options) (*Exporter, error) {
	return &Exporter{
		pretty:       o.PrettyPrint,
		outputWriter: os.Stdout,
	}, nil
}

// ExportSpan writes a SpanData in json format to stdout.
func (e *Exporter) ExportSpan(ctx context.Context, data *export.SpanData) {
	entry := common.ExtractEntry(ctx, data)
	_, _ = e.outputWriter.Write([]byte(
		fmt.Sprintf(
			"[%s:%s] %s %s:\n%s\n\n",
			entry.GetSpanID(),
			entry.GetParentID(),
			entry.GetStatus(),
			entry.GetName(),
			entry.GetStackTrace())))

	for _, event := range entry.Event {
		if event.GetTick() < 0 {
			_, _ = e.outputWriter.Write([]byte(
				fmt.Sprintf(
					"       : [%s] (%s) %s\n%s\n\n",
					common.SeverityFlag(event.GetSeverity()),
					event.GetName(),
					event.GetText(),
					event.GetStackTrace())))
		} else {
			_, _ = e.outputWriter.Write([]byte(
				fmt.Sprintf(
					"  @%4d: [%s] (%s) %s\n%s\n\n",
					event.GetTick(),
					common.SeverityFlag(event.GetSeverity()),
					event.GetName(),
					event.GetText(),
					event.GetStackTrace())))
		}
	}
}
