package io_writer

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"

	export "go.opentelemetry.io/otel/sdk/export/trace"

	"github.com/Jim3Things/CloudChamber/internal/tracing/exporters/common"
	"github.com/Jim3Things/CloudChamber/pkg/protos/log"
)

// Options are the options to be used when initializing a io_writer export.
type Options struct {
}

// Exporter is an implementation of trace.Exporter that writes spans to io_writer.
type Exporter struct {
}

var (
	// Control access to the common entries here
	mutex = sync.Mutex{}

	// Active test context, needed to emit any trace entries, or nil if not
	// currently in an active unit test
	outputWriter io.Writer = nil

	// Since Cloud Chamber code can produce async activity that occurs outside
	// of an active test context, we have to be able to handle these events.
	// When such events happen we save them here and process them as soon as we
	// see an active test context.
	queue = common.NewDeferrable(0)
)

// SetLogFileWriter establishes the IO writer to use to output the trace entries.
// Any deferred trace entries are written at this time.
func SetLogFileWriter(writer io.Writer) error {
	if outputWriter != nil {
		return errors.New("log file writer already set")
	}

	mutex.Lock()
	defer mutex.Unlock()

	outputWriter = writer

	if outputWriter != nil {
		_ = queue.Flush(context.Background(), func(ctx context.Context, item *log.Entry) error {
			return processOneEntry(item)
		})
	}

	return nil
}

// NewExporter creates a trace exporter that outputs to an IO writer
// specified later
func NewExporter(_ Options) (*Exporter, error) {
	return &Exporter{
	}, nil
}

// ExportSpan writes a SpanData in json format to io_writer.
func (e *Exporter) ExportSpan(ctx context.Context, data *export.SpanData) {
	entry := common.ExtractEntry(ctx, data)

	mutex.Lock()
	defer mutex.Unlock()

	if outputWriter != nil {
		_ = processOneEntry(entry)
	} else {
		_ = queue.Defer(entry)
	}
}

func processOneEntry(entry *log.Entry) error {
	_, _ = outputWriter.Write([]byte(
		fmt.Sprintf(
			"[%s:%s] %s %s:\n%s\n\n",
			entry.GetSpanID(),
			entry.GetParentID(),
			entry.GetStatus(),
			entry.GetName(),
			entry.GetStackTrace())))

	for _, event := range entry.Event {
		if event.GetTick() < 0 {
			_, _ = outputWriter.Write([]byte(
				fmt.Sprintf(
					"       : [%s] (%s) %s\n%s\n\n",
					common.SeverityFlag(event.GetSeverity()),
					event.GetName(),
					event.GetText(),
					event.GetStackTrace())))
		} else {
			_, _ = outputWriter.Write([]byte(
				fmt.Sprintf(
					"  @%4d: [%s] (%s) %s\n%s\n\n",
					event.GetTick(),
					common.SeverityFlag(event.GetSeverity()),
					event.GetName(),
					event.GetText(),
					event.GetStackTrace())))
		}
	}

	return nil
}
