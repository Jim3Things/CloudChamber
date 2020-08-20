package tracing_sink

import (
    "context"
    "log"
    "net"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "google.golang.org/grpc"
    "google.golang.org/grpc/metadata"
    "google.golang.org/grpc/test/bufconn"

    clienttrace "github.com/Jim3Things/CloudChamber/internal/tracing/client"
    "github.com/Jim3Things/CloudChamber/internal/tracing/exporters"
    st "github.com/Jim3Things/CloudChamber/internal/tracing/server"
    "github.com/Jim3Things/CloudChamber/internal/tracing/setup"
    pb "github.com/Jim3Things/CloudChamber/pkg/protos/trace_sink"
)

const bufSize = 1024 * 1024

var lis *bufconn.Listener
var client pb.TraceSinkClient

func init() {
    setup.Init(exporters.UnitTest)

    lis = bufconn.Listen(bufSize)
    s := grpc.NewServer(grpc.UnaryInterceptor(st.Interceptor))
    if err := Register(s); err != nil {
        log.Fatalf("Failed to register wither error: %v", err)
    }

    go func() {
        if err := s.Serve(lis); err != nil {
            log.Fatalf("Server exited with error: %v", err)
        }
    }()
}

func bufDialer(_ context.Context, _ string) (net.Conn, error) {
    return lis.Dial()
}

func commonSetup(t *testing.T) (context.Context, *grpc.ClientConn) {
    conn, err := grpc.Dial(
        "bufnet",
        grpc.WithContextDialer(bufDialer),
        grpc.WithInsecure(),
        grpc.WithUnaryInterceptor(clienttrace.Interceptor))
    assert.Nilf(t, err, "Failed to dial bufnet: %v", err)

    md := metadata.Pairs(
        "timestamp", time.Now().Format(time.StampNano),
        "client-id", "web-api-client-us-east-1",
        "user-id", "some-test-user-id",
    )
    ctx := metadata.NewOutgoingContext(context.Background(), md)

    client = pb.NewTraceSinkClient(conn)

    return ctx, conn
}



