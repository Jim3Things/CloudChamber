package tracing

const (
	StepperTicksKey = "cc-stepper-ticks"
	ReasonKey       = "cc-reason"
	MessageTextKey  = "cc-message-text"
	StackTraceKey   = "cc-stack-trace"
	SeverityKey     = "cc-severity"
	ChildSpanKey    = "cc-child-span"

	// Envelope keys
	SourceTraceID   = "cc-starting-span-trace-id"
	SourceSpanID    = "cc-starting-span-span-id"
	SourceTraceFlgs = "cc-starting-span-trace-flags"
)
