package timestamp

import (
	"context"
	"fmt"
	"sync"
	"time"

	"google.golang.org/grpc"

	"github.com/Jim3Things/CloudChamber/simulation/internal/tracing"
	ct "github.com/Jim3Things/CloudChamber/simulation/pkg/protos/common"
	pb "github.com/Jim3Things/CloudChamber/simulation/pkg/protos/services"
)

const (
	// maxRetryLimit is the maximum backoff value.  The increasing backoff will
	// level out at this value, irregardless of how long the stepper is not
	// available.
	maxRetryLimit = 5

	// perRetryDuration is the number of milliseconds to wait per backoff
	// increment.
	perRetryDuration = 100
)
// Ticker is a mechanism that continuously requests notification of the next
// simulated tick, until it is ordered to stop.
type Ticker struct {
	// ch is the channel on which to send new tick events.  This is a fan-in
	// channel, so this instance is not responsible for closing the channel.
	// That is left up to the consumer that knows how many sources are still
	// active.
	ch chan *tickEvent

	// connection data for contacting the simulated time service.
	dialName string
	dialOpts []grpc.DialOption

	// stop determines whether the goroutine should continue. When true, the
	// polling goroutine terminates at the next logical point.  It has a mutex
	// associated with it to ensure that the output channel is not closed
	// while an event is being posted.
	stop bool
	m    sync.Mutex
}

// NewTicker creates a new timer collection instance.  The configuration
// parameter provides endpoint information for the simulated time service.
func NewTicker(ch chan *tickEvent, ep string, dialOpts ...grpc.DialOption) *Ticker {
	t := &Ticker{
		ch:       ch,
		dialName: ep,
		dialOpts: dialOpts,
		stop:     false,
		m:        sync.Mutex{},
	}

	go t.listener()

	return t
}

// Stop orders the ticker goroutine to stop.  This is a lazy operation, where
// upon returning it means that the order to stop has been sent, and no further
// messages will be posted on the ticker event channel.
func (t *Ticker) Stop() {
	t.m.Lock()
	defer t.m.Unlock()

	t.stop = true
}

// listener is the goroutine that waits for a new simulated time Tick, and
// then processes each expired timer.
func (t *Ticker) listener() {
	now := int64(-1)
	startCtx := context.Background()
	worked := false

	_, span := tracing.StartSpan(
		context.Background(),
		tracing.WithName("Starting ticker goroutine"),
		tracing.AsInternal(),
		tracing.WithContextValue(OutsideTime))
	span.End()

	for retries := 0; !t.stop; {
		retries = t.waitBeforeReconnect(retries)
		now, worked = t.listenUntilFailure(startCtx, now)

		if worked {
			// At least one call worked, so restart the delay interval
			// back at its initial state.
			retries = 0
		}
	}

	_, span = tracing.StartSpan(
		context.Background(),
		tracing.WithName("Stopping ticker goroutine"),
		tracing.AsInternal(),
		tracing.WithContextValue(OutsideTime))
	span.End()

	close(t.ch)
}

// listenUntilFailure is the main worker logic in the listener goroutine.  It
// wakes after each simulated time Tick and signals the event.  It continues
// until there is an error in contacting the simulated time service.  Any
// decision to resume after some interval or exit is then made by the caller.
func (t *Ticker) listenUntilFailure(ctx context.Context, now int64) (int64, bool) {
	conn, err := grpc.Dial(t.dialName, t.dialOpts...)
	defer func() { _ = conn.Close() }()

	worked := false

	client := pb.NewStepperClient(conn)

	for err == nil && !t.stop {
		var resp *pb.StatusResponse

		resp, err = client.Delay(ctx, &pb.DelayRequest{
			AtLeast: &ct.Timestamp{Ticks: now + 1},
			Jitter:  0,
		})

		if err == nil {
			now = resp.Now
			worked = true

			t.post(&tickEvent{
				tick: resp,
				ack:  nil,
			})
		}
	}

	return now, worked
}

// post writes the tick event to the output channel, unless a stop order has
// been issued.  In that case, the event is ignored.
func (t *Ticker) post(ev *tickEvent) {
	t.m.Lock()
	defer t.m.Unlock()

	if !t.stop {
		t.ch <- ev
	}
}

// waitBeforeReconnect waits for an increasing backoff period to allow time for
// the stepper service to recover from any transient connection issues.
func (*Ticker) waitBeforeReconnect(retries int) int {
	_, span := tracing.StartSpan(
		context.Background(),
		tracing.WithName(fmt.Sprintf("reconnect wait, retries=%d", retries)),
		tracing.AsInternal(),
		tracing.WithContextValue(OutsideTime))
	span.End()

	if retries > 0 {
		time.Sleep(time.Duration(retries*perRetryDuration) * time.Millisecond)
	}

	retries++
	if retries > maxRetryLimit {
		retries = maxRetryLimit
	}

	return retries
}
