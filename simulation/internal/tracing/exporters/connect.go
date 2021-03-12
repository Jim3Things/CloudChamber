package exporters

import (
	"log"

	"go.opentelemetry.io/otel/api/global"
	sdk "go.opentelemetry.io/otel/sdk/trace"
)

// ConnectToProvider configures one or more OpenTelemetry exporters into our trace provider
func ConnectToProvider(exporters ...*Exporter) {

	options := []sdk.ProviderOption{
		sdk.WithConfig(sdk.Config{
			DefaultSampler: sdk.AlwaysSample(),
			MaxEventsPerSpan: 1024,
		}),
	}

	for _, exporter := range exporters {
		options = append(options, sdk.WithSyncer(exporter))
	}

	tp, err := sdk.NewProvider(options...)
	if err != nil {
		log.Fatal(err)
	}

	global.SetTraceProvider(tp)
}
