// This module contains the top level function for the stepper actor.

package stepper

import (
    "context"
    "errors"

    "github.com/AsynkronIT/protoactor-go/actor"
    "github.com/AsynkronIT/protoactor-go/mailbox"
    "github.com/emirpasic/gods/maps/treemap"
    "github.com/emirpasic/gods/utils"
    "google.golang.org/grpc"

    "github.com/Jim3Things/CloudChamber/internal/sm"
    "github.com/Jim3Things/CloudChamber/internal/tracing"
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

    epoch  int64                // Epoch counter (policy changes)
}

// Register the stepper actor with the supplied grpc server (via the adapter
// functions)
func Register(svc *grpc.Server, p pb.StepperPolicy) (err error) {
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

    // .. and then, connect it to the incoming grpc messages
    act.adapter.Attach(pid)
    pb.RegisterStepperServer(svc, act.adapter)

    // Finally, if there is a default policy provided, set it now.
    if p != pb.StepperPolicy_Invalid {
        if err := act.adapter.SetDefaultPolicy(p); err != nil {
            return err
        }
    }

    return nil
}

// Actor message receiver.  It handles setting up the per-actor state on
// initialization, or forwards to the current state machine state, once
// set up.
func (act *Actor) Receive(ca actor.Context) {
    ctx := sm.DecorateContext(ca)

    switch ca.Message().(type) {
    case *actor.Started:
        // Fill in the states
        act.InitializeStates()

        if err := act.mgr.Initialize(ctx, InvalidState); err != nil {
            panic(err)
        }

    default:
        act.mgr.States[act.mgr.Current].Receive(ca)
    }
}

// Determine if the current message is a 'system message' or not.
func isSystemMessage(ctx context.Context) bool {
    _, ok := sm.ActorContext(ctx).Message().(actor.SystemMessage)

    return ok
}

// Record entry into a message Receive operation
func (act *Actor) TraceOnReceive(ctx context.Context) {
    sn := act.mgr.GetStateName()
    mn := tracing.MethodName(2)

    log.Infof(ctx, act.latest, "[In Stepper Actor/%s/%s]", sn, mn)
}