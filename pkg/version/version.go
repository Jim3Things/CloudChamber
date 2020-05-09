// Package version contains the implementation of some routines and data to allow the
// versioning of the generated executables
//
package version

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/api/trace"
)

//go:generate go run generator\generate.go

// ToString is a function which returns the version information formatted as a
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
// multiple line set of values to Stdout, suiltable for responding to a
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

// TraceVersion is a simple function to insert a trace record containing the
// version information into the trace stream.
//
func Trace() {
	tr := global.TraceProvider().Tracer("")

	_ = tr.WithSpan(context.Background(), "TraceVersion", func(ctx context.Context) (err error) {
		span := trace.SpanFromContext(ctx)
		span.AddEvent(ctx, toString())
		return nil
	})
}
