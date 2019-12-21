package exporters

import (
    "log"

    "go.opentelemetry.io/otel/sdk/export/trace"

    "github.com/Jim3Things/CloudChamber/internal/tracing/exporters/stdout"
    "github.com/Jim3Things/CloudChamber/internal/tracing/exporters/unit_test"
)

func NewExporter(exportType int, exporter trace.SpanSyncer, err error) (trace.SpanSyncer, error) {
    switch exportType {
    case StdOut:
        exporter, err = stdout.NewExporter(stdout.Options{PrettyPrint: true})
        if err != nil {
            log.Fatal(err)
        }

    case UnitTest:
        exporter, err = unit_test.NewExporter(unit_test.Options{})
        if err != nil {
            log.Fatal(err)
        }

    case Production:
    }
    return exporter, err
}
