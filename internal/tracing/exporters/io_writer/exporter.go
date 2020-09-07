package io_writer

import (
	"context"
	"errors"
	"io"
	"sync"

	export "go.opentelemetry.io/otel/sdk/export/trace"

	"github.com/Jim3Things/CloudChamber/internal/tracing/exporters/common"
	"github.com/Jim3Things/CloudChamber/pkg/protos/log"
)

// Note well: This exporter type assumes that the exporter is a singleton.  That is
// sufficient for Cloud Chamber at this time, but is something that will need to be
// cleaned up in a future update.

// Exporter is an implementation of trace.Exporter that writes spans to io_writer.
type Exporter struct {
}

const (
	// noExporterRequested is the starting state, indicating that no caller has
	// yet tried to set up this exporter type
	noExporterRequested = iota

	// notInitialized indicates that an exporter has been requested, but the
	// target IO writer has not yet been established
	notInitialized

	// active is the fully operational state - there is an active exporter and
	// it has a known IO writer to send formatted trace output
	active
)

var (
	// alreadyInitialized indicates that an attempt has been made to change the
	// active IO writer.  This is not supported.
	alreadyInitialized = errors.New("trace writer is already set")

	// mutex controls access to the common entries here
	mutex = sync.Mutex{}

	// state contains the current execution state as defined above
	state = noExporterRequested

	// outputWriter is the established IO writer, needed to emit any trace
	// entries, or nil if not yet configured
	outputWriter io.Writer = nil

	// queue holds the trace entries that have arrived prior to establishing
	// the IO writer to use.
	queue = common.NewDeferrable(0)

	spanSet = newSpans()
)

// SetLogFileWriter establishes the IO writer to use to output the trace entries.
// Any deferred trace entries are written at this time.
func SetLogFileWriter(writer io.Writer) error {
	if state == noExporterRequested {
		// This trace exporter has not been chosen, ignore this call
		return nil
	}

	if state != notInitialized {
		return alreadyInitialized
	}

	mutex.Lock()
	defer mutex.Unlock()

	outputWriter = writer
	state = active

	if outputWriter != nil {
		_ = queue.Flush(context.Background(), func(ctx context.Context, item *log.Entry) error {
			return processOneEntry(item, true)
		})
	}

	return nil
}

// NewExporter creates a trace exporter that outputs to an IO writer
// specified later
func NewExporter() (*Exporter, error) {
	mutex.Lock()
	defer mutex.Unlock()

	if state == noExporterRequested {
		state = notInitialized
	}

	return &Exporter{}, nil
}

// ExportSpan writes a SpanData in json format to io_writer.
func (e *Exporter) ExportSpan(ctx context.Context, data *export.SpanData) {
	entry := common.ExtractEntry(ctx, data)

	mutex.Lock()
	defer mutex.Unlock()

	if state == active {
		_ = processOneEntry(entry, false)
	} else {
		_ = queue.Defer(entry)
	}
}

func processOneEntry(entry *log.Entry, _ bool) error {
	spanSet.add(entry, outputWriter)

	return nil
}
