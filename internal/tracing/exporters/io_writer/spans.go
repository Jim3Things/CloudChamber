package io_writer

import (
	"fmt"
	"io"
	"sync"

	"go.opentelemetry.io/otel/api/trace"

	"github.com/Jim3Things/CloudChamber/internal/tracing/exporters/common"
	"github.com/Jim3Things/CloudChamber/pkg/protos/log"
)

const (
	tab = "    "
)

type activeEntry struct {
	root string
	open map[string]bool
	closed map[string]bool
}

type spans struct {
	m sync.Mutex

	// active is keyed by traceID, with values of the currently incomplete
	// spanIDs for that trace.
	active map[string]*activeEntry

	// known is keyed by spanID, and contains the associated log entry, or nil
	// if the entry is derived from the parent span ID for a span that has not
	// yet emitted its state.
	known map[string]*log.Entry
}

func newSpans() *spans {
	return &spans{
		m:      sync.Mutex{},
		active: make(map[string]*activeEntry),
		known:  make(map[string]*log.Entry),
	}
}

func (s *spans) getOrAddActive(traceID string) *activeEntry {
	entry, ok := s.active[traceID]
	if ok {
		return entry
	}

	// new entry
	entry = &activeEntry{
		root:   "",
		open:   make(map[string]bool),
		closed: make(map[string]bool),
	}

	s.active[traceID] = entry
	return entry
}

func (s *spans) emit(a *activeEntry, spanID string, io io.Writer, indent string) {
	entry, ok := s.known[spanID]
	if !ok { panic(fmt.Sprintf("Missing span: %q", spanID))}

	spanHeader := fmt.Sprintf("\n%s\n", common.FormatEntry(entry, false, indent))
	_, _ = io.Write([]byte(spanHeader))
	for _, e := range entry.Event {
		if e.SpanStart {
			s.emit(a, e.SpanId, io, indent + tab)
			delete(s.known, e.SpanId)
			delete(a.closed, e.SpanId)
		} else {
			_, _ = io.Write([]byte(common.FormatEvent(e, indent + tab)))
		}
	}
}

// add is the point where a log entry is added to the set of active spans.  If
// it results in full closure of the parent span then that subtree is emitted
// to the IO writer and discarded.
func (s *spans) add(entry *log.Entry, io io.Writer) {
	s.m.Lock()
	defer s.m.Unlock()

	traceID := entry.TraceID
	spanID := entry.SpanID
	parentID := entry.ParentID

	_, err := trace.SpanIDFromHex(parentID)
	hasParent := err == nil

	// First, let's record this span as a known span
	s.known[spanID] = entry

	// Then, let's see if we have an active trace.  Create one if not.
	a := s.getOrAddActive(traceID)

	// add this entry, if not already present to the active entry's open list
	a.open[spanID] = true

	if !hasParent {
		a.root = spanID
	} else {
		// add this entry's parent to active, if not in the known list
		if _, ok := s.known[parentID]; !ok {
			a.open[parentID] = true
		}
	}

	// go through the full set of entries.  For each span start:
	// - add that child ID to the trace ID set, if not already known
	for _, e := range entry.Event {
		if e.SpanStart {
			if _, ok := a.closed[e.SpanId]; !ok {
				a.open[e.SpanId] = true
			}
		}
	}

	// move this span ID from the active entry's open list to the closed list.
	delete(a.open, spanID)
	a.closed[spanID] = true

	if len(a.open) == 0 {
		// if active entry's open list is now empty:
		// - issue emit on it
		// - remove the trace ID when emit returns
		s.emit(a, a.root, io, "")
		delete(s.known, a.root)
		delete(a.closed, a.root)

		// Ensure that the closed list is empty
		if len(a.closed) != 0 { panic(fmt.Sprintf("Expected all closed, %v", a))}
		delete(s.active, traceID)
	}
}
