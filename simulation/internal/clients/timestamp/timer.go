package timestamp

import (
	"context"
	"sync"
	"time"

	"github.com/Jim3Things/CloudChamber/simulation/internal/common"
	"github.com/Jim3Things/CloudChamber/simulation/pkg/errors"
	ct "github.com/Jim3Things/CloudChamber/simulation/pkg/protos/common"
	pb "github.com/Jim3Things/CloudChamber/simulation/pkg/protos/services"

	"google.golang.org/grpc"
)

// timerEntry describes a single active timer
type timerEntry struct {
	// id is the unique key assigned to this timer
	id common.PrimaryKey

	// dueTime is the simulated time when this timer expires
	dueTime common.SecondaryKey

	// ch is the channel that is to receive the expiration message
	callback func(msg interface{})

	// msg is the expiration message specified for this timer
	msg interface{}
}

func (t *timerEntry) Primary() common.PrimaryKey { return t.id }
func (t *timerEntry) Secondary(index int) common.SecondaryKey {
	if index == 0 {
		return t.dueTime
	} else {
		return nil
	}
}

// Timers is
type Timers struct {
	m sync.Mutex

	// waiters is the collection of current outstanding timers.
	waiters common.MultiMap

	// nextID holds the timer ID to assign to the next timer created.
	nextID int

	// active indicates whether or not the listener goroutine is currently
	// running.
	active bool

	// epoch is the running instance version of the listener goroutine, used to
	// detect the need to suicide by a goroutine that should be exiting.
	epoch int

	// connection data for contacting the simulated time service.
	dialName string
	dialOpts []grpc.DialOption
}

// NewTimers creates a new timer collection instance.  The configuration
// parameter provides endpoint information for the simulated time service.
func NewTimers(ep string, dialOpts ...grpc.DialOption) *Timers {
	return &Timers{
		m:        sync.Mutex{},
		waiters:  common.NewMultiMap(1),
		nextID:   1,
		active:   false,
		epoch:    1,
		dialName: ep,
		dialOpts: dialOpts,
	}
}

// Timer creates a new timer that operates in simulated time. The delay
// parameter specifies the number of ticks to wait until the timer expires. At
// that point, the supplied msg is sent on the completion channel specified by
// the parameter ch.  This function returns an id that can be used to cancel
// the timer, and an error to indicate if the timer was successfully set.
func (t *Timers) Timer(ctx context.Context, delay int64, msg interface{}, callback func(msg interface{})) (int, error) {
	t.m.Lock()
	defer t.m.Unlock()

	now := common.TickFromContext(ctx)
	entry := &timerEntry{
		id:       common.PrimaryKey(t.nextID),
		dueTime:  delay + now,
		callback: callback,
		msg:      msg,
	}

	t.waiters.Add(entry)

	t.nextID++

	if !t.active {
		t.active = true

		go t.listener(t.epoch, now)
	}

	return int(entry.id), nil
}

// Cancel removes the designated waiting timer, or returns an error if it is
// not found.
func (t *Timers) Cancel(timerID int) error {
	t.m.Lock()
	defer t.m.Unlock()

	if _, ok := t.waiters.Remove(common.PrimaryKey(timerID)); !ok {
		return errors.ErrTimerNotFound(timerID)
	}

	_ = t.mayCancelListener()

	return nil
}

// listener is the goroutine that waits for a new simulated time tick, and
// then processes each expired timer.
func (t *Timers) listener(epoch int, now int64) {
	startCtx := context.Background()
	retries := 0

	t.m.Lock()

	for t.epoch == epoch {
		t.m.Unlock()
		now = t.listenUntilFailure(startCtx, epoch, now)

		retries = waitBeforeReconnect(retries)
		t.m.Lock()
	}

	t.m.Unlock()
}

// listenUntilFailure is the main worker logic in the listener goroutine.  It
// wakes after each simulated time tick and signals all expired waiters.  It
// continues until either there are no more waiters, or until there is an error
// in contacting the simulated time service.  Any decision to resume after some
// interval or exit is then made by the caller.
func (t *Timers) listenUntilFailure(ctx context.Context, epoch int, now int64) int64 {
	conn, err := grpc.Dial(t.dialName, t.dialOpts...)
	defer func() { _ = conn.Close() }()

	client := pb.NewStepperClient(conn)

	for stop := false; err == nil && !stop; {
		var resp *pb.StatusResponse

		resp, err = client.Delay(ctx, &pb.DelayRequest{
			AtLeast: &ct.Timestamp{Ticks: now + 1},
			Jitter:  0,
		})

		if err == nil {
			var toSignal []*timerEntry

			now = resp.Now

			if toSignal, stop = t.getExpiredWaiters(now, epoch); toSignal != nil {
				for _, entry := range toSignal {
					entry.callback(entry.msg)
				}
			}
		}
	}

	return now
}

// getExpiredWaiters looks through the set of outstanding waiters, pulling out
// those that have expired to return to the caller.  It also signals whether or
// not the listener can exit because there are no remaining waiters.  If that
// is so, it returns true as the second return value.
func (t *Timers) getExpiredWaiters(now int64, epoch int) ([]*timerEntry, bool) {
	t.m.Lock()
	defer t.m.Unlock()

	// check if this listener was ordered to exit while we did not hold the
	// mutex.
	if t.epoch != epoch {
		return nil, true
	}

	// find every waiter that has expired, cleaning the waiting state maps
	// as the processing proceeds.
	var toSignal []*timerEntry

	t.waiters.ForEachSecondary(0,
		func(key common.SecondaryKey, items []common.PrimaryKey) {
			due := key.(int64)
			if due <= now {
				for _, item := range items {
					entry, ok := t.waiters.Get(item)
					if ok {
						toSignal = append(toSignal, entry.(*timerEntry))
					}
				}
			}
		})

	for _, entry := range toSignal {
		t.waiters.Remove(entry.Primary())
	}

	return toSignal, t.mayCancelListener()
}

// mayCancelListener checks if there are any waiters.  If there are none, then
// it signals that the current listener goroutine should exit.
func (t *Timers) mayCancelListener() bool {
	if t.waiters.Count() == 0 {
		t.epoch++
		t.active = false

		return true
	}

	return false
}

func waitBeforeReconnect(retries int) int {
	retries++
	if retries > 5 {
		retries = 5
	}

	time.Sleep(time.Duration(retries*100) * time.Millisecond)

	return retries
}
