// Package version contains the implementation of some routines and data to allow the
// versioning of the generated executables
//
package version

import (
	"context"
	"fmt"
	"os"
	"time"

	clients "github.com/Jim3Things/CloudChamber/internal/clients/timestamp"
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
	ctx, span := tracing.StartSpan(context.Background(),
		tracing.WithContextValue(clients.OutsideTime))
	defer span.End()

	tracing.Infof(ctx, "===== Starting %q at %s =====", fmt.Sprint(os.Args), time.Now().Format(time.RFC1123Z))

	tracing.Info(ctx, toString())
}
