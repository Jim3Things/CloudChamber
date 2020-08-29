package tracing

const (
	StepperTicksKey   = "cc-stepper-ticks"
	Reason            = "cc-reason"
	MessageTextKey    = "cc-message-text"
	StackTraceKey     = "cc-stack-trace"
	SeverityKey       = "cc-severity"

	// Envelope keys
	SourceTraceID     = "cc-starting-span-trace-id"
	SourceSpanID      = "cc-starting-span-span-id"
	SourceTraceFlgs   = "cc-starting-span-trace-flags"
)
