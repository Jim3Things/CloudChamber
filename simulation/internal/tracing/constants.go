package tracing

const (
	StepperTicksKey = "cc-stepper-ticks"
	ReasonKey       = "cc-reason"
	MessageTextKey  = "cc-message-text"
	StackTraceKey   = "cc-stack-trace"
	SeverityKey     = "cc-severity"
	ChildSpanKey    = "cc-child-span"
	ActionKey       = "cc-action"
	ImpactKey       = "cc-impact"
	SpanNameKey     = "cc-span-name"

	// link tracking keys
	LinkTagKey = "cc-link-tag"

	// rpc injection keys
	InfraSourceKey = "cc-infra"

	// rpc infra source key values
	IsInfraSource    = "infra"
	IsNotInfraSource = "not" + IsInfraSource
)

// impact data is stored in a KeyValue pair under the ImpactKey.  KV pairs have
// limited types that they support, so the impact targets are prefixed with the
// type of impact.  This leaves the value as a string that can be decoded later
// in order to properly display the impact.
const (
	ImpactCreate = "C"
	ImpactRead   = "R"
	ImpactModify = "W"
	ImpactDelete = "D"
	ImpactUse    = "E"
)
