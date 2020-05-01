package sm

import (
    "context"
    "fmt"

    "github.com/AsynkronIT/protoactor-go/actor"
    trc "go.opentelemetry.io/otel/api/trace"

    trace "github.com/Jim3Things/CloudChamber/internal/tracing/server"
    "github.com/Jim3Things/CloudChamber/pkg/protos/common"
)

type State interface {
    actor.Actor
    Enter(ctx actor.Context) error
    Leave()
}

type EmptyState struct {
}

func (*EmptyState) Enter(_ actor.Context) error { return nil }
func (*EmptyState) Receive(_ actor.Context) {}
func (*EmptyState) Leave()                  {}

type SM struct {
    Current int
    Behavior actor.Behavior
    States map[int]State
    StateNames map[int]string
}

func (sm *SM) ChangeState(c context.Context, span trc.Span, ctx actor.Context, latest int64, newState int) error {
    sm.AddEvent(c, span, latest, "Change state to %q", sm.StateNames[newState])
    cur := sm.States[sm.Current]
    cur.Leave()

    cur = sm.States[newState]
    if err := cur.Enter(ctx); err != nil {
        return trace.LogError(c, latest, err)
    }

    sm.Current = newState
    sm.Behavior.Become(cur.Receive)
    return nil
}

func (sm *SM) Initialize(c context.Context, span trc.Span, firstState int) error {
    cur := sm.States[firstState]
    if err := cur.Enter(nil); err != nil {
        return trace.LogError(c, 0, err)
    }

    sm.Current = firstState
    sm.Behavior.Become(cur.Receive)
    return nil
}

func (sm *SM) RespondWithError(c context.Context, span trc.Span, ctx actor.Context, err error) {
    ctx.Respond(&common.Completion{
        IsError: true,
        Error: err.Error(),
    })
}

func (sm *SM) GetStateName() string {
    n, ok := sm.StateNames[sm.Current]
    if !ok { n = "<unknown>"}
    return n
}

func (sm *SM) AddEvent(c context.Context, _ trc.Span, latest int64, format string, a ...interface{}) {
    msg := fmt.Sprintf(format, a...)
    trace.AddEvent(c, fmt.Sprintf("[In state %q]: %s", sm.StateNames[sm.Current], msg), latest, "")
}

