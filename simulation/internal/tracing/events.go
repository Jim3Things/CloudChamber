package tracing

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"time"

	"go.opentelemetry.io/otel/api/kv"
	"go.opentelemetry.io/otel/api/trace"

	"github.com/Jim3Things/CloudChamber/simulation/internal/common"
	pbl "github.com/Jim3Things/CloudChamber/simulation/pkg/protos/log"
)

// TraceDetail holds attributes and processing data that is used as part of the
// underlying AddTrace call.
type TraceDetail struct {
	// replacements is the list of replacement processing commands used to
	// filter the log string.
	replacements []formatEntry
}

// TraceAnnotation defines the type signature for an annotation function that
// decorates a log trace event.
type TraceAnnotation func(cfg *TraceDetail)

// formatEntry defines a single replacement processing command.  This consists
// of a compiled regexp instance and the replacement string to use in a
// ReplaceAllString call.
type formatEntry struct {
	re   *regexp.Regexp
	repl string
}

// filteredFormat returns the formatted string with all filter rules applied.
func (td *TraceDetail) filteredFormat(a ...interface{}) string {
	res := formatIf(a...)

	for _, filter := range td.replacements {
		res = filter.re.ReplaceAllString(res, filter.repl)
	}

	return res
}

// newTraceDetail creates a new empty TraceDetail instance.
func newTraceDetail() *TraceDetail {
	return &TraceDetail{
		replacements: []formatEntry{},
	}
}

// addFilter adds a filter replacement instance to the trace.  Each instance
// contains a compiled regex used to match, and the replacement string to use
// for that match.
func (td *TraceDetail) addFilter(re *regexp.Regexp, repl string) {
	td.replacements = append(td.replacements, formatEntry{
		re:   re,
		repl: repl,
	})
}

// WithReplacement states that the event text is to be searched using the match
// regex, and for every occurrence that is found, it is to be replaced by the
// value of the repl parameter.
func WithReplacement(match *regexp.Regexp, repl string) TraceAnnotation {
	return func(cfg *TraceDetail) {
		cfg.addFilter(match, repl)
	}
}

// addAnnotations processes all recognized TraceAnnotation functions in the
// supplied slice, stopping at the first non-TraceAnnotation entry.  The index
// to that first non-TraceAnnotation entry is then returned.
func addAnnotations(cfg *TraceDetail, a []interface{}) int {
	i := 0
	for _, item := range a {
		if annotation, ok := item.(TraceAnnotation); ok {
			annotation(cfg)
			i++
		}
	}

	return i
}

// There should be an Xxx method for every severity level, plus some specific
// scenario functions (such as OnEnter to log an information entry about
// arrival at a specific method).
//
// Note: The set of methods that are implemented below are based on what is
// currently needed.  Others will be added as required.

// Debug posts a debug-level trace event
func Debug(ctx context.Context, a ...interface{}) {
	cfg := newTraceDetail()
	start := addAnnotations(cfg, a)

	trace.SpanFromContext(ctx).AddEventWithTimestamp(
		ctx,
		time.Now(),
		MethodName(2),
		kv.Int64(StepperTicksKey, common.TickFromContext(ctx)),
		kv.Int64(SeverityKey, int64(pbl.Severity_Debug)),
		kv.String(StackTraceKey, StackTrace()),
		kv.String(MessageTextKey, cfg.filteredFormat(a[start:]...)))
}

// Info posts an informational trace event
func Info(ctx context.Context, a ...interface{}) {
	cfg := newTraceDetail()
	start := addAnnotations(cfg, a)

	trace.SpanFromContext(ctx).AddEventWithTimestamp(
		ctx,
		time.Now(),
		MethodName(2),
		kv.Int64(StepperTicksKey, common.TickFromContext(ctx)),
		kv.Int64(SeverityKey, int64(pbl.Severity_Info)),
		kv.String(StackTraceKey, StackTrace()),
		kv.String(MessageTextKey, cfg.filteredFormat(a[start:]...)))
}

// Warn posts a warning trace event
func Warn(ctx context.Context, a ...interface{}) {
	cfg := newTraceDetail()
	start := addAnnotations(cfg, a)

	trace.SpanFromContext(ctx).AddEventWithTimestamp(
		ctx,
		time.Now(),
		MethodName(2),
		kv.Int64(StepperTicksKey, common.TickFromContext(ctx)),
		kv.Int64(SeverityKey, int64(pbl.Severity_Warning)),
		kv.String(StackTraceKey, StackTrace()),
		kv.String(MessageTextKey, cfg.filteredFormat(a[start:]...)))
}

// Error posts an error trace event
func Error(ctx context.Context, a ...interface{}) error {
	cfg := newTraceDetail()
	start := addAnnotations(cfg, a)

	var err error

	if a == nil {
		err = errors.New("missing error details")
	} else {
		switch len(a) {
		case start:
			err = errors.New("missing error details")

		case start + 1:
			switch e := a[start].(type) {
			case error:
				err = e

			case string:
				err = errors.New(cfg.filteredFormat(e))

			default:
				err = fmt.Errorf("unexpected values: %v", e)
			}

		default:
			err = errors.New(cfg.filteredFormat(a...))
		}
	}

	trace.SpanFromContext(ctx).AddEventWithTimestamp(
		ctx,
		time.Now(),
		fmt.Sprintf("Error from %q", MethodName(3)),
		kv.Int64(StepperTicksKey, common.TickFromContext(ctx)),
		kv.Int64(SeverityKey, int64(pbl.Severity_Warning)),
		kv.String(StackTraceKey, StackTrace()),
		kv.String(MessageTextKey, err.Error()))

	return err
}

// Fatal traces the error, and then terminates the process.
func Fatal(ctx context.Context, a ...interface{}) {
	log.Fatal(Error(ctx, a))
}
