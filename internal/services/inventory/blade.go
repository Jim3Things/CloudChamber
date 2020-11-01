package inventory

import (
	"context"

	"github.com/Jim3Things/CloudChamber/internal/sm"
	"github.com/Jim3Things/CloudChamber/internal/tracing"
	"github.com/Jim3Things/CloudChamber/pkg/protos/common"
)

type workload struct {
	// temporary workload definition

	// Note that workloads will also be state machines
}

type blade struct {
	// rack holds the pointer to the rack that contains this blade.
	holder *rack

	// sm is the state machine for this blade's simulation.
	sm *sm.SimpleSM

	capacity *common.BladeCapacity

	used *common.BladeCapacity

	workloads map[string]*workload
}

const (
	bladeOffState int = iota
	bladeBootingState
	bladeWorkingState
	bladeWorkloadStoppingState
	bladeStoppingState
)

func newBlade(def *common.BladeCapacity, r *rack) *blade {
	b := &blade{
		holder: r,
		sm:     nil,
		capacity: &common.BladeCapacity{
			Cores:                  def.Cores,
			MemoryInMb:             def.MemoryInMb,
			DiskInGb:				def.DiskInGb,
			NetworkBandwidthInMbps: def.NetworkBandwidthInMbps,
			Arch:                   def.Arch,
			Accelerators:           def.Accelerators,
		},
		used: &common.BladeCapacity{},
		workloads: make(map[string]*workload),
	}

	b.sm = sm.NewSimpleSM(b,
		sm.WithFirstState(bladeOffState, &bladeOff{}),
		sm.WithState(bladeBootingState, &bladeBooting{}),
		sm.WithState(bladeWorkingState, &bladeWorking{}),
		sm.WithState(bladeWorkloadStoppingState, &bladeWorkloadStopping{}),
		sm.WithState(bladeStoppingState, &bladeStopping{}))

	return b
}

func (b *blade) Receive(ctx context.Context, msg sm.Envelope) {

}

// bladeOff is the state when the blade is fully powered off.  It can be
// powered on, but all other operations fail.
type bladeOff struct {
	sm.NullState
}

func (s *bladeOff) Receive(ctx context.Context, machine *sm.SimpleSM, msg sm.Envelope) {

}

// bladeBooting is the state when the blade is powering on, and the blade is
// waiting for the operation to complete.  It expects either a timeout that
// designates the boot has completed, or a power off command which cancels the
// timer and moves the blade to off.
type bladeBooting struct {
	sm.NullState
}

func (s *bladeBooting) Enter(ctx context.Context, sm *sm.SimpleSM) error {
	if err := s.NullState.Enter(ctx, sm); err != nil {
		return err
	}

	// set the boot timer
	return nil
}

func (s *bladeBooting) Receive(ctx context.Context, machine *sm.SimpleSM, msg sm.Envelope) {

}

// bladeWorking is the stable operational state.  This state expects workload
// operations, physical power off notifications, and planned shutdown requests.
type bladeWorking struct {
	sm.NullState
}

func (s *bladeWorking) Receive(ctx context.Context, machine *sm.SimpleSM, msg sm.Envelope) {

}

// bladeWorkloadStopping is the state when the blade has received a shutdown
// request, has notified the workloads to shut down, and is now waiting for
// them to do so, with a maximum time limit.  This state expects workload
// stopped notifications, delay timeout, or a physical power off notification.
type bladeWorkloadStopping struct {
	sm.NullState
}

func (s *bladeWorkloadStopping) Enter(ctx context.Context, sm *sm.SimpleSM) error {
	if err := s.NullState.Enter(ctx, sm); err != nil {
		return err
	}

	b := sm.Parent.(*blade)

	if len(b.workloads) == 0 {
		return sm.ChangeState(ctx, bladeStoppingState)
	}

	// Start all workload timer notifications
	for _, w := range b.workloads {
		// workload shutdown notification goes here
		tracing.Info(ctx, "Should be starting the shutdown timer for workload %v", w)
	}

	return nil
}

func (s *bladeWorkloadStopping) Receive(ctx context.Context, machine *sm.SimpleSM, msg sm.Envelope) {

}

// bladeStopping is the state when the blade has no active workloads and is
// waiting for its simulated OS shutdown to complete.  This state expects a
// delay timeout or a physical power off notification.
type bladeStopping struct {
	sm.NullState
}

func (s *bladeStopping) Enter(ctx context.Context, sm *sm.SimpleSM) error {
	if err := s.NullState.Enter(ctx, sm); err != nil {
		return err
	}

	// set the timer to simulate the delay in a planned stop
	return nil
}

func (s *bladeStopping) Receive(ctx context.Context, machine *sm.SimpleSM, msg sm.Envelope) {

}
