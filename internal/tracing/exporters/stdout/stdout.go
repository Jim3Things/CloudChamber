package stdout

import (
    "context"
    "encoding/json"
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
    var jsonSpan []byte
    var err error

    entry := common.ExtractEntry(ctx, data)

    if e.pretty {
        jsonSpan, err = json.MarshalIndent(entry, "", "\t")
    } else {
        jsonSpan, err = json.Marshal(entry)
    }
    if err != nil {
        // ignore writer failures for now
        _, _ = e.outputWriter.Write([]byte("Error converting spanData to json: " + err.Error()))
        return
    }

    // ignore writer failures for now
    _, _ = e.outputWriter.Write(append(jsonSpan, byte('\n')))
}
