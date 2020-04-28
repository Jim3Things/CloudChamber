package sm

import (
    "github.com/AsynkronIT/protoactor-go/actor"

    "github.com/Jim3Things/CloudChamber/pkg/protos/common"
)

type State interface {
    actor.Actor
    Enter(ctx actor.Context) error
    Leave()
}

type EmptyState struct {
}

func (*EmptyState) Receive(_ actor.Context) {}
func (*EmptyState) Enter(_ actor.Context) error { return nil }
func (*EmptyState) Leave() {}

type SM struct {
    Current int
    Behavior actor.Behavior
    States map[int]State
}

func (sm *SM) ChangeState(ctx actor.Context, newState int) error {
    cur := sm.States[sm.Current]
    cur.Leave()

    cur = sm.States[newState]
    if err := cur.Enter(ctx); err != nil {
        return err
    }

    sm.Current = newState
    sm.Behavior.Become(cur.Receive)
    return nil
}

func (sm *SM) Initialize(firstState int) error {
    cur := sm.States[firstState]
    if err := cur.Enter(nil); err != nil {
        return err
    }

    sm.Current = firstState
    sm.Behavior.Become(cur.Receive)
    return nil
}

func (sm *SM) RespondWithError(ctx actor.Context, err error) {
    ctx.Respond(&common.Completion{
        IsError: true,
        Error: err.Error(),
    })
}

