package inventory

import (
	"context"
	"fmt"
	"sync"

	"github.com/Jim3Things/CloudChamber/simulation/internal/clients/timestamp"
	"github.com/Jim3Things/CloudChamber/simulation/internal/common"
	"github.com/Jim3Things/CloudChamber/simulation/internal/services/inventory/messages"
	"github.com/Jim3Things/CloudChamber/simulation/internal/sm"
	"github.com/Jim3Things/CloudChamber/simulation/internal/tracing"
	"github.com/Jim3Things/CloudChamber/simulation/pkg/errors"
	pb "github.com/Jim3Things/CloudChamber/simulation/pkg/protos/inventory"
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

	sm *sm.SM

	timers *timestamp.Timers

	// startLock controls access to start and stop operations, and therefore to
	// the setup and tear down of the lister goroutine.
	startLock sync.Mutex
}

const (
	rackQueueDepth = 100
)

// newRack creates a new simulated Rack using the supplied inventory definition
// entries to determine its structure.  The resulting Rack is healthy, not yet
// started, all blades are powered off, and all network connections are not yet
// programmed.
func newRack(ctx context.Context, name string, def *pb.External_Rack, timers *timestamp.Timers) *Rack {
	return newRackInternal(ctx, name, def, timers, newPdu, newTor)
}

// newRackInternal is the implementation behind newRack.  It supports
// dependency injection, to more cleanly allow unit testing of the Rack state
// machine logic.
func newRackInternal(
	ctx context.Context,
	name string,
	def *pb.External_Rack,
	timers *timestamp.Timers,
	pduFunc func(*pb.External_Pdu, *Rack) *pdu,
	torFunc func(*pb.External_Tor, *Rack) *tor) *Rack {
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

	r.sm = sm.NewSM(r,
		sm.WithFirstState(
			pb.Actual_Rack_awaiting_start,
			sm.NullEnter,
			[]sm.ActionEntry{
				{sm.TagStartSM, startSim, pb.Actual_Rack_working, pb.Actual_Rack_terminated},
			},
			sm.UnexpectedMessage,
			sm.NullLeave),

		sm.WithState(
			pb.Actual_Rack_working,
			sm.NullEnter,
			[]sm.ActionEntry{
				{messages.TagGetStatus, process, sm.Stay, sm.Stay},
				{messages.TagSetConnection, process, sm.Stay, sm.Stay},
				{messages.TagSetPower, process, sm.Stay, sm.Stay},
				{messages.TagTimerExpiry, process, sm.Stay, sm.Stay},
				{sm.TagStopSM, stopSim, pb.Actual_Rack_terminated, sm.Stay},
			},
			sm.UnexpectedMessage,
			sm.NullLeave),

		sm.WithState(
			pb.Actual_Rack_terminated,
			sm.TerminalEnter,
			[]sm.ActionEntry{},
			messages.DropMessage,
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

	return errors.ErrInvalidTarget
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
			tracing.WithLink(msg.SpanContext(), msg.LinkID()),
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
	if r.sm.CurrentIndex == pb.Actual_Rack_awaiting_start {
		go r.simulate()

		repl := make(chan *sm.Response)

		msg := sm.NewStartSM(ctx, repl)

		r.ch <- msg

		res := <-repl

		if res != nil {
			return res.Err
		}
	}

	return errors.ErrAlreadyStarted
}

// stop terminates the simulated Rack state machine, and its handler.
func (r *Rack) stop(ctx context.Context) {
	r.startLock.Lock()
	defer r.startLock.Unlock()

	// Issue a stop to terminate the goroutine iff the state
	// machine is still active.
	if r.sm.CurrentIndex != pb.Actual_Rack_awaiting_start {
		if !r.sm.Terminated {
			repl := make(chan *sm.Response)

			msg := sm.NewStopSM(ctx, repl)

			r.ch <- msg

			<-repl
		}
	} else {
		_ = r.sm.ChangeState(ctx, pb.Actual_Rack_terminated)
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
			tracing.WithLink(msg.SpanContext(), msg.LinkID()),
			tracing.WithContextValue(timestamp.EnsureTickInContext))

		r.sm.Current.Receive(ctx, r.sm, msg)

		span.End()
	}
}

// +++ rack state machine actions

// startSim starts the rack simulation state machine, and all those of all the
// elements contained within the rack.
func startSim(ctx context.Context, machine *sm.SM, m sm.Envelope) bool {
	r := machine.Parent.(*Rack)
	at := common.TickFromContext(ctx)

	msg := m.(*sm.StartSM)

	ch := msg.Ch()
	defer close(ch)

	tracing.UpdateSpanName(
		ctx,
		"Starting the simulation of rack %q",
		r.name)

	err := r.sm.Start(ctx)

	if err == nil {
		err = r.pdu.sm.Start(ctx)
	}

	if err == nil {
		err = r.tor.sm.Start(ctx)
	}

	for _, b := range r.blades {
		if err != nil {
			break
		}

		err = b.sm.Start(ctx)
	}

	ch <- &sm.Response{
		Err: err,
		At:  at,
		Msg: nil,
	}

	return err == nil
}

// process an incoming message, forwarding to the relevant managed element.
func process(ctx context.Context, machine *sm.SM, msg sm.Envelope) bool {
	body := msg.(messages.RepairMessage)
	r := machine.Parent.(*Rack)

	if err := body.SendVia(ctx, r); err != nil {
		// Forwarding failed, so issue the response here (as no one else could
		// have handled it)
		ch := msg.Ch()
		ch <- sm.FailedResponse(common.TickFromContext(ctx), err)
		close(ch)
	}

	return true
}

// stopSim is used to stop the rack simulation, and signal that it is now done.
func stopSim(ctx context.Context, machine *sm.SM, msg sm.Envelope) bool {
	ch := msg.Ch()
	defer close(ch)

	// Stop the rack simulation.
	ch <- &sm.Response{
		Err: machine.ChangeState(ctx, pb.Actual_Rack_terminated),
		At:  common.TickFromContext(ctx),
		Msg: nil,
	}

	return true
}

// --- rack state machine actions
