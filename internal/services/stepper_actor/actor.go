package stepper

import (
    "errors"
    "time"

    "github.com/AsynkronIT/protoactor-go/actor"
    "github.com/AsynkronIT/protoactor-go/mailbox"
    "github.com/emirpasic/gods/maps/treemap"
    "google.golang.org/grpc"

    "github.com/Jim3Things/CloudChamber/internal/sm"
    pb "github.com/Jim3Things/CloudChamber/pkg/protos/Stepper"
)

const (
    ActorTimeout = time.Duration(1000)
)

var (
    ErrInvalidRequest = errors.New("invalid request message")
)

type WaitQueue struct {
    keys []int64
    values map[int64]interface{}
}

// Define the stepper actor
type Actor struct {
    actor.Actor
    mgr *sm.SM

    adapter *server
    pid *actor.PID

    // Current tick time
    latest int64

    waiters *treemap.Map

    states map[int]sm.State
}

func Register(svc *grpc.Server) (err error) {
    s := &Actor{
        adapter: &server{},
        mgr: &sm.SM{
            Behavior: actor.NewBehavior(),
            States: make(map[int]sm.State),
        },
        waiters: treemap.NewWithIntComparator(),
    }

    // Fill in the states
    s.InitializeStates()

    // Now set up the initial state
    if err := s.mgr.Initialize(InvalidState); err != nil {
        return err
    }

    // With the internals established, attach it as an actor
    ctx := actor.EmptyRootContext
    props := actor.PropsFromFunc(s.Receive).
        WithMailbox(mailbox.Unbounded()).
        WithReceiverMiddleware(). // todo: logging interceptor
        WithSenderMiddleware() // todo: logging interceptor

    if s.pid, err = ctx.SpawnNamed(props, "Stepper Service"); err != nil {
        return err
    }

    // And finally, connect it to the incoming grpc messages
    pb.RegisterStepperServer(svc, s.adapter)

    return nil
}

func (s *Actor) Receive(context actor.Context) {
    s.mgr.States[s.mgr.Current].Receive(context)
}

