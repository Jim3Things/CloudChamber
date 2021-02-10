package messages

// This file contains the policy change request messages.

import (
	"context"
	"fmt"
	"time"

	"github.com/Jim3Things/CloudChamber/simulation/internal/sm"
)

const (
	PolicyInvalid int = iota
	PolicyNoWait
	PolicyMeasured
	PolicyManual
)

type BasePolicy interface {
	sm.Envelope

	GetGuard() int64
}

// ManualPolicy contains a request for the simulated time service to use the
// manual step policy.
type ManualPolicy struct {
	sm.EnvelopeState

	Guard int64
}

func NewManualPolicy(
	ctx context.Context,
	guard int64,
	ch chan *sm.Response) *ManualPolicy {

	msg := &ManualPolicy{}
	msg.Initialize(ctx, TagManualPolicy, ch)
	msg.Guard = guard

	return msg
}

func (mp *ManualPolicy) GetGuard() int64 {
	return mp.Guard
}

func (mp *ManualPolicy) String() string {
	replacing := ""
	if mp.Guard >= 0 {
		replacing = fmt.Sprintf(", replacing policy generation %d", mp.Guard)
	}

	return fmt.Sprintf("Advance time manually%s", replacing)
}

// NoWaitPolicy contains a request for the simulated time service to use the
// policy that automatically advances to expire any waiter.
type NoWaitPolicy struct {
	sm.EnvelopeState

	Guard int64
}

func NewNoWaitPolicy(
	ctx context.Context,
	guard int64,
	ch chan *sm.Response) *ManualPolicy {

	msg := &ManualPolicy{}
	msg.Initialize(ctx, TagNoWaitPolicy, ch)
	msg.Guard = guard

	return msg
}

func (nwp *NoWaitPolicy) GetGuard() int64 {
	return nwp.Guard
}

func (nwp *NoWaitPolicy) String() string {
	replacing := ""
	if nwp.Guard >= 0 {
		replacing = fmt.Sprintf(", replacing policy generation %d", nwp.Guard)
	}

	return fmt.Sprintf("Advance time to first waiter%s", replacing)
}

// MeasuredPolicy contains a request for the simulated time service to use the
// policy that automatically advances simulated time at a given wall clock rate.
type MeasuredPolicy struct {
	sm.EnvelopeState

	Guard int64
	Delay time.Duration
}

func NewMeasuredPolicy(
	ctx context.Context,
	guard int64,
	delay time.Duration,
	ch chan *sm.Response) *MeasuredPolicy {

	msg := &MeasuredPolicy{}
	msg.Initialize(ctx, TagMeasuredPolicy, ch)
	msg.Guard = guard
	msg.Delay = delay

	return msg
}

func (mp *MeasuredPolicy) GetGuard() int64 {
	return mp.Guard
}

func (mp *MeasuredPolicy) String() string {
	replacing := ""
	if mp.Guard >= 0 {
		replacing = fmt.Sprintf(", replacing policy generation %d", mp.Guard)
	}

	return fmt.Sprintf("Advance time automatically every %v%s", mp.Delay, replacing)
}
