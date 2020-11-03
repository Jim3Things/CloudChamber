package inventory

import (
	"context"

	"github.com/Jim3Things/CloudChamber/internal/common"
	"github.com/Jim3Things/CloudChamber/internal/sm"
	"github.com/Jim3Things/CloudChamber/internal/tracing"
	pbc "github.com/Jim3Things/CloudChamber/pkg/protos/common"
)

type workload struct {
	// temporary workload definition

	// Note that workloads will also be state machines
}

const (
	capacityCores   = "_cores"
	capacityMemory  = "_memoryInMB"
	capacityDisk    = "_diskInGB"
	capacityNetwork = "_networkBandwidthInMbps"

	// acceleratorPrefix is put in front of any accelerator name to ensure that
	// there is no collision with the core capacity categories listed above.
	acceleratorPrefix = "a_"
)

// capacity defines the consumable and capability portions of a blade or
// workload.
type capacity struct {
	// consumables are named units of capacity that are used by a workload such
	// that the amount available to other workloads is reduced by that amount.
	// For example, a core may only be used by one workload at a time.
	consumables map[string]float64

	// features are statements of capabilities that are available for use, but
	// that are not consumed when used.  For example, the presence of security
	// enclave support would be a feature.
	features map[string]bool
}

func newCapacity() *capacity {
	return &capacity{
		consumables: make(map[string]float64),
		features:    make(map[string]bool),
	}
}

type blade struct {
	// rack holds the pointer to the rack that contains this blade.
	holder *rack

	// sm is the state machine for this blade's simulation.
	sm *sm.SimpleSM

	capacity *capacity

	architecture string

	used *capacity

	workloads map[string]*workload
}

const (
	bladeOffState int = iota
	bladeBootingState
	bladeWorkingState
	bladeWorkloadStoppingState
	bladeStoppingState
)

func newBlade(def *pbc.BladeCapacity, r *rack) *blade {
	b := &blade{
		holder:       r,
		sm:           nil,
		capacity:     newCapacity(),
		architecture: def.Arch,
		used:         newCapacity(),
		workloads:    make(map[string]*workload),
	}

	b.capacity.consumables[capacityCores] = float64(def.Cores)
	b.capacity.consumables[capacityMemory] = float64(def.MemoryInMb)
	b.capacity.consumables[capacityDisk] = float64(def.DiskInGb)
	b.capacity.consumables[capacityNetwork] = float64(def.NetworkBandwidthInMbps)

	for _, a := range def.Accelerators {
		b.capacity.consumables[acceleratorPrefix+a.String()] = float64(1)
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
	b.sm.Receive(ctx, msg)
}

// bladeOff is the state when the blade is fully powered off.  It can be
// powered on, but all other operations fail.
type bladeOff struct {
	nullRepairAction
}

func (s *bladeOff) Receive(ctx context.Context, machine *sm.SimpleSM, msg sm.Envelope) {
	s.handleMsg(ctx, machine, s, msg)
}

func (s *bladeOff) power(ctx context.Context, machine *sm.SimpleSM, msg *setPower) {
	occursAt := common.TickFromContext(ctx)

	machine.AdvanceGuard(occursAt)

	// if it is power on, transition to booting
	if msg.on {
		if err := machine.ChangeState(ctx, bladeBootingState); err != nil {
			respondIf(msg.GetCh(), failedResponse(occursAt, err))
		} else {
			respondIf(msg.GetCh(), droppedResponse(occursAt))
		}
	}

	// if it is power off, ignore
}

func (s *bladeOff) connect(ctx context.Context, _ *sm.SimpleSM, msg *setConnection) {
	respondIf(msg.GetCh(), droppedResponse(common.TickFromContext(ctx)))
}

// bladeBooting is the state when the blade is powering on, and the blade is
// waiting for the operation to complete.  It expects either a timeout that
// designates the boot has completed, or a power off command which cancels the
// timer and moves the blade to off.
type bladeBooting struct {
	nullRepairAction
}

func (s *bladeBooting) Enter(ctx context.Context, sm *sm.SimpleSM) error {
	if err := s.nullRepairAction.Enter(ctx, sm); err != nil {
		return err
	}

	// set the boot timer
	return nil
}

func (s *bladeBooting) Receive(ctx context.Context, machine *sm.SimpleSM, msg sm.Envelope) {

}

func (s *bladeBooting) power(ctx context.Context, machine *sm.SimpleSM, msg *setPower) {
	// power has been gated at the pdu, so I think we can ignore gating just here.
	// -- probably need to advance the guard, so that workload operations cannot cross
	// power on is ignored, as we're already powering on.
	// power off cancels the timer and moves to blade off.
}

func (s *bladeBooting) connect(ctx context.Context, _ *sm.SimpleSM, msg *setConnection) {
	msg.GetCh() <- droppedResponse(common.TickFromContext(ctx))
}

// bladeWorking is the stable operational state.  This state expects workload
// operations, physical power off notifications, and planned shutdown requests.
type bladeWorking struct {
	nullRepairAction
}

func (s *bladeWorking) Receive(ctx context.Context, machine *sm.SimpleSM, msg sm.Envelope) {

}

func (s *bladeWorking) power(ctx context.Context, machine *sm.SimpleSM, msg *setPower) {
	// power has been gated at the pdu, so I think we can ignore gating just here.
	// -- probably need to advance the guard, so that workload operations cannot cross
	// power on is ignored, as it is already working.
	// power off immediately terminates all workloads, and moves to blade off
}

// bladeWorkloadStopping is the state when the blade has received a shutdown
// request, has notified the workloads to shut down, and is now waiting for
// them to do so, with a maximum time limit.  This state expects workload
// stopped notifications, delay timeout, or a physical power off notification.
type bladeWorkloadStopping struct {
	nullRepairAction
}

func (s *bladeWorkloadStopping) Enter(ctx context.Context, sm *sm.SimpleSM) error {
	if err := s.nullRepairAction.Enter(ctx, sm); err != nil {
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
	nullRepairAction
}

func (s *bladeStopping) Enter(ctx context.Context, sm *sm.SimpleSM) error {
	if err := s.nullRepairAction.Enter(ctx, sm); err != nil {
		return err
	}

	// set the timer to simulate the delay in a planned stop
	return nil
}

func (s *bladeStopping) Receive(ctx context.Context, machine *sm.SimpleSM, msg sm.Envelope) {

}

func respondIf(ch chan *sm.Response, msg *sm.Response) {
	if ch != nil {
		ch <- msg
	}
}
