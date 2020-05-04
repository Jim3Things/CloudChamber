package tracing

const (
	StepperTicksKey = "cc-stepper-ticks"
	Reason          = "cc-reason"
	StackTraceKey   = "cc-stack-trace"

	// Header keys, used to add context on the way to the actor
	LinkingSpanID   = "cc-link-to"
)
