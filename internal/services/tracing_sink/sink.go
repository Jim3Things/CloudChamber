package tracing_sink

import (
    "context"
    "fmt"
    "sync"

    "google.golang.org/grpc"

    "github.com/Jim3Things/CloudChamber/pkg/protos/log"
    pb "github.com/Jim3Things/CloudChamber/pkg/protos/trace_sink"
)

type server struct {
    pb.UnimplementedTraceSinkServer

    mutex sync.Mutex
    entries map[int64]log.Entry
    waiters map[int64][]interface{}

    maxHeld int64
    firstId int64
    lastId int64
}

var sink *server

func Register(svc *grpc.Server) error {
    // Create the trace sink server object
    sink = &server{
        mutex:                        sync.Mutex{},
        entries:                      make(map[int64]log.Entry),
        waiters:                      make(map[int64][]interface{}),
        maxHeld:                      100,
        firstId:                      0,
        lastId:                       0,
    }

    // .. then register it with the grpc service
    pb.RegisterTraceSinkServer(svc, sink)
    return nil
}

func LocalAppend(ctx context.Context, entry *log.Entry) error {
    return sink.LocalAppend(ctx, entry)
}

func (s *server) Append(ctx context.Context, request *pb.AppendRequest) (*pb.AppendResponse, error) {
    return nil, fmt.Errorf("not yet implemented")
}

func (s *server) GetAfter(ctx context.Context, request *pb.GetAfterRequest) (*pb.GetAfterResponse, error) {
    return nil, fmt.Errorf("not yet implemented")
}

func (s *server) GetPolicy(ctx context.Context, request *pb.GetPolicyRequest) (*pb.GetPolicyResponse, error) {
    return nil, fmt.Errorf("not yet implemented")
}

func (s *server) SetPolicy(ctx context.Context, request *pb.SetPolicyRequest) (*pb.SetPolicyResponse, error) {
    return nil, fmt.Errorf("not yet implemented")
}

func (s *server) LocalAppend(ctx context.Context, entry *log.Entry) error {
    return nil
}