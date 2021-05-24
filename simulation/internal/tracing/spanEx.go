package tracing

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/api/kv"
	"go.opentelemetry.io/otel/api/trace"

	"github.com/Jim3Things/CloudChamber/simulation/internal/common"
	pbl "github.com/Jim3Things/CloudChamber/simulation/pkg/protos/log"
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

func (s SpanEx) End(opt ...trace.EndOption) {
	var opts []trace.EndOption
	opts = append(opts, trace.WithEndTime(time.Now()))
	opts = append(opts, opt...)

	s.Span.End(opt...)
}

type decorator func(ctx context.Context) context.Context

// Mask applies the policy that a child span of a span that is marked as
// infrastructure (span kind is internal) should also be market as internal.
func (s SpanEx) Mask(kind trace.SpanKind) trace.SpanKind {
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
	impact      string
}

// toKvPairs converts all span configuration fields that are implemented as
// span attributes to  the appropriate kv-pair values.  The array that is
// returned can then be passed directly to the start span call using the
// standard WithAttributes helper.
func (cfg *startSpanConfig) toKvPairs() []kv.KeyValue {
	var res []kv.KeyValue

	if len(cfg.linkTag) > 0 {
		res = append(res, kv.String(LinkTagKey, cfg.linkTag))
	}

	if len(cfg.impact) > 0 {
		res = append(res, kv.String(ImpactKey, cfg.impact))
	}

	if len(cfg.reason) > 0 {
		res = append(res, kv.String(ReasonKey, cfg.reason))
	}

	res = append(res, kv.String(StackTraceKey, cfg.stackTrace))

	return res
}

// StartSpanOption denotes optional decoration methods used on StartSpan
type StartSpanOption func(*startSpanConfig)

// WithName adds the supplied value as the name of the span under creation
func WithName(f string, a ...interface{}) StartSpanOption {
	return func(cfg *startSpanConfig) {
		cfg.name = fmt.Sprintf(f, a...)
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

// WithImpact states that the activity covered in the calling trace event had
// the specified impact on the specified element.
func WithImpact(impact string, element string) StartSpanOption {
	return func(cfg *startSpanConfig) {
		cfg.impact = impact+":"+element
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
	options ...StartSpanOption) (context.Context, SpanEx) {
	cfg := startSpanConfig{
		name:        MethodName(2),
		kind:        trace.SpanKindServer,
		stackTrace:  StackTrace(),
		tick:        -1,
		decorations: []decorator{},
		link:        trace.SpanContext{},
		linkTag:     "",
		newRoot:     false,
		impact:      "",
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
		trace.WithStartTime(time.Now()),
		trace.WithSpanKind(cfg.kind),
		mayLinkTo(cfg.link),
		mayNewRoot(cfg.newRoot),
		trace.WithAttributes(cfg.toKvPairs()...))

	if !cfg.newRoot && parent.SpanContext().HasSpanID() {
		parent.AddEventWithTimestamp(
			ctxChild,
			time.Now(),
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
	trace.SpanFromContext(ctx).SetAttribute(SpanNameKey, formatIf(a...))
}

// UpdateSpanReason replaces the current span reason with the formatted
// string provided.  The span will end up with the last reason provided.
func UpdateSpanReason(ctx context.Context, a ...interface{}) {
	trace.SpanFromContext(ctx).SetAttribute(ReasonKey, formatIf(a...))
}

// AddImpact adds an impact claim to the current span.  It does not optimize
// out any duplicates -- all of those are retained.
func AddImpact(ctx context.Context, level string, module string) {
	trace.SpanFromContext(ctx).AddEventWithTimestamp(
		ctx,
		time.Now(),
		MethodName(2),
		kv.Int64(ActionKey, int64(pbl.Action_AddImpact)),
		kv.Int64(StepperTicksKey, common.TickFromContext(ctx)),
		kv.Int64(SeverityKey, int64(pbl.Severity_Info)),
		kv.String(StackTraceKey, StackTrace()),
		kv.String(MessageTextKey, level+":"+module))
}

// AddLink adds an event that marks the point where a request was made that may
// result in a related span.  This contains the unique link tag that the target
// span should also provide, with the intention that a structured formatter can
// place that target span in the correct place in the execution sequence.
func AddLink(ctx context.Context, tag string) {
	trace.SpanFromContext(ctx).AddEventWithTimestamp(
		ctx,
		time.Now(),
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
		} else {
			return fmt.Sprintf("%+v", a[0])
		}
	}

	return fmt.Sprintf(a[0].(string), a[1:]...)
}

// -- Helper functions
