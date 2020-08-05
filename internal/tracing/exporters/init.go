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
func NewExporter(exportType int) (trace.SpanSyncer, error) {
	switch exportType {
	case IoWriter:
		if exporter, err := io_writer.NewExporter(); err != nil {
			log.Fatal(err)
		} else {
			return exporter, nil
		}

	case UnitTest:
		if exporter, err := unit_test.NewExporter(unit_test.Options{}); err != nil {
			log.Fatal(err)
		} else {
			return exporter, nil
		}

	case Production:
		if exporter, err := production.NewExporter(); err != nil {
			log.Fatal(err)
		} else {
			return exporter, nil
		}
	}

	return nil, errors.InvalidArgumentError("exportType")
}
