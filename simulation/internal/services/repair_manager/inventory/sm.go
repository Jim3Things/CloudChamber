// Contains the entry into the state machine used by the irm

// Each inventory item is defined by a triplet of storage entries: one for the
// target state, one for the last known state, and an internal scratchpad of
// in-progress details.
//
// Each of these is modified by one source.  The target state is only modified
// as a result of administrative action, such as commands to add or remove known
// inventory.  The actual state is only modified as the result of notifications
// to or observations by the inventory monitor.  The internal state is only
// modified by the repair manager.
//
// There is one exception to the modification rules above - final removal of an
// entry is driven by the repair manager.
//
// This triad is used to provide the state machine context for what is, in effect,
// an inventory instance actor.
//
// The state machine is triggered by one of two events: a change notification
// for either the target or actual states, or by a timer expiry.  Timers may be
// issued by the state machine when it triggers some action that should produce
// an effect within a specific period.

package inventory

import m "github.com/Jim3Things/CloudChamber/simulation/pkg/protos/inventory"

// Each state in the state machine implements this interface.

// Note that the signatures below describe a state machine approach that always sends
// outgoing messages before the updated state is stored.  This is sufficient so long as the
// calculations are safe and cheap to repeat.  If we find that we need to store state ahead
// of a message, then these signatures should be changed to indicate that the state machine
// needs to immediately re-run (without pulling from the channel).  That would allow the
// creation of 'intent' states that do the calculations and store the intent and then
// transition to the state that sends the messages.

type State interface {
	Dispatch(target *m.Target, actual *m.Actual, state *m.Internal) (*m.Target, *m.Actual, *m.Internal, error)
	TimerExpired(name string, target *m.Target, actual *m.Actual, state *m.Internal) (*m.Target, *m.Actual, *m.Internal, error)
}

// TODO: The assumption here is that each state machine operates in a serialized environment -
//       presumed to be a goroutine at the end of a channel.  That mechanism exists in front
//       of this structure, and is yet to be determined.
//
//       State is read before calling the state machine, and stored after the call returns.
//
//       There are a few options around actor lifetime management.  The simplest is to create
//       the goroutine either on restart or on a create call, and then keep it active. That
//       has issues with scaling - there could be a very high number of goroutines, of which
//       almost none are doing anything at any given time.
//
//       A second approach is to suspend the actor whenever there is no outstanding entry to
//       process.  This requires some interlocking to prevent a second operation from starting
//		 while the first is still executing.
//
//		 A third approach is to suspend the actor when there have been no outstanding entries
//		 in the channel for some time.  This avoids the constant restart/suspend cycles of the
//		 second option.  It uses the inherent semantics of a channel to handle overlapping
//		 operations.  But it does requires interlocking to avoid suspending an actor just as a
//		 new operation is put into the input channel...
//
//		 This actor & channel lifecycle management needs to be put into its own module so that
//		 the details can be hidden.  My current suggestion is to start with the first option to
//		 unblock the dispatch, and then to move to the third option.

// Called when a change event is recorded against either the target or actual entries.
func Dispatch(target *m.Target, actual *m.Actual, state *m.Internal) (*m.Target, *m.Actual, *m.Internal, error) {
	s := getState(target, actual, state)
	return s.Dispatch(target, actual, state)
}

// Some operations have timeouts associated with them. This is the method called when such a timer
// expires.
func TimerExpired(name string, target *m.Target, actual *m.Actual, state *m.Internal) (*m.Target, *m.Actual, *m.Internal, error) {
	s := getState(target, actual, state)
	return s.TimerExpired(name, target, actual, state)
}

// Calculate the correct state for the workload state machine, and return that state's (pure) instance
// TODO: This is nothing but the wrapper at this point
func getState(_ *m.Target, _ *m.Actual, _ *m.Internal) State {
	return nil
}
