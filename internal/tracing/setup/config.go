package setup

import (
	"log"

	"go.opentelemetry.io/otel/api/global"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"github.com/Jim3Things/CloudChamber/internal/tracing/exporters"
)

// Init configures one or more OpenTelemetry exporters into our trace provider
func Init(exportType ...int) {

	var options []sdktrace.ProviderOption
	options = append(options, sdktrace.WithConfig(sdktrace.Config{DefaultSampler: sdktrace.AlwaysSample()}))

	for _, item := range exportType {
		exporter, err := exporters.NewExporter(item)
		if err != nil {
			log.Fatal(err)
		}

		options = append(options, sdktrace.WithSyncer(exporter))
	}


	tp, err := sdktrace.NewProvider(options...)
	if err != nil {
		log.Fatal(err)
	}

	global.SetTraceProvider(tp)
}
