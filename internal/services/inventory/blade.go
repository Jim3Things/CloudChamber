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
	// bladeStart is the state where initialization of the state machine
	// begins.
	bladeStart = "start"

	// bladeOffDiscon is current when the blade has neither simulated
	// power or simulated network connectivity.
	bladeOffDiscon = "off-disconnected"

	// bladeOffConn is current when the blade has no simulated power,
	// but does have simulated network connectivity.
	bladeOffConn = "off-connected"

	// bladePoweredDiscon is current when the blade has simulated power,
	// but no simulated network connectivity.
	bladePoweredDiscon = "powered-disconnected"

	// bladePoweredConn is current when the blade has power and simulated
	// network connectivity.  If auto-boot is enabled, this state will
	// automatically transition to the following booting state.
	bladePoweredConn = "powered-connected"

	// bladeBooting is current when the blade is waiting for the simulated
	// boot delay to complete.
	bladeBooting = "booting"

	// bladeWorking is current when the blade is powered on, booted, and
	// able to handle workload requests.
	bladeWorking = "working"

	// bladeIsolated is current when the blade is powered on and booted,
	// but has not simulated network connectivity.  Existing workloads are
	// informed the connectivity has been lost, but are otherwise undisturbed.
	bladeIsolated = "isolated"

	// bladeStopping is a transitional state to clean up when the blade is
	// finally shutting down.  This may involve notifying any active workloads
	// that they have been forcibly stopped.
	bladeStopping = "stopping"

	// bladeStoppingIsolated is a transitional state parallel to the
	// bladeStopping, but where simulated network connectivity has been
	// lost.
	bladeStoppingIsolated = "stopping-isolated"

	// bladeFaulted is current when the blade has either had a processing
	// fault, such as a timer failure, or an injected fault that leaves it in
	// a position that requires an external reset/fix.
	bladeFaulted = "faulted"
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
		sm.WithFirstState(
			bladeStart,
			startedOnEnter,
			[]sm.ActionEntry{},
			sm.UnexpectedMessage,
			sm.NullLeave),

		sm.WithState(
			bladeOffDiscon,
			sm.NullEnter,
			[]sm.ActionEntry{
				{messages.TagSetConnection, setConnection, bladeOffConn, sm.Stay},
				{messages.TagSetPower, setPower, bladePoweredDiscon, sm.Stay},
			},
			sm.UnexpectedMessage,
			sm.NullLeave),

		sm.WithState(
			bladeOffConn,
			sm.NullEnter,
			[]sm.ActionEntry{
				{messages.TagSetConnection, setConnection, sm.Stay, bladeOffDiscon},
				{messages.TagSetPower, setPower, bladePoweredConn, sm.Stay},
			},
			sm.UnexpectedMessage,
			sm.NullLeave),

		sm.WithState(
			bladePoweredDiscon,
			sm.NullEnter,
			[]sm.ActionEntry{
				{messages.TagSetConnection, setConnection, bladePoweredConn, sm.Stay},
				{messages.TagSetPower, setPower, sm.Stay, bladePoweredDiscon},
			},
			sm.UnexpectedMessage,
			sm.NullLeave),

		sm.WithState(
			bladePoweredConn,
			poweredConnOnEnter,
			[]sm.ActionEntry{
				{messages.TagSetConnection, setConnection, sm.Stay, bladePoweredDiscon},
				{messages.TagSetPower, setPower, sm.Stay, bladeOffConn},
			},
			sm.UnexpectedMessage,
			sm.NullLeave),

		sm.WithState(
			bladeBooting,
			bootingOnEnter,
			[]sm.ActionEntry{
				{messages.TagSetConnection, setConnection, sm.Stay, bladePoweredDiscon},
				{messages.TagSetPower, setPower, sm.Stay, bladeOffConn},
				{messages.TagTimerExpiry, bootingTimerExpiry, bladeWorking, sm.Stay},
			},
			sm.UnexpectedMessage,
			bootingOnLeave),

		sm.WithState(
			bladeWorking,
			sm.NullEnter,
			[]sm.ActionEntry{
				{messages.TagSetConnection, setConnection, sm.Stay, bladeIsolated},
				{messages.TagSetPower, setPower, sm.Stay, bladeOffConn},
			},
			sm.UnexpectedMessage,
			sm.NullLeave),

		sm.WithState(
			bladeIsolated,
			sm.NullEnter,
			[]sm.ActionEntry{
				{messages.TagSetConnection, setConnection, bladeWorking, sm.Stay},
				{messages.TagSetPower, setPower, sm.Stay, bladeOffConn},
			},
			sm.UnexpectedMessage,
			sm.NullLeave),

		sm.WithState(
			bladeStopping,
			sm.NullEnter,
			[]sm.ActionEntry{
				{messages.TagSetConnection, setConnection, sm.Stay, bladeStoppingIsolated},
				{messages.TagSetPower, setPower, sm.Stay, bladeOffConn},
				{messages.TagTimerExpiry, stoppingTimerExpiry, bladeOffConn, sm.Stay},
			},
			sm.UnexpectedMessage,
			stoppingOnLeave),

		sm.WithState(
			bladeStoppingIsolated,
			sm.NullEnter,
			[]sm.ActionEntry{
				{messages.TagSetConnection, setConnection, bladeStopping, sm.Stay},
				{messages.TagSetPower, setPower, sm.Stay, bladeOffDiscon},
				{messages.TagTimerExpiry, stoppingTimerExpiry, bladeOffDiscon, sm.Stay},
			},
			sm.UnexpectedMessage,
			stoppingOnLeave),

		sm.WithState(
			bladeFaulted,
			faultedEnter,
			[]sm.ActionEntry{},
			messages.DropMessage,
			sm.NullLeave),
	)

	return b
}

func (b *blade) Receive(ctx context.Context, msg sm.Envelope) {
	b.sm.Receive(ctx, msg)
}

func (b *blade) me() *messages.MessageTarget {
	return messages.NewTargetBlade(b.holder.name, b.id)
}

// +++ blade state machine actions

// startedOnEnter initializes the simulation state and transitions to the
// off and disconnected state.
func startedOnEnter(ctx context.Context, machine *sm.SimpleSM) error {
	return machine.ChangeState(ctx, bladeOffDiscon)
}

// poweredConnOnEnter checks if automatic booting is enabled.  If it is, the
// blade transitions immediately into booting.
func poweredConnOnEnter(ctx context.Context, machine *sm.SimpleSM) error {
	b := machine.Parent.(*blade)

	if b.bootOnPower {
		return machine.ChangeState(ctx, bladeBooting)
	}

	return nil
}

// bootingOnEnter starts the delay timer used to simulate the time needed to
// boot.
func bootingOnEnter(ctx context.Context, machine *sm.SimpleSM) error {
	return setTimer(ctx, machine, bootDelay())
}

// bootingTimerExpiry processes the boot delay timer expiration message.  This
// will lead to a transition to the next state, if successful.
func bootingTimerExpiry(ctx context.Context, machine *sm.SimpleSM, m sm.Envelope) bool {
	return timerExpiration(ctx, machine, m, "Boot")
}

// bootingOnLeave ensures that any active boot delay timer is canceled before
// proceeding to a non-booting state.
func bootingOnLeave(ctx context.Context, machine *sm.SimpleSM, nextState string) {
	if nextState != bladeBooting {
		cancelTimer(ctx, machine, "boot")
	}
}

// stoppingTimerExpiry processes the planned shutdown delay timer expiration.
// This will allow a transition to the appropriate stopped state, if
// successful.
func stoppingTimerExpiry(ctx context.Context, machine *sm.SimpleSM, m sm.Envelope) bool {
	return timerExpiration(ctx, machine, m, "Shutdown")
}

// stoppingOnLeave ensures that any active time is canceled before proceeding
// to a non-stopping state.
func stoppingOnLeave(ctx context.Context, machine *sm.SimpleSM, nextState string) {
	if nextState != bladeStopping && nextState != bladeStoppingIsolated {
		cancelTimer(ctx, machine, "shutdown")
	}
}

// faultedEnter cancels any outstanding timers as a belt-and-braces practice,
// given that faulted is the state that the blade transitions to on any of
// several error paths.
func faultedEnter(ctx context.Context, machine *sm.SimpleSM) error {
	cancelTimer(ctx, machine, "outstanding")
	return nil
}

// setConnection processes the incoming set connection message, returning true
// if the connection is to be on, false if it is to be off.  Absent any earlier
// filtering this can result in null transitions (state-a to state-a), which
// the state machine needs to be prepared for.
func setConnection(ctx context.Context, machine *sm.SimpleSM, m sm.Envelope) bool {
	msg := m.(*messages.SetConnection)

	tracing.UpdateSpanName(
		ctx,
		"Processing network connection %s notification at %s",
		common.AOrB(msg.Enabled, "enabled", "disabled"),
		msg.Target.Describe())

	machine.AdvanceGuard(common.TickFromContext(ctx))

	return msg.Enabled
}

// setPower processes the incoming set power on/off message, returning true
// if the power is to be on, false if it is to be off.  Absent any earlier
// filtering this can result in null transitions (state-a to state-a), which
// the state machine needs to be prepared for.
func setPower(ctx context.Context, machine *sm.SimpleSM, m sm.Envelope) bool {
	msg := m.(*messages.SetPower)

	tracing.UpdateSpanName(
		ctx,
		"Processing power %s command at %s",
		common.AOrB(msg.On, "on", "off"),
		msg.Target.Describe())

	machine.AdvanceGuard(common.TickFromContext(ctx))

	return msg.On
}

// --- blade state machine actions

// +++ support functions

// setTimer establishes a new timer if one is not currently active.
func setTimer(ctx context.Context, machine *sm.SimpleSM, delay int64) error {
	b := machine.Parent.(*blade)

	if !b.hasActiveTimer {
		r := b.holder

		occursAt := common.TickFromContext(ctx)
		b.activeTimerID++

		// set the new timer
		expiryMsg := messages.NewTimerExpiry(
			b.me(),
			occursAt,
			b.activeTimerID,
			nil,
			nil)

		timerId, err := r.setTimer(ctx, delay, expiryMsg)
		if err != nil {
			return err
		}

		b.hasActiveTimer = true
		b.matchTimerExpiry = timerId
	}

	return nil
}

// timerExpiration processes a timer expiration, performing validation that it
// is still expected, and cleanup of the blade's context.
func timerExpiration(
	ctx context.Context,
	machine *sm.SimpleSM,
	m sm.Envelope,
	opCompleted string) bool {
	msg := m.(*messages.TimerExpiry)

	b := machine.Parent.(*blade)

	if b.hasActiveTimer && b.activeTimerID == msg.Id {
		tracing.UpdateSpanName(
			ctx,
			"%s completed for %s",
			opCompleted,
			msg.Target.Describe())

		b.hasActiveTimer = false

		machine.AdvanceGuard(common.TickFromContext(ctx))
		return true
	}

	return false
}

// cancelTimer attempts to cancel an outstanding timer, if one exists.  It
// clears the internal flag that denotes a timer is expected.  This ensures
// that the timer is treated as canceled, regardless of whether or not it was
// successfully canceled.
func cancelTimer(ctx context.Context, machine *sm.SimpleSM, name string) {
	b := machine.Parent.(*blade)
	r := b.holder

	if b.hasActiveTimer {
		tracing.Info(ctx, "Canceling %s timer", name)

		if err := r.cancelTimer(b.matchTimerExpiry); err != nil {
			tracing.Info(
				ctx,
				"failed to cancel %s timer (%v), ignoring remaining activity",
				name,
				err)
		}

		b.hasActiveTimer = false
	}
}

// bootDelay calculates a simulated length of time that booting should take,
// within the acceptable limits.
func bootDelay() int64 {
	return bladeBootDelayMin + rand.Int63n(bladeBootDelayMax-bladeBootDelayMin)
}
