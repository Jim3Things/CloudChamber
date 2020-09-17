package exporters

import (
	"errors"
	"fmt"
)

var (
	// ErrAlreadyOpen indicates that an attempt was made to open an exporter
	// while it was open and active
	ErrAlreadyOpen = errors.New("CloudChamber: exporter is already open")

	// ErrOpenAttrsNil indicates that an Open exporter operation was passed no
	// arguments, but argument values were required.
	ErrOpenAttrsNil = errors.New("CloudChamber: Exporter.Open attributes must not be nil")
)

// ErrInvalidOpenAttrsType indicates that the argument passed to the
// Exporter.Open call used a type that was not expected or supported.
type ErrInvalidOpenAttrsType struct {
	expected string
	actual   string
}

func (e ErrInvalidOpenAttrsType) Error() string {
	return fmt.Sprintf(
		"CloudChamber: exporter open expected type %q, received type %q",
		e.expected,
		e.actual)
}
