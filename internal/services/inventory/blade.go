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
	bladeStart int = iota

	// bladeOffDiscon is current when the blade has neither simulated
	// power or simulated network connectivity.
	bladeOffDiscon

	// bladeOffConn is current when the blade has no simulated power,
	// but does have simulated network connectivity.
	bladeOffConn

	// bladePoweredDiscon is current when the blade has simulated power,
	// but no simulated network connectivity.
	bladePoweredDiscon

	// bladePoweredConn is current when the blade has power and simulated
	// network connectivity.  If auto-boot is enabled, this state will
	// automatically transition to the following booting state.
	bladePoweredConn

	// bladeBooting is current when the blade is waiting for the simulated
	// boot delay to complete.
	bladeBooting

	// bladeWorking is current when the blade is powered on, booted, and
	// able to handle workload requests.
	bladeWorking

	// bladeIsolated is current when the blade is powered on and booted,
	// but has not simulated network connectivity.  Existing workloads are
	// informed the connectivity has been lost, but are otherwise undisturbed.
	bladeIsolated

	// bladeStopping is a transitional state to clean up when the blade is
	// finally shutting down.  This may involve notifying any active workloads
	// that they have been forcibly stopped.
	bladeStopping

	// bladeStoppingIsolated is a transitional state parallel to the
	// bladeStopping, but where simulated network connectivity has been
	// lost.
	bladeStoppingIsolated

	// bladeFaulted is current when the blade has either had a processing
	// fault, such as a timer failure, or an injected fault that leaves it in
	// a position that requires an external reset/fix.
	bladeFaulted
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
			"start",
			startedOnEnter,
			[]sm.ActionEntry{},
			UnexpectedMessage,
			sm.NullLeave),

		sm.WithState(
			bladeOffDiscon,
			"off-disconnected",
			sm.NullEnter,
			[]sm.ActionEntry{
				{messages.TagSetConnection, setConnection, bladeOffConn, sm.Stay},
				{messages.TagSetPower, setPower, bladePoweredDiscon, sm.Stay},
			},
			UnexpectedMessage,
			sm.NullLeave),

		sm.WithState(
			bladeOffConn,
			"off-connected",
			sm.NullEnter,
			[]sm.ActionEntry{
				{messages.TagSetConnection, setConnection, sm.Stay, bladeOffDiscon},
				{messages.TagSetPower, setPower, bladePoweredConn, sm.Stay},
			},
			UnexpectedMessage,
			sm.NullLeave),

		sm.WithState(
			bladePoweredDiscon,
			"powered-disconnected",
			sm.NullEnter,
			[]sm.ActionEntry{
				{messages.TagSetConnection, setConnection, bladePoweredConn, sm.Stay},
				{messages.TagSetPower, setPower, sm.Stay, bladePoweredDiscon},
			},
			UnexpectedMessage,
			sm.NullLeave),

		sm.WithState(
			bladePoweredConn,
			"powered-connected",
			poweredConnOnEnter,
			[]sm.ActionEntry{
				{messages.TagSetConnection, setConnection, sm.Stay, bladePoweredDiscon},
				{messages.TagSetPower, setPower, sm.Stay, bladeOffConn},
			},
			UnexpectedMessage,
			sm.NullLeave),

		sm.WithState(
			bladeBooting,
			"booting",
			bootingOnEnter,
			[]sm.ActionEntry{
				{messages.TagSetConnection, setConnection, sm.Stay, bladePoweredDiscon},
				{messages.TagSetPower, setPower, sm.Stay, bladeOffConn},
				{messages.TagTimerExpiry, bootingTimerExpiry, bladeWorking, sm.Stay},
			},
			UnexpectedMessage,
			bootingOnLeave),

		sm.WithState(
			bladeWorking,
			"working",
			sm.NullEnter,
			[]sm.ActionEntry{
				{messages.TagSetConnection, setConnection, sm.Stay, bladeIsolated},
				{messages.TagSetPower, setPower, sm.Stay, bladeOffConn},
			},
			UnexpectedMessage,
			sm.NullLeave),

		sm.WithState(
			bladeIsolated,
			"isolated",
			sm.NullEnter,
			[]sm.ActionEntry{
				{messages.TagSetConnection, setConnection, bladeWorking, sm.Stay},
				{messages.TagSetPower, setPower, sm.Stay, bladeOffConn},
			},
			UnexpectedMessage,
			sm.NullLeave),

		sm.WithState(
			bladeStopping,
			"stopping",
			sm.NullEnter,
			[]sm.ActionEntry{
				{messages.TagSetConnection, setConnection, sm.Stay, bladeStoppingIsolated},
				{messages.TagSetPower, setPower, sm.Stay, bladeOffConn},
				{messages.TagTimerExpiry, stoppingTimerExpiry, bladeOffConn, sm.Stay},
			},
			UnexpectedMessage,
			sm.NullLeave),

		sm.WithState(
			bladeStoppingIsolated,
			"stopping-isolated",
			sm.NullEnter,
			[]sm.ActionEntry{
				{messages.TagSetConnection, setConnection, bladeStopping, sm.Stay},
				{messages.TagSetPower, setPower, sm.Stay, bladeOffDiscon},
				{messages.TagTimerExpiry, stoppingTimerExpiry, bladeOffDiscon, sm.Stay},
			},
			UnexpectedMessage,
			sm.NullLeave),

		sm.WithState(
			bladeFaulted,
			"faulted",
			faultedEnter,
			[]sm.ActionEntry{},
			DropMessage,
			sm.NullLeave),
	)

	// TEMP TEMP TEMP
	b.sm.Start(context.Background())

	return b
}

func (b *blade) Receive(ctx context.Context, msg sm.Envelope) {
	b.sm.Receive(ctx, msg)
}

func (b *blade) me() *messages.MessageTarget {
	return messages.NewTargetBlade(b.holder.name, b.id)
}

// +++ blade state machine states

func startedOnEnter(ctx context.Context, machine *sm.SimpleSM) error {
	return machine.ChangeState(ctx, bladeOffDiscon)
}

// +++ On & Connected

func poweredConnOnEnter(ctx context.Context, machine *sm.SimpleSM) error {
	b := machine.Parent.(*blade)

	if b.bootOnPower {
		return machine.ChangeState(ctx, bladeBooting)
	}

	return nil
}

// --- On & Connected

// +++ Booting

func bootingOnEnter(ctx context.Context, machine *sm.SimpleSM) error {
	return commonSetTimer(ctx, machine)
}

func bootingTimerExpiry(ctx context.Context, machine *sm.SimpleSM, m sm.Envelope) bool {
	return commonTimeExpiryHandling(ctx, machine, m, "Boot")
}

func bootingOnLeave(ctx context.Context, machine *sm.SimpleSM) {
	b := machine.Parent.(*blade)
	r := b.holder

	if b.hasActiveTimer {
		tracing.Info(ctx, "Canceling boot timer")

		if err := r.cancelTimer(b.matchTimerExpiry); err != nil {
			tracing.Info(
				ctx,
				"failed to cancel boot operation (%v), ignoring remaining activity",
				err)
		}

		b.hasActiveTimer = false
	}
}

// --- Booting

// +++ Stopping (on, connected, booted, stop order received)

func stoppingTimerExpiry(ctx context.Context, machine *sm.SimpleSM, m sm.Envelope) bool {
	return commonTimeExpiryHandling(ctx, machine, m, "Shutdown")
}

// --- Stopping

func faultedEnter(ctx context.Context, machine *sm.SimpleSM) error {
	b := machine.Parent.(*blade)
	r := b.holder

	if b.hasActiveTimer {
		tracing.Info(ctx, "Canceling outstanding timer")

		if err := r.cancelTimer(b.matchTimerExpiry); err != nil {
			tracing.Info(
				ctx,
				"failed to cancel the outstanding timer (%v), ignoring any notification",
				err)
		}

		b.hasActiveTimer = false
	}

	return nil
}

// --- blade state machine states

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

func commonSetTimer(ctx context.Context, machine *sm.SimpleSM) error {
	b := machine.Parent.(*blade)

	if !b.hasActiveTimer {
		r := b.holder

		occursAt := common.TickFromContext(ctx)
		b.activeTimerID++

		// set the new timer
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
	}

	return nil
}

func commonTimeExpiryHandling(
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

func bootDelay() int64 {
	return bladeBootDelayMin + rand.Int63n(bladeBootDelayMax-bladeBootDelayMin)
}
