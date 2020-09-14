package exporters

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"

	pbl "github.com/Jim3Things/CloudChamber/pkg/protos/log"
)

// NameToWriter converts a file name, or 'stdout' into an active IO.Writer
func NameToWriter(name string) (io.Writer, error) {
	if name == "" || strings.EqualFold(name, "stdout") {
		// If no trace file specified, use stdout
		return os.Stdout, nil
	} else {
		writer, err := os.OpenFile(
			name,
			os.O_APPEND | os.O_CREATE | os.O_WRONLY,
			0644)

		if err != nil {
			return nil, fmt.Errorf("error creating trace file (%q), err=%v", name, err)
		}

		return writer, nil
	}
}

// IOWForwarder implements a Forwarder to a designated IO writer.  The output
// is formatted into a functional tree, grouped by trace ID.
type IOWForwarder struct {
	writer io.Writer
	spans *ioSpans
}

// NewIOWForwarder creates a closed IOwForwarder instance
func NewIOWForwarder() *IOWForwarder {
	return &IOWForwarder{
		writer: nil,
		spans: newSpans(),
	}
}

// Open connects the forwarder to the supplied IO.Writer
func (iow *IOWForwarder) Open(attrs interface{}) error {
	if attrs == nil {
		return ErrOpenAttrsNil
	}

	writer, ok := attrs.(io.Writer)

	if !ok {
		return ErrInvalidOpenAttrsType{
			expected: reflect.TypeOf(iow.writer).String(),
			actual:   reflect.TypeOf(attrs).String(),
		}
	}

	iow.writer = writer
	return nil
}

// Close detaches the forwarder from the active IO.Writer
func (iow *IOWForwarder) Close() {
	iow.writer = nil
}

// Forward emits the log entries to the IO.Writer channel in functional order,
// grouped by trace ID
func (iow *IOWForwarder) Forward(entry *pbl.Entry, _ bool) error {
	iow.spans.add(entry, iow.writer)

	return nil
}
