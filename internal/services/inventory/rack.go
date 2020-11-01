package inventory

import (
	"context"
	"sync"

	"github.com/Jim3Things/CloudChamber/internal/clients/timestamp"
	"github.com/Jim3Things/CloudChamber/internal/common"
	"github.com/Jim3Things/CloudChamber/internal/sm"
	"github.com/Jim3Things/CloudChamber/internal/tracing"
	pb "github.com/Jim3Things/CloudChamber/pkg/protos/inventory"
)

// rack holds a simulated rack, consisting of a TOR (top of rack switch), a
// PDU (power distribution unit), and some number of blades.  These are
// governed by a mesh of state machines rooted in the one for the rack as a
// whole.
type rack struct {
	name string

	// ch is the channel to send requests along to the rack's goroutine, which
	// is where the state machine runs.
	ch chan sm.Envelope

	tor    *tor
	pdu    *pdu
	blades map[int64]*blade

	sm *sm.SimpleSM

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

type startSim struct {
	envelopeState
}

type stopSim struct {
	envelopeState
}

// newRack creates a new simulated rack using the supplied inventory definition
// entries to determine its structure.  The resulting rack is healthy, not yet
// started, all blades are powered off, and all network connections are not yet
// programmed.
func newRack(ctx context.Context, name string, def *pb.ExternalRack) *rack {
	return newRackInternal(ctx, name, def, newPdu, newTor)
}

// newRackInternal is the implementation behind newRack.  It supports
// dependency injection, to more cleanly allow unit testing of the rack state
// machine logic.
func newRackInternal(
	ctx context.Context,
	name string,
	def *pb.ExternalRack,
	pduFunc func(*pb.ExternalPdu, *rack) *pdu,
	torFunc func(*pb.ExternalTor, *rack) *tor) *rack {
	r := &rack{
		name:      name,
		ch:        make(chan sm.Envelope, rackQueueDepth),
		tor:       nil,
		pdu:       nil,
		blades:    make(map[int64]*blade),
		sm:        nil,
		startLock: sync.Mutex{},
	}

	r.sm = sm.NewSimpleSM(r,
		sm.WithFirstState(rackAwaitingStartState, &rackAwaitingStart{}),
		sm.WithState(rackWorkingState, &rackWorking{}),
		sm.WithState(rackTerminalState, &sm.TerminalState{}),
	)

	r.pdu = pduFunc(def.Pdu, r)
	r.tor = torFunc(def.Tor, r)

	for i, item := range def.Blades {
		r.blades[i] = newBlade(item, r)

		// These two calls are temporary fix-ups until the inventory definition
		// includes the tor and pdu connectors
		r.pdu.fixConnection(ctx, i)
		r.tor.fixConnection(ctx, i)
	}

	return r
}

func (r *rack) viaTor(ctx context.Context, msg sm.Envelope) error {
	tracing.Info(ctx, "Forwarding %v to TOR in rack %q", msg, r.name)

	r.tor.Receive(ctx, msg)

	return nil
}

func (r *rack) viaPDU(ctx context.Context, msg sm.Envelope) error {
	tracing.Info(ctx, "Forwarding '%v' to PDU in rack %q", msg, r.name)
	r.pdu.Receive(ctx, msg)

	return nil
}

// forwardToBlade is a helper function that forwards a message to the target
// blade in this rack.  It returns true if the message was forwarded, false if
// no target blade could be found.
func (r *rack) forwardToBlade(ctx context.Context, id int64, msg sm.Envelope) bool {
	if b, ok := r.blades[id]; ok {
		b.Receive(ctx, msg)
		return true
	}

	return false
}

// start initializes the simulated rack state machine handler, and its state
// machine context.
func (r *rack) start(ctx context.Context) error {
	r.startLock.Lock()
	defer r.startLock.Unlock()

	// Only start the rack state machine once.  If it has already been started
	// then ignore this call.
	if r.sm.CurrentIndex == rackAwaitingStartState {
		go r.simulate()

		repl := make(chan *sm.Response)

		msg := &startSim{}
		msg.Initialize(ctx, repl)

		r.ch <- msg

		res := <-repl

		if res != nil {
			return res.Err
		}
	}

	return ErrRepairMessageDropped
}

// stop terminates the simulated rack state machine, and its handler.
func (r *rack) stop(ctx context.Context) {
	r.startLock.Lock()
	defer r.startLock.Unlock()

	// Issue a stop to terminate the goroutine iff the state
	// machine is still active.
	if r.sm.CurrentIndex != rackAwaitingStartState {
		if !r.sm.Terminated {
			repl := make(chan *sm.Response)

			msg := &stopSim{}
			msg.Initialize(ctx, repl)

			r.ch <- msg

			<-repl
		}
	} else {
		_ = r.sm.ChangeState(ctx, rackTerminalState)
	}
}

// Receive handles incoming requests from outside, forwarding to the rack's
// state machine handler.
func (r *rack) Receive(ctx context.Context, msg sm.Envelope, ch chan *sm.Response) {
	msg.Initialize(ctx, ch)

	r.ch <- msg
}

// simulate is the main function for the rack simulation.
func (r *rack) simulate() {
	for !r.sm.Terminated {
		msg := <-r.ch

		ctx, span := tracing.StartSpan(context.Background(),
			tracing.WithName("Executing simulated inventory operation"),
			tracing.WithNewRoot(),
			tracing.WithLink(msg.GetSpanContext(), msg.GetLinkID()),
			tracing.WithContextValue(timestamp.EnsureTickInContext))

		r.sm.Current.Receive(ctx, r.sm, msg)

		span.End()
	}
}

// +++ rack state machine states

// rackAwaitingStart is the initial state for a rack.  It only expects to be
// started.  All other operations are considered errors.  (Stopping prior to
// starting is handled before it gets to the state machine)
type rackAwaitingStart struct {
	sm.NullState
}

// Receive handles incoming messages for the rack.
func (s *rackAwaitingStart) Receive(ctx context.Context, machine *sm.SimpleSM, msg sm.Envelope) {
	r := machine.Parent.(*rack)
	at := common.TickFromContext(ctx)

	switch body := msg.(type) {
	case *startSim:
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

	default:
		_ = tracing.Error(ctx, "Encountered an unexpected message: %v", body)

		msg.GetCh() <- unexpectedMessageResponse(s, at, body)
	}
}

// Name returns the friendly name for this state.
func (s *rackAwaitingStart) Name() string { return "AwaitingStart" }

// rackWorking is the state where the simulated rack is functional and able to
// take incoming requests.
type rackWorking struct {
	sm.NullState
}

// Receive handles incoming messages for the rack.
func (s *rackWorking) Receive(ctx context.Context, machine *sm.SimpleSM, msg sm.Envelope) {
	r := machine.Parent.(*rack)

	switch body := msg.(type) {
	case *timerExpiry:
		panic("implement me")

	case *stopSim:
		// Stop the rack simulation.
		msg.GetCh() <- &sm.Response{
			Err: machine.ChangeState(ctx, rackTerminalState),
			At:  common.TickFromContext(ctx),
			Msg: nil,
		}

	case repairMessage:
		if err := body.SendVia(ctx, r); err != nil {
			msg.GetCh() <- failedResponse(common.TickFromContext(ctx), err)
		}

	case statusMessage:
		if err := body.SendVia(ctx, r); err != nil {
			msg.GetCh() <- failedResponse(common.TickFromContext(ctx), err)
		}

	default:
		_ = tracing.Error(ctx, "Encountered an unexpected message: %v", body)

		msg.GetCh() <- unexpectedMessageResponse(s, common.TickFromContext(ctx), body)
	}
}

// Name returns the friendly name for this state.
func (s *rackWorking) Name() string { return "working" }

// --- rack state machine states
