package exporters

import (
	"log"

	"go.opentelemetry.io/otel/sdk/export/trace"
	"golang.org/x/crypto/openpgp/errors"

	"github.com/Jim3Things/CloudChamber/internal/tracing/exporters/stdout"
	"github.com/Jim3Things/CloudChamber/internal/tracing/exporters/unit_test"
)

func NewExporter(exportType int) (trace.SpanSyncer, error) {
	var err error

	switch exportType {
	case StdOut:
		exporter, err := stdout.NewExporter(stdout.Options{PrettyPrint: true})
		if err != nil {
			log.Fatal(err)
		}
		return exporter, nil

	case UnitTest:
		exporter, err := unit_test.NewExporter(unit_test.Options{})
		if err != nil {
			log.Fatal(err)
		}
		return exporter, nil

	case Production:
		err = errors.InvalidArgumentError("exportType")

	case LocalProduction:
		err = errors.InvalidArgumentError("exportType")
	}

	return nil, err
}
