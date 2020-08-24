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
)
