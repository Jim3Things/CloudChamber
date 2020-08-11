package tracing

const (
	StepperTicksKey = "cc-stepper-ticks"
	Reason          = "cc-reason"
	MessageTextKey  = "cc-message-text"
	StackTraceKey   = "cc-stack-trace"
	SeverityKey     = "cc-severity"

	// Header keys, used to add context on the way to the actor
	LinkingSpanID   = "cc-link-to"

	// Envelope keys
	SourceTraceID   = "cc-starting-span-trace-id"
	SourceSpanID    = "cc-starting-span-span-id"
	SourceTraceFlgs = "cc-starting-span-trace-flags"
)
