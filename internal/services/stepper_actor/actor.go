package stepper

import (
    "context"
    "errors"
    "fmt"

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
    actor.Actor
    mgr *sm.SM

    policyToIndex map[pb.StepperPolicy]int

    adapter *server

    // Current tick time
    latest int64

    waiters *treemap.Map
}

func Register(svc *grpc.Server) (err error) {
    s := &Actor{
        adapter: &server{},
        mgr: &sm.SM{ Behavior: actor.NewBehavior() },
        waiters: treemap.NewWith(utils.Int64Comparator),
    }

    // Fill in the states
    s.InitializeStates()

    // Now set up the initial state
    // c, span := getSpan()
    // defer span.End()

    if err := s.mgr.Initialize(context.Background(), nil, InvalidState); err != nil {
        return err
    }

    // With the internals established, attach it as an actor
    ctx := actor.EmptyRootContext
    props := actor.PropsFromFunc(s.Receive).
        WithMailbox(mailbox.Unbounded()).
        WithReceiverMiddleware(log.ReceiveLogger).
        WithSenderMiddleware(log.SendLogger)

    pid, err := ctx.SpawnNamed(props, "Stepper Service")
    if err != nil {
        return err
    }

    // And finally, connect it to the incoming grpc messages
    s.adapter.Attach(pid)
    pb.RegisterStepperServer(svc, s.adapter)

    return nil
}

func (s *Actor) Receive(context actor.Context) {
    s.mgr.States[s.mgr.Current].Receive(context)
}

func (s *Actor) getSpan(ca actor.Context) (context.Context, trc.Span) {
    sn := s.mgr.GetStateName()
    mn := log.MethodName(2)
    span := log.GetSpan(ca.Self())
    ctx := context.Background()

    log.AddEvent(ctx, span, fmt.Sprintf("[In Stepper Actor/%s/%s]", sn, mn), s.latest, "")

    return ctx, span
}
