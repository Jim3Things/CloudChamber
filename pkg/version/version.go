// Package version contains the implementation of some routines and data to allow the
// versioning of the generated executables
//
package version

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/api/trace"

	"github.com/Jim3Things/CloudChamber/internal/tracing"
)

//go:generate go run generator/generate.go

// toString is a function which returns the version information formatted as a
// single line, suitable for dumping to a log file.
//
func toString() string {
	return fmt.Sprintf(
		"Version: %v BuildHost: %v BuildDate: %v Branch: %v (%v \\ %v)",
		BuildVersion,
		BuildHost,
		BuildDate,
		BuildBranch,
		BuildBranchDate,
		BuildBranchHash)
}

// Show is a function which prints the version information formatted as a
// multiple line set of values to Stdout, suitable for responding to a
// version option flag
//
func Show() {
	fmt.Printf(
		"Version: %v\nBuildHost: %v\nBuildDate: %v\nBranch: %v (%v \\ %v)\n",
		BuildVersion,
		BuildHost,
		BuildDate,
		BuildBranch,
		BuildBranchDate,
		BuildBranchHash)
}

// Trace is a simple function to insert the leading trace events that show the
// startup of the service, and the version information associated with it.
//
func Trace() {
	tr := global.TraceProvider().Tracer("")

	_ = tr.WithSpan(context.Background(), tracing.MethodName(1), func(ctx context.Context) (err error) {
		span := trace.SpanFromContext(ctx)

		span.AddEvent(
			ctx,
			fmt.Sprintf(
				"===== Starting %q at %s =====",
				fmt.Sprint(os.Args),
				time.Now().Format(time.RFC1123Z)))

		span.AddEvent(ctx, toString())
		return nil
	})
}
