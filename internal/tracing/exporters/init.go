package exporters

import (
	"log"

	"go.opentelemetry.io/otel/sdk/export/trace"
	"golang.org/x/crypto/openpgp/errors"

	"github.com/Jim3Things/CloudChamber/internal/tracing/exporters/io_writer"
	"github.com/Jim3Things/CloudChamber/internal/tracing/exporters/production"
	"github.com/Jim3Things/CloudChamber/internal/tracing/exporters/unit_test"
)

// NewExporter creates a trace exporter instance based on the export type
func NewExporter(exportType int) (exporter trace.SpanSyncer, err error) {
	switch exportType {
	case IoWriter:
		exporter, err = io_writer.NewExporter()

	case UnitTest:
		exporter, err = unit_test.NewExporter()

	case Production:
		exporter, err = production.NewExporter()

	default:
		return nil, errors.InvalidArgumentError("exportType")
	}

	if err != nil {
		log.Fatal(err)
	}

	return exporter, nil
}
