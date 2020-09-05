package setup

import (
	"fmt"
	"log"
	"os"
	"strings"

	"go.opentelemetry.io/otel/api/global"
	sdk "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc"

	"github.com/Jim3Things/CloudChamber/internal/tracing/exporters"
	"github.com/Jim3Things/CloudChamber/internal/tracing/exporters/io_writer"
	"github.com/Jim3Things/CloudChamber/internal/tracing/exporters/production"
)

// Init configures one or more OpenTelemetry exporters into our trace provider
func Init(exportType ...int) {

	var options []sdk.ProviderOption

	options = append(options, sdk.WithConfig(sdk.Config{DefaultSampler: sdk.AlwaysSample()}))

	for _, item := range exportType {
		exporter, err := exporters.NewExporter(item)
		if err != nil {
			log.Fatal(err)
		}

		options = append(options, sdk.WithSyncer(exporter))
	}

	tp, err := sdk.NewProvider(options...)
	if err != nil {
		log.Fatal(err)
	}

	global.SetTraceProvider(tp)
}

// SetFileWriter sets up the IO writer for the trace exporter that outputs to
// a file.  It defaults to stdout if no file name is specified.
func SetFileWriter(name string) error {
	if name == "" || strings.EqualFold(name, "stdout") {
		// If no trace file specified, use stdout
		return io_writer.SetLogFileWriter(os.Stdout)
	} else {
		writer, err := os.OpenFile(
			name,
			os.O_APPEND | os.O_CREATE | os.O_WRONLY,
			0644)

		if err != nil {
			return fmt.Errorf("error creating trace file (%q), err=%v", name, err)
		}

		return io_writer.SetLogFileWriter(writer)
	}
}

// SetEndpoint supplies the endpoint to the trace sink service for the
// 'production' trace exporter variant
func SetEndpoint(endpoint string) error {
	return production.SetEndpoint(endpoint, grpc.WithInsecure())
}
