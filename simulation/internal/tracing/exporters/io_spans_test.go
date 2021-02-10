package exporters

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Jim3Things/CloudChamber/simulation/pkg/protos/log"
)

func TestSingleRoot(t *testing.T) {
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
			{
				Tick:       0,
				Severity:   0,
				Name:       "test1",
				Text:       "text1",
				StackTrace: "stack1",
				Impacted:   nil,
				EventAction: log.Action_Trace,
				SpanId:     "",
			},
			{
				Tick:       1,
				Severity:   0,
				Name:       "test2",
				Text:       "text2",
				StackTrace: "stack2",
				Impacted:   nil,
				EventAction: log.Action_Trace,
				SpanId:     "",
			},
		},
		Infrastructure: false,
	}

	s.add(entry, &io)

	assert.Equal(t,
		"\n[0102030405060708:0000000000000000] ok root ():\n"+
			"    stacks\n\n"+
			"      @   0: [D] (test1) text1\n"+
			"        stack1\n"+
			"      @   1: [D] (test2) text2\n"+
			"        stack2\n", io.String())

	assert.Equal(t, 0, len(s.known))
	assert.Equal(t, 0, len(s.active))
}

func TestDoubleRoot(t *testing.T) {
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
			{
				Tick:       0,
				Severity:   0,
				Name:       "test1",
				Text:       "text1",
				StackTrace: "stack1",
				Impacted:   nil,
				EventAction: log.Action_Trace,
				SpanId:     "",
			},
			{
				Tick:       1,
				Severity:   0,
				Name:       "test2",
				Text:       "text2",
				StackTrace: "stack2",
				Impacted:   nil,
				EventAction: log.Action_Trace,
				SpanId:     "",
			},
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
			{
				Tick:       2,
				Severity:   0,
				Name:       "test3",
				Text:       "text3",
				StackTrace: "stack3",
				Impacted:   nil,
				EventAction: log.Action_Trace,
				SpanId:     "",
			},
			{
				Tick:       3,
				Severity:   0,
				Name:       "test4",
				Text:       "text4",
				StackTrace: "stack4",
				Impacted:   nil,
				EventAction: log.Action_Trace,
				SpanId:     "",
			},
		},
		Infrastructure: false,
	}

	s.add(entry, &io)

	assert.Equal(t,
		"\n[0102030405060708:0000000000000000] ok root ():\n"+
			"    stacks\n\n"+
			"      @   0: [D] (test1) text1\n"+
			"        stack1\n"+
			"      @   1: [D] (test2) text2\n"+
			"        stack2\n", io.String())

	assert.Equal(t, 0, len(s.known))
	assert.Equal(t, 0, len(s.active))

	io = bytes.Buffer{}
	s.add(entry2, &io)

	assert.Equal(t,
		"\n[1102030405060708:0000000000000000] ok root ():\n"+
			"    stacks\n\n"+
			"      @   2: [D] (test3) text3\n"+
			"        stack3\n"+
			"      @   3: [D] (test4) text4\n"+
			"        stack4\n",
		io.String())

	assert.Equal(t, 0, len(s.known))
	assert.Equal(t, 0, len(s.active))
}

func TestSimpleChildFirst(t *testing.T) {
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
			{
				Tick:       0,
				Severity:   0,
				Name:       "test1",
				Text:       "text1",
				StackTrace: "stack1",
				Impacted:   nil,
				EventAction: log.Action_Trace,
				SpanId:     "",
			},
			{
				Tick:       1,
				Severity:   0,
				Name:       "test2",
				Text:       "text2",
				StackTrace: "stack2",
				Impacted:   nil,
				EventAction: log.Action_SpanStart,
				SpanId:     "1102030405060708",
			},
			{
				Tick:       1,
				Severity:   0,
				Name:       "test2",
				Text:       "text2",
				StackTrace: "stack2",
				Impacted:   nil,
				EventAction: log.Action_Trace,
				SpanId:     "",
			},
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
			{
				Tick:       2,
				Severity:   0,
				Name:       "test3",
				Text:       "text3",
				StackTrace: "stack3",
				Impacted:   nil,
				EventAction: log.Action_Trace,
				SpanId:     "",
			},
			{
				Tick:       3,
				Severity:   0,
				Name:       "test4",
				Text:       "text4",
				StackTrace: "stack4",
				Impacted:   nil,
				EventAction: log.Action_Trace,
				SpanId:     "",
			},
		},
		Infrastructure: false,
	}

	s.add(entry2, &io)

	assert.Equal(t, 0, len(io.String()))
	assert.Equal(t, 1, len(s.known))
	assert.Equal(t, 1, len(s.active))

	a := s.active[traceID]
	assert.Equal(t, 1, len(a.open))
	assert.Equal(t, 1, len(a.closed))

	io = bytes.Buffer{}
	s.add(entry, &io)

	assert.Equal(t,
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

	assert.Equal(t, 0, len(s.known))
	assert.Equal(t, 0, len(s.active))

}

func TestSimpleChildLast(t *testing.T) {
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
			{
				Tick:       0,
				Severity:   0,
				Name:       "test1",
				Text:       "text1",
				StackTrace: "stack1",
				Impacted:   nil,
				EventAction: log.Action_Trace,
				SpanId:     "",
			},
			{
				Tick:       1,
				Severity:   0,
				Name:       "test2",
				Text:       "text2",
				StackTrace: "stack2",
				Impacted:   nil,
				EventAction: log.Action_SpanStart,
				SpanId:     "1102030405060708",
			},
			{
				Tick:       1,
				Severity:   0,
				Name:       "test2",
				Text:       "text2",
				StackTrace: "stack2",
				Impacted:   nil,
				EventAction: log.Action_Trace,
				SpanId:     "",
			},
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
			{
				Tick:       2,
				Severity:   0,
				Name:       "test3",
				Text:       "text3",
				StackTrace: "stack3",
				Impacted:   nil,
				EventAction: log.Action_Trace,
				SpanId:     "",
			},
			{
				Tick:       3,
				Severity:   0,
				Name:       "test4",
				Text:       "text4",
				StackTrace: "stack4",
				Impacted:   nil,
				EventAction: log.Action_Trace,
				SpanId:     "",
			},
		},
		Infrastructure: false,
	}

	s.add(entry, &io)

	assert.Equal(t, 0, len(io.String()))
	assert.Equal(t, 1, len(s.known))
	assert.Equal(t, 1, len(s.active))

	a := s.active[traceID]
	assert.Equal(t, 1, len(a.open))
	assert.Equal(t, 1, len(a.closed))

	io = bytes.Buffer{}
	s.add(entry2, &io)

	assert.Equal(t,
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

	assert.Equal(t, 0, len(s.known))
	assert.Equal(t, 0, len(s.active))
}
