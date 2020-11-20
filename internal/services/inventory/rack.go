package inventory

import (
	"context"
	"fmt"
	"sync"

	"github.com/Jim3Things/CloudChamber/internal/clients/timestamp"
	"github.com/Jim3Things/CloudChamber/internal/common"
	"github.com/Jim3Things/CloudChamber/internal/services/inventory/messages"
	"github.com/Jim3Things/CloudChamber/internal/sm"
	"github.com/Jim3Things/CloudChamber/internal/tracing"
	pb "github.com/Jim3Things/CloudChamber/pkg/protos/inventory"
)

// Rack holds a simulated Rack, consisting of a TOR (top of Rack switch), a
// PDU (power distribution unit), and some number of blades.  These are
// governed by a mesh of state machines rooted in the one for the Rack as a
// whole.
type Rack struct {
	name string

	// ch is the channel to send requests along to the Rack's goroutine, which
	// is where the state machine runs.
	ch chan sm.Envelope

	tor    *tor
	pdu    *pdu
	blades map[int64]*blade

	sm *sm.SimpleSM

	timers *timestamp.Timers

	// startLock controls access to start and stop operations, and therefore to
	// the setup and tear down of the lister goroutine.
	startLock sync.Mutex
}

const (
	rackAwaitingStartState int = iota
	rackWorkingState
	rackTerminalState
)

const (
	rackQueueDepth = 100
)

// newRack creates a new simulated Rack using the supplied inventory definition
// entries to determine its structure.  The resulting Rack is healthy, not yet
// started, all blades are powered off, and all network connections are not yet
// programmed.
func newRack(ctx context.Context, name string, def *pb.ExternalRack, timers *timestamp.Timers) *Rack {
	return newRackInternal(ctx, name, def, timers, newPdu, newTor)
}

// newRackInternal is the implementation behind newRack.  It supports
// dependency injection, to more cleanly allow unit testing of the Rack state
// machine logic.
func newRackInternal(
	ctx context.Context,
	name string,
	def *pb.ExternalRack,
	timers *timestamp.Timers,
	pduFunc func(*pb.ExternalPdu, *Rack) *pdu,
	torFunc func(*pb.ExternalTor, *Rack) *tor) *Rack {
	r := &Rack{
		name:      name,
		ch:        make(chan sm.Envelope, rackQueueDepth),
		tor:       nil,
		pdu:       nil,
		blades:    make(map[int64]*blade),
		sm:        nil,
		timers:    timers,
		startLock: sync.Mutex{},
	}

	r.sm = sm.NewSimpleSM(r,
		sm.WithFirstState(
			rackAwaitingStartState,
			"awaiting-start",
			sm.NullEnter,
			[]sm.ActionEntry{
				{messages.TagStartSim, startSim, rackWorkingState, rackTerminalState},
			},
			UnexpectedMessage,
			sm.NullLeave),

		sm.WithState(
			rackWorkingState,
			"working",
			sm.NullEnter,
			[]sm.ActionEntry{
				{messages.TagSetConnection, process, sm.Stay, sm.Stay},
				{messages.TagSetPower, process, sm.Stay, sm.Stay},
				{messages.TagTimerExpiry, process, sm.Stay, sm.Stay},
				{messages.TagStopSim, stopSim, rackTerminalState, sm.Stay},
			},
			UnexpectedMessage,
			sm.NullLeave),

		sm.WithState(
			rackTerminalState,
			"terminated",
			sm.TerminalEnter,
			[]sm.ActionEntry{},
			DropMessage,
			sm.NullLeave),
	)

	r.pdu = pduFunc(def.Pdu, r)
	r.tor = torFunc(def.Tor, r)

	for i, item := range def.Blades {
		r.blades[i] = newBlade(item, r, i)

		// These two calls are temporary fix-ups until the inventory definition
		// includes the tor and pdu connectors
		r.pdu.fixConnection(ctx, i)
		r.tor.fixConnection(ctx, i)
	}

	return r
}

// ViaTor forwards the supplied message to the Rack's simulated TOR for
// processing.  This will likely not be the final destination, but requires
// operation by the TOR in order to reach its final destination.
func (r *Rack) ViaTor(ctx context.Context, msg sm.Envelope) error {
	tracing.Info(ctx, "Forwarding %v to TOR in rack %q", msg, r.name)

	r.tor.Receive(ctx, msg)

	return nil
}

// ViaPDU forwards the supplied message to the Rack's simulated PDU for
// processing.  This may or may not impact the full PDU and the all blades,
// or only one blade's power state.
func (r *Rack) ViaPDU(ctx context.Context, msg sm.Envelope) error {
	tracing.Info(ctx, "Forwarding '%v' to PDU in rack %q", msg, r.name)
	r.pdu.Receive(ctx, msg)

	return nil
}

// ViaBlade forwards the supplied message directly to the target blade, without
// any intermediate hops.  This should only be used by events that do not need
// to simulate a working network connection for reachability, or a working power
// cable for execution.
func (r *Rack) ViaBlade(ctx context.Context, id int64, msg sm.Envelope) error {
	tracing.Info(ctx, "Forwarding '%v' to blade %d in rack %q", msg, id, r.name)
	if b, ok := r.blades[id]; ok {
		b.Receive(ctx, msg)
		return nil
	}

	return messages.ErrInvalidTarget
}

// forwardToBlade is a helper function that forwards a message to the target
// blade in this Rack.  It returns true if the message was forwarded, false if
// no target blade could be found.
func (r *Rack) forwardToBlade(ctxIn context.Context, id int64, msg sm.Envelope) bool {
	if b, ok := r.blades[id]; ok {
		ctx, span := tracing.StartSpan(
			ctxIn,
			tracing.WithName(fmt.Sprintf("Processing message %q on blade", msg)),
			tracing.WithNewRoot(),
			tracing.WithLink(msg.GetSpanContext(), msg.GetLinkID()),
			tracing.WithContextValue(timestamp.EnsureTickInContext))
		defer span.End()

		b.Receive(ctx, msg)
		return true
	}

	return false
}

// setTimer registers for a notification at a future point in simulated time.
// When it expires, the supplied message is delivered for processing.
func (r *Rack) setTimer(ctx context.Context, delay int64, msg sm.Envelope) (int, error) {
	return r.timers.Timer(ctx, delay, msg, func(msg interface{}) {
		m := msg.(sm.Envelope)
		r.Receive(m)
	})
}

// cancelTimer attempts to cancel a previously registered timer.
func (r *Rack) cancelTimer(id int) error {
	return r.timers.Cancel(id)
}

// start initializes the simulated Rack state machine handler, and its state
// machine context.
func (r *Rack) start(ctx context.Context) error {
	r.startLock.Lock()
	defer r.startLock.Unlock()

	// Only start the rack state machine once.  If it has already been started
	// then ignore this call.
	if r.sm.CurrentIndex == rackAwaitingStartState {
		go r.simulate()

		repl := make(chan *sm.Response)

		msg := messages.NewStartSim(ctx, repl)

		r.ch <- msg

		res := <-repl

		if res != nil {
			return res.Err
		}
	}

	return messages.ErrRepairMessageDropped
}

// stop terminates the simulated Rack state machine, and its handler.
func (r *Rack) stop(ctx context.Context) {
	r.startLock.Lock()
	defer r.startLock.Unlock()

	// Issue a stop to terminate the goroutine iff the state
	// machine is still active.
	if r.sm.CurrentIndex != rackAwaitingStartState {
		if !r.sm.Terminated {
			repl := make(chan *sm.Response)

			msg := messages.NewStopSim(ctx, repl)

			r.ch <- msg

			<-repl
		}
	} else {
		_ = r.sm.ChangeState(ctx, rackTerminalState)
	}
}

// Receive handles incoming requests from outside, forwarding to the rack's
// state machine handler.
func (r *Rack) Receive(msg sm.Envelope) {
	r.ch <- msg
}

// simulate is the main function for the Rack simulation.
func (r *Rack) simulate() {
	for !r.sm.Terminated {
		msg := <-r.ch

		ctx, span := tracing.StartSpan(
			context.Background(),
			tracing.WithName("Executing simulated inventory operation"),
			tracing.WithNewRoot(),
			tracing.WithLink(msg.GetSpanContext(), msg.GetLinkID()),
			tracing.WithContextValue(timestamp.EnsureTickInContext))

		r.sm.Current.Receive(ctx, r.sm, msg)

		span.End()
	}
}

// +++ rack state machine states

func startSim(ctx context.Context, machine *sm.SimpleSM, m sm.Envelope) bool {
	r := machine.Parent.(*Rack)
	at := common.TickFromContext(ctx)

	msg := m.(*messages.StartSim)

	tracing.UpdateSpanName(
		ctx,
		"Starting the simulation of rack %q",
		r.name)

	err := r.sm.Start(ctx)
	if err == nil {
		err = machine.ChangeState(ctx, rackWorkingState)
	}

	msg.GetCh() <- &sm.Response{
		Err: err,
		At:  at,
		Msg: nil,
	}

	return err == nil
}

func process(ctx context.Context, machine *sm.SimpleSM, msg sm.Envelope) bool {
	body := msg.(messages.RepairMessage)
	r := machine.Parent.(*Rack)

	if err := body.SendVia(ctx, r); err != nil {
		msg.GetCh() <- messages.FailedResponse(common.TickFromContext(ctx), err)
	}

	return true
}

func stopSim(ctx context.Context, machine *sm.SimpleSM, msg sm.Envelope) bool {
	// Stop the rack simulation.
	msg.GetCh() <- &sm.Response{
		Err: machine.ChangeState(ctx, rackTerminalState),
		At:  common.TickFromContext(ctx),
		Msg: nil,
	}

	return true
}

// --- rack state machine states
