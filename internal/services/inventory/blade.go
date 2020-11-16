package inventory

import (
	"context"
	"math/rand"

	"github.com/Jim3Things/CloudChamber/internal/common"
	"github.com/Jim3Things/CloudChamber/internal/services/inventory/messages"
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

const (
	bladeBootDelayMin = int64(3)
	bladeBootDelayMax = int64(5)
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
	// Rack holds the pointer to the Rack that contains this blade.
	holder *Rack

	// id is the index used to identify this blade within the Rack.
	id int64

	// sm is the state machine for this blade's simulation.
	sm *sm.SimpleSM

	capacity *capacity

	architecture string

	used *capacity

	workloads map[string]*workload

	// bootOnPower indicates if the blade should immediately begin booting
	// when power is applied, or if it should wait for a boot console command.
	bootOnPower bool

	// hasActiveTimer indicates if there is an outstanding timer.
	hasActiveTimer bool

	// activeTimerID is the timer supplied ID for an outstanding timer, if
	// there is one.
	activeTimerID int64

	// matchTimerExpiry is the blade supplied ID embedded in the timer expired
	// message.
	matchTimerExpiry int
}

const (
	// bladeOffState is current when the blade has no simulated power.
	bladeOffState int = iota

	// bladePoweredSState is current when the blade has power, but auto boot
	// has not been enabled.
	bladePoweredState

	// bladeBootingState is current when the blade is waiting for the simulated
	// boot delay to complete.
	bladeBootingState

	// bladeWorkingState is current when the blade is powered on, booted, and
	// able to handle workload requests.
	bladeWorkingState

	// bladeWorkloadStoppingState is current when the blade has been told to
	// shut down and it is waiting either for all active workloads to stop, or
	// for the bounding shutdown timer to expire.
	bladeWorkloadStoppingState

	// bladeStoppingState is a transitional state to clean up when the blade is
	// finally shutting down.  This may involve notifying any active workloads
	// that they have been forcibly stopped.
	bladeStoppingState

	// bladeFaultedState is current when the blade has either had a processing
	// fault, such as a timer failure, or an injected fault that leaves it in
	// a position that requires an external reset/fix.
	bladeFaultedState
)

func newBlade(def *pbc.BladeCapacity, r *Rack, id int64) *blade {
	b := &blade{
		holder:           r,
		id:               id,
		sm:               nil,
		capacity:         newCapacity(),
		architecture:     def.Arch,
		used:             newCapacity(),
		workloads:        make(map[string]*workload),
		bootOnPower:      true,
		hasActiveTimer:   false,
		activeTimerID:    0,
		matchTimerExpiry: 0,
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
		sm.WithState(bladePoweredState, &bladePowered{}),
		sm.WithState(bladeBootingState, &bladeBooting{}),
		sm.WithState(bladeWorkingState, &bladeWorking{}),
		sm.WithState(bladeWorkloadStoppingState, &bladeWorkloadStopping{}),
		sm.WithState(bladeStoppingState, &bladeStopping{}),
		sm.WithState(bladeFaultedState, &bladeFaulted{}))

	return b
}

func (b *blade) Receive(ctx context.Context, msg sm.Envelope) {
	b.sm.Receive(ctx, msg)
}

func (b *blade) me() *messages.MessageTarget {
	return messages.NewTargetBlade(b.holder.name, b.id)
}

func (b *blade) powerOnState() int {
	if b.bootOnPower {
		return bladeBootingState
	}

	return bladePoweredState
}

// +++ blade state machine states

// bladeOff is the state when the blade is fully powered off.  It can be
// powered on, but all other operations fail.
type bladeOff struct {
	dropRepairAction
}

func (s *bladeOff) Receive(ctx context.Context, machine *sm.SimpleSM, msg sm.Envelope) {
	s.handleMsg(ctx, machine, s, msg)
}

func (s *bladeOff) Name() string { return "off" }

func (s *bladeOff) Power(ctx context.Context, machine *sm.SimpleSM, msg *messages.SetPower) {
	tracing.UpdateSpanName(
		ctx,
		"Processing power %s command at %s",
		common.AOrB(msg.On, "on", "off"),
		msg.Target.Describe())

	b := machine.Parent.(*blade)

	occursAt := common.TickFromContext(ctx)

	machine.AdvanceGuard(occursAt)

	if msg.On {
		faultingChange(ctx, machine, b.powerOnState(), msg)
	}

	// if it is power off, ignore
}

// bladePowered is the state where the blade has had power applied, and is
// waiting for a boot command.
type bladePowered struct {
	dropRepairAction
}

func (s *bladePowered) Receive(ctx context.Context, machine *sm.SimpleSM, msg sm.Envelope) {
	s.handleMsg(ctx, machine, s, msg)
}

func (s *bladePowered) Name() string { return "powered" }

func (s *bladePowered) Power(ctx context.Context, machine *sm.SimpleSM, msg *messages.SetPower) {
	tracing.UpdateSpanName(
		ctx,
		"Processing power %s command at %s",
		common.AOrB(msg.On, "on", "off"),
		msg.Target.Describe())

	occursAt := common.TickFromContext(ctx)

	machine.AdvanceGuard(occursAt)

	if !msg.On {
		faultingChange(ctx, machine, bladeOffState, msg)
	}

	// if it is power on, ignore
}

// bladeBooting is the state when the blade is powering on, and the blade is
// waiting for the operation to complete.  It expects either a timeout that
// designates the boot has completed, or a power off command which cancels the
// timer and moves the blade to off.
type bladeBooting struct {
	dropRepairAction
}

func (s *bladeBooting) Enter(ctx context.Context, machine *sm.SimpleSM) error {
	if err := s.dropRepairAction.Enter(ctx, machine); err != nil {
		return err
	}

	b := machine.Parent.(*blade)
	r := b.holder

	occursAt := common.TickFromContext(ctx)
	b.activeTimerID++

	// set the boot timer
	expiryMsg := messages.NewTimerExpiry(
		ctx,
		b.me(),
		occursAt,
		b.activeTimerID,
		nil,
		nil)

	timerId, err := r.setTimer(ctx, bootDelay(), expiryMsg)
	if err != nil {
		return err
	}

	b.hasActiveTimer = true
	b.matchTimerExpiry = timerId
	return nil
}

func (s *bladeBooting) Receive(ctx context.Context, machine *sm.SimpleSM, msg sm.Envelope) {
	s.handleMsg(ctx, machine, s, msg)
}

func (s *bladeBooting) Name() string { return "booting" }

func (s *bladeBooting) Power(ctx context.Context, machine *sm.SimpleSM, msg *messages.SetPower) {
	tracing.UpdateSpanName(
		ctx,
		"Processing power %s command at %s",
		common.AOrB(msg.On, "on", "off"),
		msg.Target.Describe())

	occursAt := common.TickFromContext(ctx)

	machine.AdvanceGuard(occursAt)

	if !msg.On {
		b := machine.Parent.(*blade)
		r := b.holder

		if err := r.cancelTimer(b.matchTimerExpiry); err != nil {
			tracing.Info(
				ctx,
				"failed to cancel boot operation (%v), ignoring remaining activity",
				err)
		}

		b.hasActiveTimer = false
		faultingChange(ctx, machine, bladeOffState, msg)
	}

	// if it is power on, ignore
}

func (s *bladeBooting) Timeout(ctx context.Context, machine *sm.SimpleSM, msg *messages.TimerExpiry) {
	b := machine.Parent.(*blade)

	if b.hasActiveTimer && b.activeTimerID == msg.Id {
		tracing.UpdateSpanName(
			ctx,
			"Boot completed for %s",
			msg.Target.Describe())

		b.hasActiveTimer = false

		machine.AdvanceGuard(common.TickFromContext(ctx))
		faultingChange(ctx, machine, bladeWorkingState, msg)
	}
}

// bladeWorking is the stable operational state.  This state expects workload
// operations, physical power off notifications, and planned shutdown requests.
type bladeWorking struct {
	nullRepairAction
}

func (s *bladeWorking) Receive(ctx context.Context, machine *sm.SimpleSM, msg sm.Envelope) {

}

func (s *bladeWorking) Name() string { return "working" }

func (s *bladeWorking) power(ctx context.Context, machine *sm.SimpleSM, msg *messages.SetPower) {
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

func (s *bladeWorkloadStopping) Name() string { return "workloadStopping" }

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

func (s *bladeStopping) Name() string { return "stopping" }

type bladeFaulted struct {
	dropRepairAction
}

// --- blade state machine states

func respondIf(ch chan *sm.Response, msg *sm.Response) {
	if ch != nil {
		ch <- msg
	}
}

func faultingChange(ctx context.Context, machine *sm.SimpleSM, stateIndex int, msg sm.Envelope) {
	occursAt := common.TickFromContext(ctx)

	if err := machine.ChangeState(ctx, stateIndex); err != nil {
		respondIf(msg.GetCh(), messages.FailedResponse(occursAt, err))
		_ = machine.ChangeState(ctx, bladeFaultedState)
	} else {
		respondIf(msg.GetCh(), messages.SuccessResponse(occursAt))
	}
}

func bootDelay() int64 {
	return bladeBootDelayMin + rand.Int63n(bladeBootDelayMax-bladeBootDelayMin)
}
