package tracing_sink

// this is a skeleton trace sink service.  It has a basic Append implementation
// but nothing else.  It emits the traces to its local stdout as they arrive.
// This is only an interim service for this current development step.

import (
	"container/list"
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"

	"github.com/Jim3Things/CloudChamber/internal/common"
	"github.com/Jim3Things/CloudChamber/internal/tracing"
	st "github.com/Jim3Things/CloudChamber/internal/tracing/server"
	"github.com/Jim3Things/CloudChamber/pkg/protos/log"
	pb "github.com/Jim3Things/CloudChamber/pkg/protos/trace_sink"
)

type listEntry struct {
	id    int
	entry *log.Entry
}

type waitResponse struct {
	err error
	res *pb.GetAfterResponse
}

type waiter struct {
	maxEntries int64
	ch         <-chan waitResponse
}

type server struct {
	pb.UnimplementedTraceSinkServer

	mutex   sync.Mutex
	entries *list.List
	waiters map[int][]interface{}

	maxHeld int
	firstId int
	lastId  int
}

var sink *server

func Register(svc *grpc.Server) error {
	// Create the trace sink server object
	sink = &server{
		mutex:   sync.Mutex{},
		entries: list.New(),
		waiters: make(map[int][]interface{}),
		maxHeld: 100,
		firstId: 0,
		lastId:  0,
	}

	// .. then register it with the grpc service
	pb.RegisterTraceSinkServer(svc, sink)
	return nil
}

// LocalAppend adds a trace entry to list of known entries, but does so without
// invoking the grpc channel, or any other feature that will itself produce new
// trace entries.
//
// This is intended to provide a mechanism for the trace sink support services,
// and other simulation support services, to trace their activity without
// creating a recursive trace production stream.
func LocalAppend(ctx context.Context, entry *log.Entry) error {
	return sink.LocalAppend(ctx, entry)
}

// LocalAppend is the trace sink instance's implementation of the global
// function declared above.
func (s *server) LocalAppend(_ context.Context, entry *log.Entry) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.addEntry(entry)

	return nil
}

// Append adds a trace entry to the list of known entries.
func (s *server) Append(ctx context.Context, request *pb.AppendRequest) (*empty.Empty, error) {
	err := st.WithSpan(ctx, tracing.MethodName(1), func(ctx context.Context) error {
		if err := request.Validate(); err != nil {
			return err
		}

		return s.LocalAppend(ctx, request.Entry)
	})

	return &empty.Empty{}, err
}

func (s *server) GetAfter(ctx context.Context, request *pb.GetAfterRequest) (*pb.GetAfterResponse, error) {
	var resp waitResponse

	err := st.WithSpan(ctx, tracing.MethodName(1), func(ctx context.Context) error {
		if err := request.Validate(); err != nil {
			return err
		}

		if !request.Wait {
			resp = s.processWaiter(request.Id, request.MaxEntries)
			return resp.err
		}

		resp = <-s.wait(request)
		return resp.err
	})

	if err != nil {
		st.Infof(
			ctx, common.Tick(),
			"GetAfter failed with err=%v", err)
	} else {
		st.Infof(
			ctx, common.Tick(),
			"GetAfter returning; %d entries, missed=%v, lastID=%d",
			len(resp.res.Entries), resp.res.Missed, resp.res.LastId)
	}

	return resp.res, err
}

func (s *server) GetPolicy(ctx context.Context, _ *pb.GetPolicyRequest) (*pb.GetPolicyResponse, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var resp *pb.GetPolicyResponse

	err := st.WithSpan(ctx, tracing.MethodName(1), func(ctx context.Context) error {
		resp = &pb.GetPolicyResponse{
			MaxEntriesHeld: int64(s.maxHeld),
			FirstId:        int64(s.firstId),
		}

		return nil
	})

	return resp, err
}

func (s *server) SetPolicy(ctx context.Context, request *pb.SetPolicyRequest) (*pb.SetPolicyResponse, error) {
	return nil, fmt.Errorf("not yet implemented")
}

func (s *server) Reset(_ context.Context, _ *pb.ResetRequest) (*empty.Empty, error)  {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.entries = list.New()
	s.lastId = 0
	s.firstId = 0
	s.waiters = make(map[int][]interface{})

	return &empty.Empty{}, nil
}

func (s *server) addEntry(entry *log.Entry) {
	item := listEntry{
		id:    s.lastId,
		entry: entry,
	}

	s.entries.PushBack(item)

	if s.entries.Len() > s.maxHeld {
		firstEntry := s.entries.Front().Value.(listEntry)

		fmt.Printf("    : Deleting entry %d\n\n", firstEntry.id)

		s.entries.Remove(s.entries.Front())
		s.firstId = s.entries.Front().Value.(listEntry).id
	}

	fmt.Printf("%s: %d(%d): %v\n\n", time.Now().Format(time.RFC822), s.lastId, s.entries.Len(), entry)
	s.lastId++

	s.signalWaiters()
}

func (s *server) signalWaiters() {

}

func (s *server) wait(request *pb.GetAfterRequest) <-chan waitResponse {
	ch := make(<-chan waitResponse)

	return ch
}

// processWaiter runs through the outstanding trace entries that are after the
// startID, up to the maximum number.  It assembles and returns them in a reply
// packet that can be sent back to the caller.
func (s *server) processWaiter(startID int64, maxEntries int64) waitResponse {
	resp := waitResponse{
		err: nil,
		res: &pb.GetAfterResponse{
			LastId:  startID,
			Missed:  false,
			Entries: []*pb.GetAfterResponseTraceEntry{},
		},
	}

	var count int64 = 0

	for e := s.entries.Front(); (e != nil) && (count <= maxEntries); e = e.Next() {
		item, ok := e.Value.(listEntry)
		if !ok {
			resp.err = fmt.Errorf("unexpected type in list: %v", e.Value)
			return resp
		}

		id := int64(item.id)
		if id > startID {
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
