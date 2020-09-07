package exporters

// This module defines the base Exporter class, and the prototype interface for
// forwarders.  The Exporter implements the common state logic, and is provided
// with a forwarder, which is responsible for sending any trace events to the
// final sink, whether it be a file, another object, or an endpoint.

import (
	"context"
	"log"
	"sync"

	export "go.opentelemetry.io/otel/sdk/export/trace"

	pbl "github.com/Jim3Things/CloudChamber/pkg/protos/log"
)

const (
	// maxBuffer is the number of trace entries that can be pending before
	// a new ExportSpan event will block
	maxBuffer = 100
)

// forwarder defines the interface that a particular trace forwarder
// must implement.
type forwarder interface {
	// Open establishes the target sink to use.  The type is specific to a
	// given forwarder.  The actual connection to the target sink may be
	// deferred, so long as Open is sufficient to ensure that it can be
	// eventually established.
	Open(attrs interface{}) error

	// Close terminates the connection to a target sink.
	Close()

	// Forward sends the supplied entry to the target sink.  This is called
	// in a worker goroutine.  'early' is set to true if the entry arrived
	// prior to the Open operation.
	Forward(entry *pbl.Entry, early bool) error
}

// +++ Internal synchronization messages

// closeMsg indicates that the forwarder should be closed.  It contains a
// response channel to indicate that the closer operation is complete.
type closeMsg struct {
	ch chan bool
}

// openMsg indicates when in the sequence of log entries the Open operation
// occurred.  This is used to distinguish entries that arrived prior to the
// Open operation to those that arrived after.
type openMsg struct {
}

// --- Internal synchronization messages

// Exporter defines the base span exporter implementation
type Exporter struct {
	m sync.Mutex

	// closed is true if the path to the forwarder's trace sink is not
	// currently established.  While the Exporter is closed, any span export
	// operations will be buffered internally.
	closed bool

	// ch is used to send log entry events from the calling context to the
	// worker goroutine
	ch chan interface{}

	// proc contains the associated forwarder instance that processes the trace
	// events.
	proc forwarder
}

// NewExporter creates a new span exporter instance with the specified
// forwarded.
func NewExporter(proc forwarder) *Exporter {
	return &Exporter{
		m: sync.Mutex{},
		closed: true,
		ch:   make(chan interface{}, maxBuffer),
		proc: proc,
	}
}

// ExportSpan sends a completed span to be forwarded to the sink.
func (e *Exporter) ExportSpan(ctx context.Context, data *export.SpanData) {
	e.ch <- extractEntry(ctx, data)
}

// Open establishes the connection to the sink, and starts the worker flush
// process
func (e *Exporter) Open(attrs interface{}) error {
	e.m.Lock()
	defer e.m.Unlock()

	if !e.closed {
		return ErrAlreadyOpen
	}

	if err := e.proc.Open(attrs); err != nil {
		return err
	}

	// All checks passed, the exporter is now logically open.
	e.closed = false

	// Begin processing
	go e.processLoop(e.ch)
	e.ch <- &openMsg{}

	return nil
}

// Close gracefully terminates the connection to the sink, after allowing all
// previously buffered spans to be forwarded
func (e *Exporter) Close() {
	e.m.Lock()
	defer e.m.Unlock()

	if !e.closed {
		rsp := make(chan bool)
		e.ch <- &closeMsg{ ch: rsp }

		<-rsp
	}
}

// processLoop is the worker goroutine that is responsible for forwarding the
// buffered log entries to the final sink.
func (e *Exporter) processLoop(ch chan interface{}) {
	early := true

	for !e.closed {
		msg := <- ch

		switch pkt := msg.(type) {
		case *pbl.Entry:
			if err := e.proc.Forward(pkt, early); err != nil {
				log.Fatalf("Error forwarding log entry: %v returned %v", e.proc, err)
			}

		case *openMsg:
			early = false

		case *closeMsg:
			e.proc.Close()
			e.closed = true

			pkt.ch <- true
		}
	}
}
