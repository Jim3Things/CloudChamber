package inventory

import (
	"context"

	"github.com/Jim3Things/CloudChamber/internal/clients/timestamp"
	"github.com/Jim3Things/CloudChamber/internal/common"
	"github.com/Jim3Things/CloudChamber/internal/sm"
	"github.com/Jim3Things/CloudChamber/internal/tracing"
	pb "github.com/Jim3Things/CloudChamber/pkg/protos/inventory"
	"github.com/Jim3Things/CloudChamber/pkg/protos/services"
)

// rack holds a simulated rack, consisting of a TOR (top of rack switch), a
// PDU (power distribution unit), and some number of blades.  These are
// governed by a mesh of state machines rooted in the one for the rack as a
// whole.
type rack struct {
	name string

	// ch is the channel to send requests along to the rack's goroutine, which
	// is where the state machine runs.
	ch chan *sm.Envelope

	tor    *tor
	pdu    *pdu
	blades map[int64]*blade

	sm *sm.SimpleSM
}

const (
	rackAwaitingStartState int = iota
	rackWorkingState
	rackTerminalState
)

const (
	rackQueueDepth = 100
)

type startSim struct{}

type stopSim struct{}

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
		name:   name,
		ch:     make(chan *sm.Envelope, rackQueueDepth),
		tor:    nil,
		pdu:    nil,
		blades: make(map[int64]*blade),
		sm:     nil,
	}

	r.sm = sm.NewSimpleSM(r,
		sm.WithFirstState(rackAwaitingStartState, &rackAwaitingStart{}),
		sm.WithState(rackWorkingState, &rackWorking{}),
		sm.WithState(rackTerminalState, &sm.TerminalState{}),
	)

	r.pdu = pduFunc(def.Pdu, r)
	r.tor = torFunc(def.Tor, r)

	for i, item := range def.Blades {
		r.blades[i] = newBlade(item)

		// These two calls are temporary fix-ups until the inventory definition
		// includes the tor and pdu connectors
		r.pdu.fixConnection(ctx, i)
		r.tor.fixConnection(ctx, i)
	}

	return r
}

// start initializes the simulated rack state machine handler, and its state
// machine context.
func (r *rack) start(ctx context.Context) error {
	// TODO: Need a guard to detect and handle multiple start calls.

	go r.simulate()

	repl := make(chan *sm.Response)

	r.ch <- sm.NewEnvelope(ctx, &startSim{}, repl)

	res := <-repl

	if res != nil {
		return res.Err
	}

	return ErrRepairMessageDropped
}

// stop terminates the simulated rack state machine, and its handler.
func (r *rack) stop(ctx context.Context) {
	repl := make(chan *sm.Response)

	r.ch <- sm.NewEnvelope(ctx, &stopSim{}, repl)

	<-repl
}

// Receive handles incoming requests from outside, forwarding to the rack's
// state machine handler.
func (r *rack) Receive(ctx context.Context, msg interface{}, ch chan *sm.Response) {
	r.ch <- sm.NewEnvelope(ctx, msg, ch)
}

// simulate is the main function for the rack simulation.
func (r *rack) simulate() {
	for {
		msg := <-r.ch

		ctx, span := tracing.StartSpan(context.Background(),
			tracing.WithName("Executing simulated inventory operation"),
			tracing.WithNewRoot(),
			tracing.WithLink(msg.Span, msg.Link),
			tracing.WithContextValue(timestamp.EnsureTickInContext))

		r.sm.Current.Receive(ctx, r.sm, msg)

		span.End()
	}
}

// +++ rack state machine states

// rackAwaitingStart is the initial state for a rack.  It only expects to be
// started or stopped.  All other operations are considered errors.
type rackAwaitingStart struct {
	sm.NullState
}

func (s *rackAwaitingStart) Receive(ctx context.Context, machine *sm.SimpleSM, msg *sm.Envelope) {
	r := machine.Parent.(*rack)
	at := common.TickFromContext(ctx)

	switch body := msg.Msg.(type) {
	case *startSim:
		tracing.UpdateSpanName(
			ctx,
			"Starting the simulation of rack %q",
			r.name)

		err := r.sm.Start(ctx)
		if err == nil {
			err = machine.ChangeState(ctx, rackWorkingState)
		}

		msg.Ch <- &sm.Response{
			Err: err,
			At:  at,
			Msg: nil,
		}

	case *stopSim:
		tracing.UpdateSpanName(
			ctx,
			"Stopping the simulation of rack %q",
			r.name)

		msg.Ch <- &sm.Response{
			Err: machine.ChangeState(ctx, rackTerminalState),
			At:  at,
			Msg: nil,
		}

	default:
		_ = tracing.Error(ctx, "Encountered an unexpected message: %v", body)

		msg.Ch <- unexpectedMessageResponse(s, at, body)
	}
}

func (s *rackAwaitingStart) Name() string { return "AwaitingStart" }

// rackWorking is the state where the simulated rack is functional and able to
// take incoming requests.
type rackWorking struct {
	sm.NullState
}

func (s *rackWorking) Receive(ctx context.Context, machine *sm.SimpleSM, msg *sm.Envelope) {
	r := machine.Parent.(*rack)

	switch body := msg.Msg.(type) {
	case *services.InventoryStatusMsg:
		// Get the status a simulated element.  Note that the rack itself
		// currently has no status, so this request is always forwarded
		// to one or more contained elements.
		addr := body.Target

		switch addr.Element.(type) {
		case *services.InventoryAddress_Pdu:
			// The PDU is the one component that we can access without
			// using the TOR as a network hop.  This reflects an
			// intentional simplification in the simulation: that the
			// TOR cannot disconnect the PDU, and the PDU cannot stop
			// the TOR.  In real life, these are probably possible, but
			// for now we simplify our world by avoiding that.
			r.pdu.Receive(ctx, msg)

		default:
			// By default, all messages are routed through the TOR
			r.tor.Receive(ctx, msg)
		}

	case *services.InventoryRepairMsg:
		// Repair an element in the rack.
		switch body.GetAction().(type) {
		case *services.InventoryRepairMsg_Power:
			// All power operations are routed through the PDU
			r.pdu.Receive(ctx, msg)

		default:
			// By default, all messages are routed through the TOR
			r.tor.Receive(ctx, msg)
		}

	case *stopSim:
		// Stop the rack simulation.
		msg.Ch <- &sm.Response{
			Err: machine.ChangeState(ctx, rackTerminalState),
			At:  common.TickFromContext(ctx),
			Msg: nil,
		}

	default:
		msg.Ch <- unexpectedMessageResponse(s, common.TickFromContext(ctx), body)
	}
}

// Name returns the friendly name for this state.
func (s *rackWorking) Name() string { return "working" }

// --- rack state machine states

// forwardToBlade is a helper function that forwards a message to the target
// blade in this rack.
func (r *rack) forwardToBlade(ctx context.Context, id int64, msg *sm.Envelope) {
	if b, ok := r.blades[id]; ok {
		b.Receive(ctx, msg)
	}
}
