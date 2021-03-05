package exporters

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Jim3Things/CloudChamber/simulation/pkg/protos/log"
)

type ioSpanTestSuite struct {
	suite.Suite
}

func (ts *ioSpanTestSuite) newTraceEvent(i int, tick int64, s log.Severity) *log.Event {
	return &log.Event{
		Tick:        tick,
		Severity:    s,
		Name:        fmt.Sprintf("test%d", i),
		Text:        fmt.Sprintf("text%d", i),
		StackTrace:  fmt.Sprintf("stack%d", i),
		Impacted:    nil,
		EventAction: log.Action_Trace,
		SpanId:      "",
		LinkId:      "",
	}
}

func (ts *ioSpanTestSuite) newSpanStartEvent(i int, tick int64, spanId string) *log.Event {
	return &log.Event{
		Tick:        tick,
		Severity:    0,
		Name:        fmt.Sprintf("test%d", i),
		Text:        fmt.Sprintf("text%d", i),
		StackTrace:  fmt.Sprintf("stack%d", i),
		Impacted:    nil,
		EventAction: log.Action_SpanStart,
		SpanId:      spanId,
		LinkId:      "",
	}
}

func (ts *ioSpanTestSuite) TestSingleRoot() {
	assert := ts.Assert()

	s := newSpans()

	io := bytes.Buffer{}

	entry := &log.Entry{
		Name:       "root",
		SpanID:     "0102030405060708",
		ParentID:   "0000000000000000",
		TraceID:    "11020304050607080102030405060708",
		Status:     "ok",
		StackTrace: "stacks",
		Event: []*log.Event{
			ts.newTraceEvent(1, 0, 0),
			ts.newTraceEvent(2, 1, 0),
		},
		Infrastructure: false,
	}

	s.add(entry, &io)

	assert.Equal(
		"\n[0102030405060708:0000000000000000] ok root ():\n"+
			"    stacks\n\n"+
			"      @   0: [D] (test1) text1\n"+
			"        stack1\n"+
			"      @   1: [D] (test2) text2\n"+
			"        stack2\n", io.String())

	assert.Equal(0, len(s.known))
	assert.Equal(0, len(s.active))
}

func (ts *ioSpanTestSuite) TestDoubleRoot() {
	assert := ts.Assert()

	s := newSpans()

	io := bytes.Buffer{}

	entry := &log.Entry{
		Name:       "root",
		SpanID:     "0102030405060708",
		ParentID:   "0000000000000000",
		TraceID:    "11020304050607080102030405060708",
		Status:     "ok",
		StackTrace: "stacks",
		Event: []*log.Event{
			ts.newTraceEvent(1, 0, 0),
			ts.newTraceEvent(2, 1, 0),
		},
		Infrastructure: false,
	}

	entry2 := &log.Entry{
		Name:       "root",
		SpanID:     "1102030405060708",
		ParentID:   "0000000000000000",
		TraceID:    "11020304050607080102030405060708",
		Status:     "ok",
		StackTrace: "stacks",
		Event: []*log.Event{
			ts.newTraceEvent(3, 2, 0),
			ts.newTraceEvent(4, 3, 0),
		},
		Infrastructure: false,
	}

	s.add(entry, &io)

	assert.Equal(
		"\n[0102030405060708:0000000000000000] ok root ():\n"+
			"    stacks\n\n"+
			"      @   0: [D] (test1) text1\n"+
			"        stack1\n"+
			"      @   1: [D] (test2) text2\n"+
			"        stack2\n", io.String())

	assert.Equal(0, len(s.known))
	assert.Equal(0, len(s.active))

	io = bytes.Buffer{}
	s.add(entry2, &io)

	assert.Equal(
		"\n[1102030405060708:0000000000000000] ok root ():\n"+
			"    stacks\n\n"+
			"      @   2: [D] (test3) text3\n"+
			"        stack3\n"+
			"      @   3: [D] (test4) text4\n"+
			"        stack4\n",
		io.String())

	assert.Equal(0, len(s.known))
	assert.Equal(0, len(s.active))
}

func (ts *ioSpanTestSuite) TestSimpleChildFirst() {
	assert := ts.Assert()

	traceID := "11020304050607080102030405060708"
	s := newSpans()

	io := bytes.Buffer{}

	entry := &log.Entry{
		Name:       "root",
		SpanID:     "0102030405060708",
		ParentID:   "0000000000000000",
		TraceID:    traceID,
		Status:     "ok",
		StackTrace: "stacks",
		Event: []*log.Event{
			ts.newTraceEvent(1, 0, 0),
			ts.newSpanStartEvent(2, 1, "1102030405060708"),
			ts.newTraceEvent(2, 1, 0),
		},
		Infrastructure: false,
	}

	entry2 := &log.Entry{
		Name:       "root",
		SpanID:     "1102030405060708",
		ParentID:   "0102030405060708",
		TraceID:    traceID,
		Status:     "ok",
		StackTrace: "stacks",
		Event: []*log.Event{
			ts.newTraceEvent(3, 2, 0),
			ts.newTraceEvent(4, 3, 0),
		},
		Infrastructure: false,
	}

	s.add(entry2, &io)

	assert.Equal(0, len(io.String()))
	assert.Equal(1, len(s.known))
	assert.Equal(1, len(s.active))

	a := s.active[traceID]
	assert.Equal(1, len(a.open))
	assert.Equal(1, len(a.closed))

	io = bytes.Buffer{}
	s.add(entry, &io)

	assert.Equal(
		"\n[0102030405060708:0000000000000000] ok root ():\n"+
			"    stacks\n\n"+
			"      @   0: [D] (test1) text1\n"+
			"        stack1\n"+
			"\n    [1102030405060708:0102030405060708] ok root ():\n"+
			"        stacks\n\n"+
			"          @   2: [D] (test3) text3\n"+
			"            stack3\n"+
			"          @   3: [D] (test4) text4\n"+
			"            stack4\n"+
			"      @   1: [D] (test2) text2\n"+
			"        stack2\n",
		io.String())

	assert.Equal(0, len(s.known))
	assert.Equal(0, len(s.active))

}

func (ts *ioSpanTestSuite) TestSimpleChildLast() {
	assert := ts.Assert()

	traceID := "11020304050607080102030405060708"
	s := newSpans()

	io := bytes.Buffer{}

	entry := &log.Entry{
		Name:       "root",
		SpanID:     "0102030405060708",
		ParentID:   "0000000000000000",
		TraceID:    traceID,
		Status:     "ok",
		StackTrace: "stacks",
		Event: []*log.Event{
			ts.newTraceEvent(1, 0, 0),
			ts.newSpanStartEvent(2, 1, "1102030405060708"),
			ts.newTraceEvent(2, 1, 0),
		},
		Infrastructure: false,
	}

	entry2 := &log.Entry{
		Name:       "root",
		SpanID:     "1102030405060708",
		ParentID:   "0102030405060708",
		TraceID:    traceID,
		Status:     "ok",
		StackTrace: "stacks",
		Event: []*log.Event{
			ts.newTraceEvent(3, 2, 0),
			ts.newTraceEvent(4, 3, 0),
		},
		Infrastructure: false,
	}

	s.add(entry, &io)

	assert.Equal(0, len(io.String()))
	assert.Equal(1, len(s.known))
	assert.Equal(1, len(s.active))

	a := s.active[traceID]
	assert.Equal(1, len(a.open))
	assert.Equal(1, len(a.closed))

	io = bytes.Buffer{}
	s.add(entry2, &io)

	assert.Equal(
		"\n[0102030405060708:0000000000000000] ok root ():\n"+
			"    stacks\n\n"+
			"      @   0: [D] (test1) text1\n"+
			"        stack1\n"+
			"\n    [1102030405060708:0102030405060708] ok root ():\n"+
			"        stacks\n\n"+
			"          @   2: [D] (test3) text3\n"+
			"            stack3\n"+
			"          @   3: [D] (test4) text4\n"+
			"            stack4\n"+
			"      @   1: [D] (test2) text2\n"+
			"        stack2\n",
		io.String())

	assert.Equal(0, len(s.known))
	assert.Equal(0, len(s.active))
}

func TestIoSpans(t *testing.T) {
	suite.Run(t, new(ioSpanTestSuite))
}
