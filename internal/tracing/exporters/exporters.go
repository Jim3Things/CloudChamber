package exporters

const (
	// IoWriter formats and writes the trace calls to a specified trace file
	IoWriter = iota

	// UnitTest formats and writes the trace calls to the unit test framework's
	// trace channel
	UnitTest

	// Production forwards the trace calls to a sink that another caller can
	// retrieve them from for display on the UI
	Production

	// LocalProduction is a shortcut used by the process that hosts the
	// production trace sink to avoid appends to the trace log recursively
	// expanding forever - and to still be able to trace those calls.
	LocalProduction
)
