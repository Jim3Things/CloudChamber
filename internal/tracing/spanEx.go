package tracing

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync/atomic"

	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/api/kv"
	"go.opentelemetry.io/otel/api/trace"

	"github.com/Jim3Things/CloudChamber/internal/common"
	pbl "github.com/Jim3Things/CloudChamber/pkg/protos/log"
)

// linkID is a global number that is used to ensure that add-link traces and
// associated link-tag values are unique, at least within a span.  Using a
// global like this, with an atomic update, avoids modifying the SpanEx state,
// and the resulting need to create a new updated context.
var linkID int64 = 0

// SpanEx extends the OpenTelemetry Span structure with additional features
// needed by CloudChamber
type SpanEx struct {
	// Original OpenTelemetry span
	trace.Span

	// isInternal is true if this span started with SpanKindInternal.
	isInternal bool
}

type decorator func(ctx context.Context) context.Context

// Mask applies the policy that a child span of a span that is marked as
// infrastructure (span kind is internal) should also be market as internal.
func (s *SpanEx) Mask(kind trace.SpanKind) trace.SpanKind {
	if s.isInternal {
		return trace.SpanKindInternal
	}

	return kind
}

// startSpanConfig defines the attributes used when starting a new span, and
// that can be overridden by the caller.
type startSpanConfig struct {
	name        string
	kind        trace.SpanKind
	stackTrace  string
	tick        int64
	decorations []decorator
	reason      string
	link        trace.SpanContext
	linkTag     string
	newRoot     bool
}

// StartSpanOption denotes optional decoration methods used on StartSpan
type StartSpanOption func(*startSpanConfig)

// WithName adds the supplied value as the name of the span under creation
func WithName(name string) StartSpanOption {
	return func(cfg *startSpanConfig) {
		cfg.name = name
	}
}

// AsInternal sets the span as an infrastructure (or internal) span kind.
func AsInternal() StartSpanOption {
	return func(cfg *startSpanConfig) {
		cfg.kind = trace.SpanKindInternal
	}
}

// WithKind sets the span kind to that specified in the argument.
func WithKind(kind trace.SpanKind) StartSpanOption {
	return func(cfg *startSpanConfig) {
		cfg.kind = kind
	}
}

// WithContextValue decorates the resulting context using the supplied function
func WithContextValue(action decorator) StartSpanOption {
	return func(cfg *startSpanConfig) {
		cfg.decorations = append(cfg.decorations, action)
	}
}

// WithReason adds a friendly description for the reason behind the logic in
// the span.
func WithReason(reason string) StartSpanOption {
	return func(cfg *startSpanConfig) {
		cfg.reason = reason
	}
}

// WithLink adds a link-to target, if there is one, to the span.
func WithLink(sc trace.SpanContext, tag string) StartSpanOption {
	return func(cfg *startSpanConfig) {
		cfg.link = sc
		cfg.linkTag = tag
	}
}

// mayLinkTo is used in the underlying trace span call.  It either returns
// via normal LinkedTo, if there is a link-to span context, or a null
// operation that does not decorate the trace span, if it does not.
func mayLinkTo(sc trace.SpanContext) trace.StartOption {
	if sc.HasSpanID() && sc.HasTraceID() {
		return trace.LinkedTo(sc)
	}

	return nullOption()
}

// mayLinkTag is a parallel helper function that supplied the unique add-link
// value to allow for this span to be correctly placed relative to the caller's
// sequence of actions.  If no such tag is present, this adds nothing to the
// start span operation.
func mayLinkTag(tag string) trace.StartOption {
	if len(tag) > 0 {
		return trace.WithAttributes(kv.String(LinkTagKey, tag))
	}

	return nullOption()
}

// WithNewRoot specifies that the span is to start a new top level span, even
// if there is a potential parent span available in the context.
func WithNewRoot() StartSpanOption {
	return func(cfg *startSpanConfig) {
		cfg.newRoot = true
	}
}

// mayNewRoot is used in the underlying trace span call.  It either signals
// that this is a new root, if that option had been selected, or a null
// operation that does not decorate the trace span call, if it does not.
func mayNewRoot(newRoot bool) trace.StartOption {
	if newRoot {
		return trace.WithNewRoot()
	}

	return nullOption()
}

// nullOption is a helper function that can be added to a trace span call, but
// performs no decoration.
func nullOption() trace.StartOption {
	return func(s *trace.StartConfig) {}
}

// StartSpan creates a new tracing span, with the attributes and linkages
// that the Cloud Chamber logging system expects
func StartSpan(
	ctx context.Context,
	options ...StartSpanOption) (context.Context, trace.Span) {
	cfg := startSpanConfig{
		name:        MethodName(2),
		kind:        trace.SpanKindServer,
		stackTrace:  StackTrace(),
		tick:        -1,
		decorations: []decorator{},
		link:        trace.SpanContext{},
		linkTag:     "",
		newRoot:     false,
	}

	for _, opt := range options {
		opt(&cfg)
	}

	parent := trace.SpanFromContext(ctx)

	if s, ok := parent.(SpanEx); ok {
		cfg.kind = s.Mask(cfg.kind)
	}

	tr := global.TraceProvider().Tracer("")

	ctxChild, span := tr.Start(ctx, cfg.name,
		trace.WithSpanKind(cfg.kind),
		mayLinkTo(cfg.link),
		mayLinkTag(cfg.linkTag),
		mayNewRoot(cfg.newRoot),
		trace.WithAttributes(kv.String(ReasonKey, cfg.reason)),
		trace.WithAttributes(kv.String(StackTraceKey, cfg.stackTrace)))

	if !cfg.newRoot && parent.SpanContext().HasSpanID() {
		parent.AddEvent(
			ctxChild,
			cfg.name,
			kv.Int64(ActionKey, int64(pbl.Action_SpanStart)),
			kv.Int64(StepperTicksKey, cfg.tick),
			kv.Int64(SeverityKey, int64(pbl.Severity_Info)),
			kv.String(StackTraceKey, StackTrace()),
			kv.String(ChildSpanKey, span.SpanContext().SpanID.String()))
	}

	ccSpan := SpanEx{
		Span:       span,
		isInternal: cfg.kind == trace.SpanKindInternal,
	}

	for _, action := range cfg.decorations {
		ctxChild = action(ctxChild)
	}

	return trace.ContextWithSpan(ctxChild, ccSpan), ccSpan
}

// UpdateSpanName replaces the current span name string with the
// formatted string provided.  The span will end up with the last
// name provided.
func UpdateSpanName(ctx context.Context, a ...interface{}) {
	trace.SpanFromContext(ctx).AddEvent(
		ctx,
		MethodName(2),
		kv.Int64(ActionKey, int64(pbl.Action_UpdateSpanName)),
		kv.Int64(StepperTicksKey, common.TickFromContext(ctx)),
		kv.Int64(SeverityKey, int64(pbl.Severity_Info)),
		kv.String(StackTraceKey, StackTrace()),
		kv.String(MessageTextKey, formatIf(a...)))
}

// UpdateSpanReason replaces the current span reason with the formatted
// string provided.  The span will end up with the last reason provided.
func UpdateSpanReason(ctx context.Context, a ...interface{}) {
	trace.SpanFromContext(ctx).AddEvent(
		ctx,
		MethodName(2),
		kv.Int64(ActionKey, int64(pbl.Action_UpdateReason)),
		kv.Int64(StepperTicksKey, common.TickFromContext(ctx)),
		kv.Int64(SeverityKey, int64(pbl.Severity_Info)),
		kv.String(StackTraceKey, StackTrace()),
		kv.String(MessageTextKey, formatIf(a...)))
}

// AddLink adds an event that marks the point where a request was made that may
// result in a related span.  This contains the unique link tag that the target
// span should also provide, with the intention that a structured formatter can
// place that target span in the correct place in the execution sequence.
func AddLink(ctx context.Context, tag string) {
	trace.SpanFromContext(ctx).AddEvent(
		ctx,
		MethodName(2),
		kv.Int64(ActionKey, int64(pbl.Action_AddLink)),
		kv.Int64(StepperTicksKey, common.TickFromContext(ctx)),
		kv.Int64(SeverityKey, int64(pbl.Severity_Info)),
		kv.String(StackTraceKey, StackTrace()),
		kv.String(LinkTagKey, tag))
}

// GetAndMarkLink returns a string link tag, assuming that this is an extended
// span.
func GetAndMarkLink(parent trace.Span) (string, bool) {
	if _, ok := parent.(SpanEx); ok {
		val := atomic.AddInt64(&linkID, 1)

		return fmt.Sprintf("link-%d", val), true
	}

	return "", false
}

// TraceDetail holds the attribute data that is passed to the underlying
// AddTrace call.  These attributes are assumed to be key value pairs that have
// arrays of strings for their values.
type TraceDetail struct {
	details map[string][]string
}

// newTraceDetail creates a new empty TraceDetail instance.
func newTraceDetail() *TraceDetail {
	return &TraceDetail{
		details: make(map[string][]string),
	}
}

// addEntry adds a key-value pair.  If the key already exists, the new value
// is appended into the value array.  This is in support of assembling the
// results from multiple TraceAnnotation calls, such as multiple WithImpactXxx
// calls.
func (td *TraceDetail) addEntry(key string, value string) {
	item, ok := td.details[key]
	if !ok {
		item = []string{}
	}

	item = append(item, value)
	td.details[key] = item
}

// toKvPairs converts the returns of the TraceDetail instance as one or more
// KeyValue instances.
func (td *TraceDetail) toKvPairs() []kv.KeyValue {
	var res []kv.KeyValue

	for key, val := range td.details {
		res = append(res, kv.Array(key, val))
	}

	return res
}

type TraceAnnotation func(cfg *TraceDetail)

// WithImpactCreate states that the activity covered in the calling trace event
// created the specified element.
func WithImpactCreate(element string) TraceAnnotation {
	return func(cfg *TraceDetail) {
		cfg.addEntry(ImpactKey, ImpactCreate+":"+element)
	}
}

// WithImpactRead states that the activity covered in the calling trace event
// read the specified element's state.
func WithImpactRead(element string) TraceAnnotation {
	return func(cfg *TraceDetail) {
		cfg.addEntry(ImpactKey, ImpactRead+":"+element)
	}
}

// WithImpactModify states that the activity covered in the calling trace event
// modified the specified element.
func WithImpactModify(element string) TraceAnnotation {
	return func(cfg *TraceDetail) {
		cfg.addEntry(ImpactKey, ImpactModify+":"+element)
	}
}

// WithImpactDelete states that the activity covered in the calling trace event
// deleted the specified element.
func WithImpactDelete(element string) TraceAnnotation {
	return func(cfg *TraceDetail) {
		cfg.addEntry(ImpactKey, ImpactDelete+":"+element)
	}
}

// WithImpactUse states that the activity covered in the calling trace event
// used the specified element as part of its processing.
func WithImpactUse(element string) TraceAnnotation {
	return func(cfg *TraceDetail) {
		cfg.addEntry(ImpactKey, ImpactUse+":"+element)
	}
}

// addAnnotations processes all recognized TraceAnnotation functions in the
// supplied slice, stopping at the first non-TraceAnnotation entry.  The index
// to that first non-Traceannotation entry is then returned.
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

// Info posts an informational trace event
func Info(ctx context.Context, a ...interface{}) {
	cfg := newTraceDetail()
	start := addAnnotations(cfg, a)

	res := append(
		cfg.toKvPairs(),
		kv.Int64(StepperTicksKey, common.TickFromContext(ctx)),
		kv.Int64(SeverityKey, int64(pbl.Severity_Info)),
		kv.String(StackTraceKey, StackTrace()),
		kv.String(MessageTextKey, formatIf(a[start:]...)))

	trace.SpanFromContext(ctx).AddEvent(
		ctx,
		MethodName(2),
		res...)
}

// OnEnter posts an informational trace event describing the entry into a
// function
func OnEnter(ctx context.Context, a ...interface{}) {
	cfg := newTraceDetail()
	start := addAnnotations(cfg, a)

	res := append(
		cfg.toKvPairs(),
		kv.Int64(StepperTicksKey, common.TickFromContext(ctx)),
		kv.Int64(SeverityKey, int64(pbl.Severity_Info)),
		kv.String(StackTraceKey, StackTrace()),
		kv.String(MessageTextKey, formatIf(a[start:]...)))

	trace.SpanFromContext(ctx).AddEvent(
		ctx,
		fmt.Sprintf("On %q entry", MethodName(2)),
		res...)
}

// Warn posts a warning trace event
func Warn(ctx context.Context, a ...interface{}) {
	cfg := newTraceDetail()
	start := addAnnotations(cfg, a)

	res := append(
		cfg.toKvPairs(),
		kv.Int64(StepperTicksKey, common.TickFromContext(ctx)),
		kv.Int64(SeverityKey, int64(pbl.Severity_Warning)),
		kv.String(StackTraceKey, StackTrace()),
		kv.String(MessageTextKey, formatIf(a[start:]...)))

	trace.SpanFromContext(ctx).AddEvent(
		ctx,
		MethodName(2),
		res...)
}

// Error posts an error trace event
func Error(ctx context.Context, a ...interface{}) error {
	cfg := newTraceDetail()
	start := addAnnotations(cfg, a)

	res := append(
		cfg.toKvPairs(),
		kv.Int64(StepperTicksKey, common.TickFromContext(ctx)),
		kv.Int64(SeverityKey, int64(pbl.Severity_Warning)),
		kv.String(StackTraceKey, StackTrace()))

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
				err = errors.New(e)

			default:
				err = fmt.Errorf("unexpected values: %v", e)
			}

		default:
			err = fmt.Errorf(a[start].(string), a[start+1:]...)
		}
	}

	res = append(res, kv.String(MessageTextKey, err.Error()))

	trace.SpanFromContext(ctx).AddEvent(
		ctx,
		fmt.Sprintf("Error from %q", MethodName(3)),
		res...)

	return err
}

// Fatal traces the error, and then terminates the process.
func Fatal(ctx context.Context, a ...interface{}) {
	log.Fatal(Error(ctx, a))
}

// --- Exported trace invocation methods

// +++ Helper functions

// formatIf determines if this is a simple string, or something to format
// before returning.
func formatIf(a ...interface{}) string {
	if a == nil || len(a) == 0 {
		return ""
	}

	if len(a) == 1 {
		if s, ok := a[0].(string); ok {
			return s
		}
		return fmt.Sprintf("%+v", a[0])
	}

	return fmt.Sprintf(a[0].(string), a[1:]...)
}

// --- Helper functions
