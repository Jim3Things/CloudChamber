package setup

import (
	"log"

	"go.opentelemetry.io/otel/sdk/export/trace"

	"go.opentelemetry.io/otel/api/global"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"github.com/Jim3Things/CloudChamber/internal/tracing/exporters"
)

// Init configures an OpenTelemetry exporter and trace provider
func Init(exportType int) trace.SpanSyncer {

	var exporter trace.SpanSyncer
	var err error

	exporter, err = exporters.NewExporter(exportType, exporter, err)
	if err != nil {
		log.Fatal(err)
	}

	tp, err := sdktrace.NewProvider(
		sdktrace.WithConfig(sdktrace.Config{DefaultSampler: sdktrace.AlwaysSample()}),
		sdktrace.WithSyncer(exporter),
	)
	if err != nil {
		log.Fatal(err)
	}

	global.SetTraceProvider(tp)

	return exporter
}
