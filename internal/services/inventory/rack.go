package inventory

import (
	"context"

	"go.opentelemetry.io/otel/api/trace"

	"github.com/Jim3Things/CloudChamber/internal/clients/timestamp"
	"github.com/Jim3Things/CloudChamber/internal/common"
	"github.com/Jim3Things/CloudChamber/internal/sm"
	"github.com/Jim3Things/CloudChamber/internal/tracing"
	pb "github.com/Jim3Things/CloudChamber/pkg/protos/inventory"
	"github.com/Jim3Things/CloudChamber/pkg/protos/services"
)

type rack struct {
	// ch is the channel to send requests along to the rack's goroutine, which
	// is where the state machine runs.
	ch chan *envelope

	tor *tor
	pdu *pdu
	blades map[int64]*blade

	sm *sm.SimpleSM
}

const (
	rackAwaitingStartState int = iota
	rackWorkingState
	rackDisabledState
	rackFailedState
	rackTerminalState
)

const (
	rackQueueDepth = 100
)

type envelope struct {
	ch chan *sm.Response
	span trace.SpanContext
	msg interface{}
}

type startSim struct {}

type stopSim struct {}

func newRack(ctx context.Context, def *pb.ExternalRack) *rack {
	return newRackInternal(ctx, def, newPdu, newTor)
}

func newRackInternal(
	ctx context.Context,
	def *pb.ExternalRack,
	pduFunc func(*pb.ExternalPdu, *rack) *pdu,
	torFunc func(*pb.ExternalTor, *rack) *tor) *rack {
	r := &rack{
		ch:     make(chan *envelope, rackQueueDepth),
		tor:    nil,
		pdu:    nil,
		blades: make(map[int64]*blade),
		sm:     nil,
	}

	r.sm = sm.NewSimpleSM(r,
		sm.WithFirstState(rackAwaitingStartState, &rackAwaitingStart{}),
		sm.WithState(rackWorkingState, &rackWorking{}),
		sm.WithState(rackDisabledState, &rackDisabled{}),
		sm.WithState(rackFailedState, &rackFailed{}),
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

func (r *rack) start(ctx context.Context) error {
	go r.simulation()

	repl := make(chan *sm.Response)

	r.ch <- &envelope{
		ch: repl,
		span: trace.SpanFromContext(ctx).SpanContext(),
		msg: &startSim{},
	}

	res := <- repl

	if res != nil {
		return res.Err
	}

	return nil
}

func (r *rack) stop(ctx context.Context) {
	repl := make(chan *sm.Response)

	r.ch <- &envelope{
		ch: repl,
		span: trace.SpanFromContext(ctx).SpanContext(),
		msg: &stopSim{},
	}

	<-repl
}

func (r *rack) Receive(ctx context.Context, msg interface{}, ch chan *sm.Response) {
	r.ch <- &envelope{
		ch:  ch,
		span: trace.SpanFromContext(ctx).SpanContext(),
		msg: msg,
	}
}

func (r *rack) simulation() {
	for {
		msg := <-r.ch
		r.sm.Current.Receive(context.Background(), r.sm, msg.msg, msg.ch)
	}
}

type rackAwaitingStart struct {
	sm.NullState
}

func (s *rackAwaitingStart) Receive(ctx context.Context, machine *sm.SimpleSM, msg interface{}, ch chan *sm.Response) {
	ctx, span := tracing.StartSpan(ctx,
		tracing.WithName("Executing rack operation"),
		tracing.WithContextValue(timestamp.EnsureTickInContext))
	defer span.End()

	r := machine.Parent.(*rack)

	switch msg.(type) {
	case *startSim:
		ch <- &sm.Response {
			Err: r.sm.Start(ctx),
			At: common.TickFromContext(ctx),
			Msg: nil,
		}
		_ = machine.ChangeState(ctx, rackWorkingState)

	case *stopSim:
		ch <- nil
	}
}

func (s *rackAwaitingStart) Name() string { return "AwaitingStart" }

type rackWorking struct {
	sm.NullState
}

func (s *rackWorking) Receive(ctx context.Context, machine *sm.SimpleSM, msg interface{}, ch chan *sm.Response) {
	ctx, span := tracing.StartSpan(ctx,
		tracing.WithName("Executing rack operation"),
		tracing.WithContextValue(timestamp.EnsureTickInContext))
	defer span.End()

	r := machine.Parent.(*rack)

	switch body := msg.(type) {
	case *services.InventoryStatusMsg:

		addr := body.Target
		switch addr.Element.(type) {
		case *services.InventoryAddress_Pdu:
			// The PDU is the one component that we can access without
			// using the TOR as a network hop.  This reflects an
			// intentional simplification in the simulation: that the
			// TOR cannot disconnect the PDU, and hte PDU cannot stop
			// the TOR.  In real life, these are probably possible, but
			// for now we simplify our world by avoiding that.
			r.pdu.Receive(ctx, body, ch)

		default:
			// By default, all messages are routed through the TOR
			r.tor.Receive(ctx, body, ch)
		}

	case *services.InventoryRepairMsg:

		switch body.GetAction().(type) {
		case *services.InventoryRepairMsg_Power:
			// All power operations are routed through the PDU
			r.pdu.Receive(ctx, body, ch)

		default:
			// By default, all messages are routed through the TOR
			r.tor.Receive(ctx, body, ch)
		}

	case *stopSim:
		ch <- &sm.Response{
			Err: machine.ChangeState(ctx, rackTerminalState),
			At:  common.TickFromContext(ctx),
			Msg: nil,
		}

	case *startSim:
		ch <- failedResponse(common.TickFromContext(ctx), &sm.UnexpectedMessage{
			Msg:   "startSim",
			State: s.Name(),
		})
	}
}

// Name returns the friendly name for this state.
func (s *rackWorking) Name() string { return "working" }

// rackDisabled is the state where the rack has been defined, but either has
// not yet been authorized for use, or is in the process of being removed from
// the defined inventory.
type rackDisabled struct {
	sm.NullState
}

// Name returns the friendly name for this state.
func (s *rackDisabled) Name() string { return "disabled" }

// rackFailed is the state when the rack has failed.  It is unresponsive to
// normal traffic, but may or may not still have functional elements.
type rackFailed struct {
	sm.NullState
}

// Name returns the friendly name for this state.
func (s *rackFailed) Name() string { return "failed" }

// forwardToBlade is a helper function that forwards a message to the target
// blade in this rack.
func (r *rack) forwardToBlade(ctx context.Context, id int64, msg interface{}, ch chan *sm.Response) {
	if b, ok := r.blades[id]; ok {
		b.Receive(ctx, msg, ch)
	}
}
