// This module contains the top level function for the stepper actor.

package stepper

import (
    "context"
    "errors"

    "github.com/AsynkronIT/protoactor-go/actor"
    "github.com/AsynkronIT/protoactor-go/mailbox"
    "github.com/emirpasic/gods/maps/treemap"
    "github.com/emirpasic/gods/utils"
    trc "go.opentelemetry.io/otel/api/trace"
    "google.golang.org/grpc"

    "github.com/Jim3Things/CloudChamber/internal/sm"
    log "github.com/Jim3Things/CloudChamber/internal/tracing/server"
    pb "github.com/Jim3Things/CloudChamber/pkg/protos/Stepper"
)

var (
    ErrInvalidRequest = errors.New("invalid request message")
)

// Define the stepper actor
type Actor struct {
    actor.Actor                 // Common actor definitions
    mgr *sm.SM                  // .. with a state machine
    adapter *server             // .. and a grpc adapter

    latest int64                // current simulate time, in ticks
    waiters *treemap.Map        // waiting delay operations
}

// Register the stepper actor with the supplied grpc server (via the adapter
// functions)
func Register(svc *grpc.Server) (err error) {
    // Create the base actor object
    act := &Actor{
        adapter: &server{},
        mgr: &sm.SM{ Behavior: actor.NewBehavior() },
        waiters: treemap.NewWith(utils.Int64Comparator),
    }

    // .. then with the internals established, attach it as an actor
    ctx := actor.EmptyRootContext
    props := actor.PropsFromFunc(act.Receive).
        WithMailbox(mailbox.Unbounded()).
        WithReceiverMiddleware(log.ReceiveLogger).
        WithSenderMiddleware(log.SendLogger)

    pid, err := ctx.SpawnNamed(props, "Stepper Service")
    if err != nil {
        return err
    }

    // .. and finally, connect it to the incoming grpc messages
    act.adapter.Attach(pid)
    pb.RegisterStepperServer(svc, act.adapter)

    return nil
}

// Actor message receiver.  It handles setting up the per-actor state on
// initialization, or forwards to the current state machine state, once
// set up.
func (act *Actor) Receive(ctx actor.Context) {
    switch ctx.Message().(type) {
    case *actor.Started:
        // Fill in the states
        act.InitializeStates()

        if err := act.mgr.Initialize(context.Background(), nil, InvalidState); err != nil {
            panic(err)
        }

    default:
        act.mgr.States[act.mgr.Current].Receive(ctx)
    }
}

// Get the active span for this actor.  It is used by the state machine
// implementations to get the span context for their log entries.
func (act *Actor) getSpan(ca actor.Context) (context.Context, trc.Span) {
    sn := act.mgr.GetStateName()
    mn := log.MethodName(2)
    span := log.GetSpan(ca.Self())
    ctx := context.Background()

    log.Infof(ctx, span, act.latest, "[In Stepper Actor/%s/%s]", sn, mn)

    return ctx, span
}

// Determine if the current message is a 'system message' or not.
func isSystemMessage(_ context.Context, _ trc.Span, ctx actor.Context) bool {
    _, ok := ctx.Message().(actor.SystemMessage)

    return ok
}