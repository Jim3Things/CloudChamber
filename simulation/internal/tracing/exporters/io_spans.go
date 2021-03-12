package exporters

import (
	"fmt"
	"io"
	stdLog "log"
	"sync"

	"go.opentelemetry.io/otel/api/trace"

	"github.com/Jim3Things/CloudChamber/simulation/pkg/protos/log"
)

type ioSpans struct {
	m sync.Mutex

	// active is keyed by traceID, with values of the currently incomplete
	// spanIDs for that trace.
	active map[string]*activeEntry

	// known is keyed by spanID, and contains the associated log entry, or nil
	// if the entry is derived from the parent span ID for a span that has not
	// yet emitted its state.
	known map[string]*log.Entry
}

// activeEntry defines the set of spans associated with a given TraceID.  It
// manages enough state to determine when the information set is complete and
// the entries for that TraceID can be emitted.
//
// A SpanIDs is introduced as 'open' as soon as it is seen.  It is then moved
// to 'closed' as it completes its export operation.  Since all span IDs must
// be seen prior to its export completing, once the 'open' set is empty, all
// spans for the given TraceID must have been processed, and we can emit the
// functionally ordered traces.
type activeEntry struct {
	// root is the span ID that started the sequence for the traceID
	root string

	// open are the SpanIDs that have been found for this traceID, but have not
	// yet completed an export operation.
	open map[string]bool

	// closed are the SpanIDs that have been found an have completed their
	// export operation
	closed map[string]bool
}

// newSpan creates a new, empty, spans instance
func newSpans() *ioSpans {
	return &ioSpans{
		m:      sync.Mutex{},
		active: make(map[string]*activeEntry),
		known:  make(map[string]*log.Entry),
	}
}

// getOrAddActive either retrieves an existing activeEntry for the given
// TraceID, or a new one is created and returned.
func (s *ioSpans) getOrAddActive(traceID string) *activeEntry {
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

// emit processes the indicated span, sending the formatted output to the
// supplied writer.  It will recursively process child spans, and manages the
// line indent amount to indicate descent level
func (s *ioSpans) emit(a *activeEntry, spanID string, io io.Writer, indent string) {
	entry, ok := s.known[spanID]
	if !ok {
		stdLog.Fatalf("Missing span: %q", spanID)
	}

	// Spans are set off by surrounding blank lines
	_, _ = io.Write([]byte(
		fmt.Sprintf(
			"\n%s\n",
			formatEntry(entry, false, indent))))

	for _, e := range entry.Event {
		if e.EventAction == log.Action_SpanStart {
			// This entry is for a child span creation event.
			// Recursively process it.
			s.emit(a, e.SpanId, io, indent+tab)
			delete(s.known, e.SpanId)
			delete(a.closed, e.SpanId)
		} else {
			_, _ = io.Write([]byte(formatEvent(e, indent+tab)))
		}
	}
}

// add is the point where a log entry is added to the set of active spans.  If
// it results in full closure of the parent span then that subtree is emitted
// to the IO writer and discarded.
func (s *ioSpans) add(entry *log.Entry, io io.Writer) {
	s.m.Lock()
	defer s.m.Unlock()

	traceID := entry.TraceID
	spanID := entry.SpanID
	parentID := entry.ParentID

	_, err := trace.SpanIDFromHex(parentID)
	hasParent := err == nil

	// First, let's record this span as a known span
	s.known[spanID] = entry

	// Then, get the active entries for its traceID
	a := s.getOrAddActive(traceID)

	// add this entry, if not already present to the active entry's open list
	a.open[spanID] = true

	if !hasParent {
		if a.root != "" && a.root != spanID {
			s.recordFatal(
				a,
				entry,
				fmt.Sprintf(
					"expected only one root, tried to replace %q with %q",
					a.root,
					spanID))
		}

		a.root = spanID
	} else if _, ok := s.known[parentID]; !ok {
		// add this entry's parent to active, if not in the known list
		a.open[parentID] = true
	}

	// go through the full set of entries.  For each span start, add that
	// child ID to the trace ID set, if not already known
	for _, e := range entry.Event {
		if e.EventAction == log.Action_SpanStart {
			if _, ok := s.known[e.SpanId]; !ok {
				a.open[e.SpanId] = true
			}
		}
	}

	// move this span ID from the active entry's open list to the closed list.
	delete(a.open, spanID)
	a.closed[spanID] = true

	if len(a.open) == 0 {
		// if active entry's open list is now empty, we're done.  So format
		// the tree of traces.
		s.emit(a, a.root, io, "")
		delete(s.known, a.root)
		delete(a.closed, a.root)

		// Ensure that the closed list is empty
		if len(a.closed) != 0 {
			s.recordFatal(a, entry, "Expected all closed")
		}

		delete(s.active, traceID)
	}
}

func (s *ioSpans) recordFatal(a *activeEntry, entry *log.Entry, cause string) {
	msg := fmt.Sprintf("%s, %v: \n", cause, a)
	msg = fmt.Sprintf("%sCurrent entry:\n%s\n", msg, formatEntry(entry, false, tab))
	for _, event := range entry.Event {
		msg = fmt.Sprintf("%s%s\n", msg, formatEvent(event, tab + tab))
	}
	msg = fmt.Sprintf("%sClosed (%d):\n", msg, len(a.closed))
	for id := range a.closed {
		sp, ok := s.known[id]
		if !ok {
			msg = fmt.Sprintf("%s id: %s not found\n", msg, id)
		} else {
			msg = fmt.Sprintf("%s%s\n", msg, formatEntry(sp, false, tab))

			for _, event := range sp.Event {
				msg = fmt.Sprintf("%s%s\n", msg, formatEvent(event, tab + tab))
			}
		}
	}

	msg = fmt.Sprintf("%sOpen (%d):\n", msg, len(a.open))
	for id := range a.open {
		sp, ok := s.known[id]
		if !ok {
			msg = fmt.Sprintf("%s id: %s not found\n", msg, id)
		} else {
			msg = fmt.Sprintf("%s%s\n", msg, formatEntry(sp, false, tab))

			for _, event := range sp.Event {
				msg = fmt.Sprintf("%s%s\n", msg, formatEvent(event, tab+tab))
			}
		}
	}

	stdLog.Fatal(msg)

}
