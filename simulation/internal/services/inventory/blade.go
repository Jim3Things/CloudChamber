package inventory

import (
	"context"
	"math/rand"

	"github.com/golang/protobuf/proto"

	"github.com/Jim3Things/CloudChamber/simulation/internal/common"
	"github.com/Jim3Things/CloudChamber/simulation/internal/services/inventory/messages"
	"github.com/Jim3Things/CloudChamber/simulation/internal/sm"
	"github.com/Jim3Things/CloudChamber/simulation/internal/tracing"

	pb "github.com/Jim3Things/CloudChamber/simulation/pkg/protos/inventory"
)

type workload struct {
	// temporary workload definition

	// Note that workloads will also be state machines
}

const (
	bladeBootDelayMin = int64(3)
	bladeBootDelayMax = int64(5)
)

type blade struct {
	// Rack holds the pointer to the Rack that contains this blade.
	holder *Rack

	// id is the index used to identify this blade within the Rack.
	id int64

	// sm is the state machine for this blade's simulation.
	sm *sm.SM

	capacity *messages.Capacity

	architecture string

	used *messages.Capacity

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

	expiration int64
}

func newBlade(ctx context.Context, def *pb.Definition_Blade, name string, r *Rack, id int64) *blade {
	capacity := def.GetCapacity()

	b := &blade{
		holder:           r,
		id:               id,
		sm:               nil,
		capacity:         messages.NewCapacity(),
		architecture:     capacity.GetArch(),
		used:             messages.NewCapacity(),
		workloads:        make(map[string]*workload),
		bootOnPower:      def.GetBootOnPowerOn(),
		hasActiveTimer:   false,
		activeTimerID:    0,
		matchTimerExpiry: 0,
	}

	b.capacity.Consumables[messages.CapacityCores] = float64(capacity.GetCores())
	b.capacity.Consumables[messages.CapacityMemory] = float64(capacity.GetMemoryInMb())
	b.capacity.Consumables[messages.CapacityDisk] = float64(capacity.GetDiskInGb())
	b.capacity.Consumables[messages.CapacityNetwork] = float64(capacity.GetNetworkBandwidthInMbps())

	for _, a := range capacity.GetAccelerators() {
		b.capacity.Consumables[messages.AcceleratorPrefix+a.String()] = float64(1)
	}

	b.sm = sm.NewSM(b,
		name,
		sm.WithFirstState(
			pb.BladeState_start,
			startedOnEnter,
			[]sm.ActionEntry{},
			sm.UnexpectedMessage,
			sm.NullLeave),

		sm.WithState(
			pb.BladeState_off_disconnected,
			sm.NullEnter,
			[]sm.ActionEntry{
				{messages.TagSetConnection, setConnection, pb.BladeState_off_connected, sm.Stay},
				{messages.TagSetPower, setPower, pb.BladeState_powered_disconnected, sm.Stay},
			},
			sm.UnexpectedMessage,
			sm.NullLeave),

		sm.WithState(
			pb.BladeState_off_connected,
			sm.NullEnter,
			[]sm.ActionEntry{
				{messages.TagGetStatus, sm.Ignore, sm.Stay, sm.Stay},
				{messages.TagSetConnection, setConnection, sm.Stay, pb.BladeState_off_disconnected},
				{messages.TagSetPower, setPower, pb.BladeState_powered_connected, sm.Stay},
			},
			sm.UnexpectedMessage,
			sm.NullLeave),

		sm.WithState(
			pb.BladeState_powered_disconnected,
			sm.NullEnter,
			[]sm.ActionEntry{
				{messages.TagSetConnection, setConnection, pb.BladeState_powered_connected, sm.Stay},
				{messages.TagSetPower, setPower, sm.Stay, pb.BladeState_powered_disconnected},
			},
			sm.UnexpectedMessage,
			sm.NullLeave),

		sm.WithState(
			pb.BladeState_powered_connected,
			poweredConnOnEnter,
			[]sm.ActionEntry{
				{messages.TagGetStatus, bladeGetStatus, sm.Stay, sm.Stay},
				{messages.TagSetConnection, setConnection, sm.Stay, pb.BladeState_powered_disconnected},
				{messages.TagSetPower, setPower, sm.Stay, pb.BladeState_off_connected},
			},
			sm.UnexpectedMessage,
			sm.NullLeave),

		sm.WithState(
			pb.BladeState_booting,
			bootingOnEnter,
			[]sm.ActionEntry{
				{messages.TagGetStatus, bladeGetStatus, sm.Stay, sm.Stay},
				{messages.TagSetConnection, setConnection, sm.Stay, pb.BladeState_powered_disconnected},
				{messages.TagSetPower, setPower, sm.Stay, pb.BladeState_off_connected},
				{messages.TagTimerExpiry, bootingTimerExpiry, pb.BladeState_working, sm.Stay},
			},
			sm.UnexpectedMessage,
			bootingOnLeave),

		sm.WithState(
			pb.BladeState_working,
			sm.NullEnter,
			[]sm.ActionEntry{
				{messages.TagGetStatus, bladeGetStatus, sm.Stay, sm.Stay},
				{messages.TagSetConnection, setConnection, sm.Stay, pb.BladeState_isolated},
				{messages.TagSetPower, setPower, sm.Stay, pb.BladeState_off_connected},
			},
			sm.UnexpectedMessage,
			sm.NullLeave),

		sm.WithState(
			pb.BladeState_isolated,
			sm.NullEnter,
			[]sm.ActionEntry{
				{messages.TagSetConnection, setConnection, pb.BladeState_working, sm.Stay},
				{messages.TagSetPower, setPower, sm.Stay, pb.BladeState_off_connected},
			},
			sm.UnexpectedMessage,
			sm.NullLeave),

		sm.WithState(
			pb.BladeState_stopping,
			sm.NullEnter,
			[]sm.ActionEntry{
				{messages.TagGetStatus, bladeGetStatus, sm.Stay, sm.Stay},
				{messages.TagSetConnection, setConnection, sm.Stay, pb.BladeState_stopping_isolated},
				{messages.TagSetPower, setPower, sm.Stay, pb.BladeState_off_connected},
				{messages.TagTimerExpiry, stoppingTimerExpiry, pb.BladeState_off_connected, sm.Stay},
			},
			sm.UnexpectedMessage,
			stoppingOnLeave),

		sm.WithState(
			pb.BladeState_stopping_isolated,
			sm.NullEnter,
			[]sm.ActionEntry{
				{messages.TagSetConnection, setConnection, pb.BladeState_stopping, sm.Stay},
				{messages.TagSetPower, setPower, sm.Stay, pb.BladeState_off_disconnected},
				{messages.TagTimerExpiry, stoppingTimerExpiry, pb.BladeState_off_disconnected, sm.Stay},
			},
			sm.UnexpectedMessage,
			stoppingOnLeave),

		sm.WithState(
			pb.BladeState_faulted,
			faultedEnter,
			[]sm.ActionEntry{},
			messages.DropMessage,
			sm.NullLeave),
	)

	tracing.AddImpact(ctx, tracing.ImpactCreate, name)

	return b
}

// Save returns a protobuf instance containing the information needed to later
// restore this state machine instance to the logical equivalence of its current
// state.
func (b *blade) Save() (proto.Message, error) {
	cur, entered, terminal, guard := b.sm.Savable()

	return &pb.Actual_Blade{
		Condition:    pb.Actual_operational,
		SmState:      cur.(pb.BladeState_SM),
		StateExpires: b.hasActiveTimer,
		Expiration:   b.expiration,
		Core: &pb.Actual_MachineCore{
			EnteredAt: entered,
			Terminal:  terminal,
			Guard:     guard,
		},
	}, nil
}

func (b *blade) Receive(ctx context.Context, msg sm.Envelope) {
	b.sm.Receive(ctx, msg)
}

func (b *blade) me() *messages.MessageTarget {
	return messages.NewTargetBlade(b.holder.sm.Name, b.id, 0)
}

// +++ blade state machine actions

// startedOnEnter initializes the simulation state and transitions to the
// off and disconnected state.
func startedOnEnter(ctx context.Context, machine *sm.SM) error {
	return machine.ChangeState(ctx, pb.BladeState_off_disconnected)
}

// poweredConnOnEnter checks if automatic booting is enabled.  If it is, the
// blade transitions immediately into booting.
func poweredConnOnEnter(ctx context.Context, machine *sm.SM) error {
	b := machine.Parent.(*blade)

	if b.bootOnPower {
		return machine.ChangeState(ctx, pb.BladeState_booting)
	}

	return nil
}

// bootingOnEnter starts the delay timer used to simulate the time needed to
// boot.
func bootingOnEnter(ctx context.Context, machine *sm.SM) error {
	return setTimer(ctx, machine, bootDelay())
}

// bootingTimerExpiry processes the boot delay timer expiration message.  This
// will lead to a transition to the next state, if successful.
func bootingTimerExpiry(ctx context.Context, machine *sm.SM, m sm.Envelope) bool {
	return timerExpiration(ctx, machine, m, "Boot")
}

// bootingOnLeave ensures that any active boot delay timer is canceled before
// proceeding to a non-booting state.
func bootingOnLeave(ctx context.Context, machine *sm.SM, nextState sm.StateIndex) {
	if nextState != pb.BladeState_booting {
		cancelTimer(ctx, machine, "boot")
	}
}

// stoppingTimerExpiry processes the planned shutdown delay timer expiration.
// This will allow a transition to the appropriate stopped state, if
// successful.
func stoppingTimerExpiry(ctx context.Context, machine *sm.SM, m sm.Envelope) bool {
	return timerExpiration(ctx, machine, m, "Shutdown")
}

// stoppingOnLeave ensures that any active time is canceled before proceeding
// to a non-stopping state.
func stoppingOnLeave(ctx context.Context, machine *sm.SM, nextState sm.StateIndex) {
	if nextState != pb.BladeState_stopping &&
		nextState != pb.BladeState_stopping_isolated {
		cancelTimer(ctx, machine, "shutdown")
	}
}

// faultedEnter cancels any outstanding timers as a belt-and-braces practice,
// given that faulted is the state that the blade transitions to on any of
// several error paths.
func faultedEnter(ctx context.Context, machine *sm.SM) error {
	cancelTimer(ctx, machine, "outstanding")
	return nil
}

// bladeGetStatus returns a summary of this blade's current execution status.
func bladeGetStatus(ctx context.Context, machine *sm.SM, m sm.Envelope) bool {
	b := machine.Parent.(*blade)

	ch := m.Ch()
	defer close(ch)

	status := &messages.BladeStatus{
		StatusBody: messages.StatusBody{
			State:     b.sm.CurrentIndex.String(),
			EnteredAt: b.sm.EnteredAt,
		},
		Capacity:  b.capacity.Clone(),
		Used:      b.used.Clone(),
		Workloads: []string{},
	}

	for k := range b.workloads {
		status.Workloads = append(status.Workloads, k)
	}

	tracing.AddImpact(ctx, tracing.ImpactRead, machine.Name)

	ch <- messages.NewStatusResponse(common.TickFromContext(ctx), status)
	return true
}

// setConnection processes the incoming set connection message, returning true
// if the connection is to be on, false if it is to be off.  Absent any earlier
// filtering this can result in null transitions (state-a to state-a), which
// the state machine needs to be prepared for.
func setConnection(ctx context.Context, machine *sm.SM, m sm.Envelope) bool {
	msg := m.(*messages.SetConnection)

	tracing.UpdateSpanName(
		ctx,
		"Processing network connection %s notification at %s",
		common.AOrB(msg.Enabled, "enabled", "disabled"),
		msg.Target.Describe())

	tracing.AddImpact(ctx, tracing.ImpactModify, machine.Name)

	machine.AdvanceGuard(common.TickFromContext(ctx))

	return msg.Enabled
}

// setPower processes the incoming set power on/off message, returning true
// if the power is to be on, false if it is to be off.  Absent any earlier
// filtering this can result in null transitions (state-a to state-a), which
// the state machine needs to be prepared for.
func setPower(ctx context.Context, machine *sm.SM, m sm.Envelope) bool {
	msg := m.(*messages.SetPower)

	tracing.UpdateSpanName(
		ctx,
		"Processing power %s command at %s",
		common.AOrB(msg.On, "on", "off"),
		msg.Target.Describe())

	tracing.AddImpact(ctx, tracing.ImpactModify, machine.Name)

	machine.AdvanceGuard(common.TickFromContext(ctx))

	return msg.On
}

// --- blade state machine actions

// +++ support functions

// setTimer establishes a new timer if one is not currently active.
func setTimer(ctx context.Context, machine *sm.SM, delay int64) error {
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
		b.expiration = occursAt + delay
	}

	return nil
}

// timerExpiration processes a timer expiration, performing validation that it
// is still expected, and cleanup of the blade's context.
func timerExpiration(
	ctx context.Context,
	machine *sm.SM,
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

		tracing.AddImpact(ctx, tracing.ImpactModify, machine.Name)

		machine.AdvanceGuard(common.TickFromContext(ctx))
		return true
	}

	return false
}

// cancelTimer attempts to cancel an outstanding timer, if one exists.  It
// clears the internal flag that denotes a timer is expected.  This ensures
// that the timer is treated as canceled, regardless of whether or not it was
// successfully canceled.
func cancelTimer(ctx context.Context, machine *sm.SM, name string) {
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
