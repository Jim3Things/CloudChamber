// Package tracing_sink implements the trace sink service.  This service acts
// as a concentrator and store for the traces collected by Cloud Chamber.  It
// then is used by the cloud chamber UI to query and retrieve the full trace
// data stream.

package tracing_sink

import (
	"container/list"
	"context"
	"sync"

	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"

	"github.com/Jim3Things/CloudChamber/internal/clients/timestamp"
	"github.com/Jim3Things/CloudChamber/internal/common"
	"github.com/Jim3Things/CloudChamber/internal/tracing"
	"github.com/Jim3Things/CloudChamber/pkg/protos/log"
	pb "github.com/Jim3Things/CloudChamber/pkg/protos/services"
)

// server defines the interface and associated state for a trace sink instance
type server struct {
	pb.UnimplementedTraceSinkServer

	// mutex protects access to the rest of the state
	mutex sync.Mutex

	// entries is the set of known trace entries, ordered by arrival
	entries *list.List

	// waiters are the set of outstanding GetAfter waiting callers.  They are
	// indexed by the required trace entry id.  The structure allows for
	// multiple waiters to be waiting for the same trace entry id.
	waiters map[int][]waiter

	// maxHeld contains the maximum number of trace entries that are kept.  As
	// more arrive, the oldest ones are removed in order stay within this
	// limit.
	maxHeld int

	// nextId is the trace entry id that will be used on the next Append call
	nextId int

	// nextNonInternalId is the trace entry id that immediately follows the
	// last trace entry that was not marked internal.
	//
	// This is used to gate the release of outstanding waiters, avoiding the
	// calls to request the latest traces themselves triggering an immediate
	// response.
	nextNonInternalId int
}

// listEntry is used to track and identify a trace entry stored by the sink
type listEntry struct {
	id    int
	entry *log.Entry
}

// waitResponse contains the state needed to complete a stalled GetAfter call.
type waitResponse struct {
	// err contains the asynchronous error state, or nil
	err error

	// res contains the return structure to use
	res *pb.GetAfterResponse
}

// waiter contains the state needed to track a stalled GetAfter call: the
// limits on the size of the response, and how to signal the completion the
// operation
type waiter struct {
	maxEntries int64
	ch         chan waitResponse
}

// Register instantiates a sink service instance, and registers it with the
// grpc service.
func Register(svc *grpc.Server) error {
	// Create the trace sink server object
	sink := &server{
		mutex:             sync.Mutex{},
		entries:           list.New(),
		waiters:           make(map[int][]waiter),
		maxHeld:           100,
		nextId:            0,
		nextNonInternalId: 0,
	}

	// .. then register it with the grpc service
	pb.RegisterTraceSinkServer(svc, sink)
	return nil
}

// Append adds a trace entry to the list of known entries.
func (s *server) Append(_ context.Context, request *pb.AppendRequest) (*empty.Empty, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if err := request.Validate(); err != nil {
		return &empty.Empty{}, err
	}

	entry := request.Entry
	item := listEntry{
		id:    s.nextId,
		entry: entry,
	}

	s.entries.PushBack(item)

	if s.entries.Len() > s.maxHeld {
		s.entries.Remove(s.entries.Front())
	}

	s.nextId++
	if !entry.Infrastructure {
		s.nextNonInternalId = s.nextId
	}

	s.signalWaiters()

	return &empty.Empty{}, nil
}

// GetAfter retrieves trace entries starting after the supplied ID.  The
// caller also specifies the maximum number of entries to return in one call,
// and whether or not to wait if there are not entries currently outstanding.
func (s *server) GetAfter(ctx context.Context, request *pb.GetAfterRequest) (*pb.GetAfterResponse, error) {
	var resp waitResponse
	var err error = nil

	ctx, span := tracing.StartSpan(context.Background(),
		tracing.WithContextValue(clients.EnsureTickInContext),
		tracing.AsInternal())

	defer func() {
		// Pick up the current time to avoid repeatedly fetching the same value
		ctx = common.ContextWithTick(ctx, clients.Tick(ctx))

		if err != nil {
			tracing.Warnf(ctx, "GetAfter failed: %v", err)
		} else {
			tracing.Infof(ctx, "GetAfter returning; %d entries, missed=%v, lastID=%d", len(resp.res.Entries), resp.res.Missed, resp.res.LastId)
		}

		span.End()
	}()

	if err = request.Validate(); err != nil {
		return resp.res, err
	}

	s.mutex.Lock()

	// If we either can't wait, or there are active traces to return,
	// do so now.
	id := request.Id + 1
	if !request.Wait || (id < int64(s.nextNonInternalId)) {
		resp = s.processWaiter(id, request.MaxEntries)

		s.mutex.Unlock()
		return resp.res, err
	}

	ch := s.wait(id, request.MaxEntries)

	s.mutex.Unlock()

	resp = <-ch

	return resp.res, resp.err
}

// GetPolicy supplies information on the range of trace entries held by the
// sink service, and the limits on how many it will retain.
func (s *server) GetPolicy(ctx context.Context, _ *pb.GetPolicyRequest) (*pb.GetPolicyResponse, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var resp *pb.GetPolicyResponse

	ctx, span := tracing.StartSpan(context.Background(),
		tracing.WithContextValue(clients.EnsureTickInContext),
		tracing.AsInternal())
	defer span.End()

	resp = &pb.GetPolicyResponse{
		MaxEntriesHeld: int64(s.maxHeld),
		FirstId:        int64(s.getFirstId() - 1),
	}

	tracing.Infof(ctx, "GetPolicy returning; firstId=%d, maxEntriesHeld=%d", resp.FirstId, resp.MaxEntriesHeld)

	return resp, nil
}

// Reset is a test support function that forcibly resets the instance's state
// to its initial values.
func (s *server) Reset(_ context.Context, _ *pb.ResetRequest) (*empty.Empty, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.entries = list.New()
	s.nextId = 0
	s.nextNonInternalId = 0
	s.waiters = make(map[int][]waiter)

	return &empty.Empty{}, nil
}

// signalWaiters looks for waiters who are waiting for new trace entries that
// are now within the set held by the service, and processes them.
func (s *server) signalWaiters() {
	for id, waiters := range s.waiters {
		if id < s.nextNonInternalId {

			// We have something to fill in, so handle it now.
			for _, entry := range waiters {
				rsp := s.processWaiter(int64(id), entry.maxEntries)
				entry.ch <- waitResponse{
					err: rsp.err,
					res: rsp.res,
				}
			}

			delete(s.waiters, id)
		}
	}
}

// wait sets up a blocked request using the supplied values, returning
// the channel to wait on for the final outcome.
func (s *server) wait(id int64, maxEntries int64) chan waitResponse {
	ch := make(chan waitResponse)

	item := waiter{
		maxEntries: maxEntries,
		ch:         ch,
	}

	waiters, ok := s.waiters[int(id)]
	if !ok {
		waiters = []waiter{}
	}

	waiters = append(waiters, item)
	s.waiters[int(id)] = waiters

	// As a final step, quick check if any outstanding waiters can be
	// completed, including this one.
	s.signalWaiters()

	return ch
}

// processWaiter runs through the outstanding trace entries that are at or after
// the startID, up to the maximum number.  It assembles and returns them in a
// reply packet that can be sent back to the caller.
func (s *server) processWaiter(startID int64, maxEntries int64) waitResponse {
	resp := waitResponse{
		err: nil,
		res: &pb.GetAfterResponse{
			LastId:  startID - 1,
			Missed:  false,
			Entries: []*pb.GetAfterResponseTraceEntry{},
		},
	}

	// Entries have been missed if there are entries saved, but the oldest is
	// newer than the starting point
	if s.entries.Len() > 0 {
		resp.res.Missed = int64(s.entries.Front().Value.(listEntry).id) > startID
	}

	var count int64 = 0

	for e := s.entries.Front(); (e != nil) && (count < maxEntries); e = e.Next() {
		item := e.Value.(listEntry)

		id := int64(item.id)
		if id >= startID {
			resp.res.Entries = append(resp.res.Entries, &pb.GetAfterResponseTraceEntry{
				Id:    id,
				Entry: item.entry,
			})

			count++
		}

		resp.res.LastId = id
	}

	return resp
}

// getFirstId is a helper function that returns the oldest id currently
// held in the store (or zero, if none are).
func (s *server) getFirstId() int {
	if s.entries.Len() > 0 {
		return s.entries.Front().Value.(listEntry).id
	}

	return 0
}
