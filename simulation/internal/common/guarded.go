package common

// Guarded provides basic support for access gated by last sequence value.  It
// maintains the last sequence number used, and has functions that can be used
// to refuse access based on out of order sequencing, as well as the maintenance
// of the guard value itself.
type Guarded struct {
	// Guard is the last sequence number used
	Guard int64
}

// Pass checks that the provided check value is not earlier than the guard.  If
// it passes, the guard is updated with supplied current sequence value from
// the at parameter.  This allows for checks based on values the caller last
// saw, and for updates to provide something close to a linear sequence across
// guarded objects.  Pass returns true if the check passes, and false if it
// didn't.
func (g *Guarded) Pass(check int64, at int64) bool {
	if check < g.Guard {
		return false
	}

	g.Guard = at
	return true
}

// AdvanceGuard updates the guard value if the new value is greater than the
// value the object already holds.
func (g *Guarded) AdvanceGuard(at int64) {
	g.Guard = MaxInt64(g.Guard, at)
}

