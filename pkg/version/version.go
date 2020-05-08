// Package version contains the implementation of some routines and data to allow the
// versioning of the generated executables
//
package version

import "fmt"

//go:generate go run generator\generate.go

// ToString is a function which returns the version information formatted as a
// single line, suitable for dumpting to a log file.
//
func ToString() string {
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
