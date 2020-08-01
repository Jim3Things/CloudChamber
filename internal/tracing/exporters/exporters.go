package exporters

const (
	StdOut = iota
	UnitTest

	// Production forwards the log calls to a sink that another caller can
	// retrieve them from for display on the UI
	Production

	// LocalProduction is a shortcut used by the process that hosts the
	// production trace sink to avoid appends to the trace log recursively
	// expanding forever - and to still be able to trace those calls.
	LocalProduction
)
