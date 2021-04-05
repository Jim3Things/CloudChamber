package timestamp

import (
	"context"
	"fmt"
	"sync"

	"github.com/golang/protobuf/ptypes/duration"
	"google.golang.org/grpc"

	"github.com/Jim3Things/CloudChamber/simulation/internal/common"
	"github.com/Jim3Things/CloudChamber/simulation/internal/tracing"
	"github.com/Jim3Things/CloudChamber/simulation/pkg/errors"
	pb "github.com/Jim3Things/CloudChamber/simulation/pkg/protos/services"
)

// This module contains the implementation for the Listener class and related
// communication structures.  The Listener provides a goroutine that processes
// changes in stepper status, including the simulated time tick, and issues
// notifications to channels that have subscribed to be notified of such changes
// as one-shot events.

// psItem defines the common interface that all order structs implement.
//
// The underlying data type is used as a discriminant to determine which order
// the instance represents.
type psItem interface {
	common.BimapItem
	GetNotify() chan Completion
}

// waiter describes a single registration for a notification when the simulated
// time matches or exceeds the requested due time.
type waiter struct {
	id      int
	dueTime int64
	name    string
	notify  chan Completion
}

func (w *waiter) Key() int                   { return w.id }
func (w *waiter) Secondary() int64           { return w.dueTime }
func (w *waiter) GetNotify() chan Completion { return w.notify }

// canceler describes an order to cancel a previously registered waiter.
type canceler struct {
	id     int
	notify chan Completion
}

func (w *canceler) Key() int                   { return w.id }
func (w *canceler) Secondary() int64           { return -1 }
func (w *canceler) GetNotify() chan Completion { return w.notify }

// Completion contains the final result of the registered event notification.
// It is either an updated status, or an error that arose during processing.
type Completion struct {
	Status *pb.StatusResponse
	Err    error
}

// tickEvent contains an incoming update to the simulated time status, and the
// optional acknowledgement channel that can be used to verify that the status
// has been updated.
type tickEvent struct {
	tick *pb.StatusResponse
	ack  chan bool
}

func (t *tickEvent) String() string {
	return fmt.Sprintf(
		"Status: %s, ack needed: %v",
		t.tick.Describe(),
		t.ack != nil)
}

// Listener defines the data associated with a running instance.
type Listener struct {
	m       sync.Mutex // Protect the working state flag
	working bool       // True, if new requests can be added.

	tick chan *tickEvent // Internal channel to receive status changes
	ps   chan psItem     // Internal channel to receive registration changes

	stopper chan bool // Order to stop
	stopped chan bool // goroutine stopped, no more events will be sent

	// waiters is the collection of current outstanding timers.
	waiters *common.Bimap

	// connection data for contacting the simulated time service.
	dialName string
	dialOpts []grpc.DialOption

	now *pb.StatusResponse // Last simulated time status

	nextId int // This is the next id to assign to a subscription
}

// NewListener creates and returns a new Listener instance, with the background
// goroutine initialized and running.
func NewListener(ep string, dialOpts ...grpc.DialOption) *Listener {
	l := &Listener{
		m:        sync.Mutex{},
		tick:     make(chan *tickEvent, 2),
		ps:       make(chan psItem, 1),
		stopper:  make(chan bool, 1),
		stopped:  make(chan bool, 1),
		working:  true,
		waiters:  common.NewBimap(),
		dialName: ep,
		dialOpts: dialOpts,
		now: &pb.StatusResponse{
			Policy: pb.StepperPolicy_Invalid,
			MeasuredDelay: &duration.Duration{
				Seconds: 0,
				Nanos:   0,
			},
			Now:         -1,
			WaiterCount: 0,
			Epoch:       0,
		},
		nextId: 0,
	}

	go l.listener()

	return l
}

// After requests a notification once the simulated time has passed the dueTime
// value.  It returns the unique ID, that can be used to cancel the request, the
// channel to listen for the notification event, and an error, if the request
// could not be added.
func (l *Listener) After(name string, dueTime int64) (int, chan Completion, error) {
	l.m.Lock()
	defer l.m.Unlock()

	if !l.working {
		return 0, nil, errors.ErrTimerCanceled(-1)
	}

	notify := make(chan Completion, 1)

	l.nextId++

	s := &waiter{
		id:      l.nextId,
		dueTime: dueTime,
		name:    name,
		notify:  notify,
	}

	l.ps <- s

	return s.id, notify, nil
}

// Cancel attempts to cancel an outstanding notification request.  The id
// specifies the request to cancel.  It returns a channel to listen to for
// confirmation of the cancellation, and an error that indicates if the cancel
// request was accepted.  If the error value is nil, then the returned channel
// will receive a confirmation message once either the specified id is not
// found, or it is successfully canceled.
func (l *Listener) Cancel(id int) (chan Completion, error) {
	l.m.Lock()
	defer l.m.Unlock()

	if !l.working {
		return nil, errors.ErrTimerCanceled(id)
	}

	notify := make(chan Completion, 1)
	s := &canceler{id: id, notify: notify}

	l.ps <- s

	return notify, nil
}

// UpdateStatus is a function that posts an update to the simulated time status.
// It proceeds returns once the status has been successfully updated.
func (l *Listener) UpdateStatus(status *pb.StatusResponse) {
	l.m.Lock()
	defer l.m.Unlock()

	if l.working {
		ch := make(chan bool, 1)
		l.tick <- &tickEvent{
			tick: status,
			ack:  ch,
		}

		<-ch
	}
}

// Stop is a function that terminates the background activities of the listener,
// and marks the listener as closed, so as to prevent future requests to it.
func (l *Listener) Stop() {
	l.m.Lock()
	defer l.m.Unlock()

	if l.working {
		l.working = false

		close(l.ps)

		close(l.stopper)
		<-l.stopped
	}
}

// listener is the core of the Listener class. It is the goroutine that manages
// updates to the simulated time status, requests for notification of changes,
// cancellation of same, and handles the request to stop at the end.
func (l *Listener) listener() {
	l.log("Starting listener goroutine")
	stopping := false
	var tickMsg string

	ticker := NewTicker(l.tick, l.dialName, l.dialOpts...)

	// Before processing any new notification requests, ensure that the
	// simulate time status has been loaded at least once.  This avoids a race
	// during startup where a notification request that should be immediately
	// processed returns uninitialized state.
	select {
	case t := <-l.tick:
		l.processTick(t)
		tickMsg = t.String()

	case _, _ = <-l.stopper:
		stopping = true
		tickMsg = "Stopping"
	}

	l.log("Listener goroutine state initialized: %s", tickMsg)

	// This is the main processing loop.  Wait for either a status update, a
	// new request (or cancellation), or a stop order.  Process each as they
	// arrive.
	for !stopping {
		select {
		case t := <-l.tick:
			l.processTick(t)

		case pub, ok := <-l.ps:
			if !ok {
				stopping = true
			} else {
				l.processPubSub(pub)
			}

		case _, _ = <-l.stopper:
			stopping = true
		}
	}

	// And finally, on exit ensure that all outstanding notification requests
	// have been canceled, and that no further background activity will occur.
	l.log("Stopping listener goroutine")
	l.cancelAndDrainAll()
	ticker.Stop()

	l.stopped <- true
	close(l.stopped)
}

// processTick handles a status update.  The new state is stored, and each
// outstanding notification request is examined, and those that have met their
// requested criteria are completed.
func (l *Listener) processTick(t *tickEvent) {
	l.now = latestEvent(l.now, t.tick)
	var expired []*waiter

	l.waiters.ForEachSecondary(func(key int64) bool {
		return key <= l.now.Now
	},
		func(item common.BimapItem) {
			s := item.(*waiter)
			expired = append(expired, s)
		})

	for _, item := range expired {
		l.send(item.notify, nil)
		l.waiters.Remove(item.Key())
	}

	if t.ack != nil {
		t.ack <- true
		close(t.ack)
	}
}

// processPubSub handles a new request for notification, or the cancellation of
// an existing request.  Note that new notifications are completed immediately
// if either the dueTime has passed, or the id is already in use.
func (l *Listener) processPubSub(item interface{}) {
	switch pub := item.(type) {
	case *waiter:
		if _, ok := l.waiters.Get(pub.Key()); ok {
			l.send(pub.notify, errors.ErrTimerIdAlreadyExists(pub.Key()))
			return
		}

		if pub.dueTime <= l.now.Now {
			l.send(pub.notify, nil)
		} else {
			l.waiters.Add(pub)
		}
		break

	case *canceler:
		if old, ok := l.waiters.Get(pub.id); ok {
			// Have an established timer - cancel it
			l.send(old.(*waiter).notify, errors.ErrTimerCanceled(pub.id))

			l.waiters.Remove(pub.id)
		}

		l.send(pub.notify, nil)
		break

	default:
		l.bugCheck("Invalid message type encountered.  Msg is %v", pub)
	}
}

// cancelAndDrainAll forcibly cancels all outstanding notification requests.
func (l *Listener) cancelAndDrainAll() {
	l.waiters.ForEachSecondary(
		func(_ int64) bool { return true },
		func(item common.BimapItem) {
			s := item.(*waiter)
			l.send(s.notify, errors.ErrTimerCanceled(s.Key()))
		})

	l.waiters.Clear()

	s, ok := <-l.ps
	for ok {
		l.send(s.GetNotify(), errors.ErrTimerCanceled(s.Key()))
		s, ok = <-l.ps
	}
}

// send posts the completion event and closes the channel
func (l *Listener) send(notify chan Completion, err error) {
	notify <- Completion{
		Status: l.now,
		Err:    err,
	}
	close(notify)
}

// log is a helper function that posts an empty span with the supplied text.
func (l *Listener) log(f string, args ...interface{}) {
	_, span := tracing.StartSpan(
		context.Background(),
		tracing.WithName(fmt.Sprintf(f, args...)),
		tracing.AsInternal(),
		tracing.WithContextValue(OutsideTime))
	span.End()
}

// bugCheck is a helper function that posts a fatal error inside a containing
// bugcheck span.  There is no expectation that this will successfully return.
func (l *Listener) bugCheck(args ...interface{}) {
	ctx, span := tracing.StartSpan(
		context.Background(),
		tracing.WithName("BUGCHECK: Internal failure"),
		tracing.WithContextValue(OutsideTime))
	defer span.End()

	tracing.Fatal(ctx, args...)
}

// latestEvent is a helper function that returns the later of the two status
// response instances.  This ensures that multiple sources of status responses
// will not result in the status jumping backwards.
func latestEvent(a *pb.StatusResponse, b *pb.StatusResponse) *pb.StatusResponse {
	if a.Epoch > b.Epoch {
		return a
	}

	if a.Epoch == b.Epoch && a.Now >= b.Now {
		return a
	}

	return b
}
