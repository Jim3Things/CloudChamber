package client

import (
	"github.com/Jim3Things/CloudChamber/pkg/protos/log"
)

type override struct {
	matchMethod   string
	matchSeverity log.Severity
	severity      log.Severity
}

// overrides is the table that lists the GRPC method names that are to have
// their severity replaced by the one in the table, but only if the incoming
// severity also matches.  This is used to designate some internal only GRPC
// calls as debug-only events, so as to not fill the UI log stream with cruft.
var overrides = []override{
	{"/services.Stepper/Now", log.Severity_Info, log.Severity_Debug},
}

// overrideSeverity looks in the table of method names that are designated to
// have specific severity overrides.  If a match with the supplied method name
// and with the supplied severity is found, the returned severity is from the
// override value, otherwise it is the one supplied.
func overrideSeverity(method string, sev log.Severity) log.Severity {
	for _, o := range overrides {
		if o.matchMethod == method && o.matchSeverity == sev {
			return o.severity
		}
	}

	return sev
}
