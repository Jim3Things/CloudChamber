package production

import (
    "context"
    "errors"
    "fmt"
    "log"
    "sync"

    export "go.opentelemetry.io/otel/sdk/export/trace"
    "google.golang.org/grpc"

    "github.com/Jim3Things/CloudChamber/internal/tracing/exporters/common"
    pb "github.com/Jim3Things/CloudChamber/pkg/protos/trace_sink"
    pbl "github.com/Jim3Things/CloudChamber/pkg/protos/log"
)

const (
    reconnectBackoff = 100
    maxBackoff = 5000
)

var (
    mutex = sync.Mutex{}

    configSet = false
    endpoint string
    opts []grpc.DialOption

    conn *grpc.ClientConn = nil
    client pb.TraceSinkClient

    // Track the number of times that we attempt to connect to the trace server
    connectCount int
    reconnectInProgress = false
)

func SetEndpoint(name string, dialOptions ...grpc.DialOption) error {
    mutex.Lock()
    defer mutex.Unlock()

    if configSet {
        return fmt.Errorf("endpoint configuration already set to %q, options: %v", endpoint, opts)
    }

    endpoint = name
    copy(opts, dialOptions)
    configSet = true

    return nil
}

// Exporter is an implementation of trace.Exporter that writes spans to io_writer.
type Exporter struct {
    queue *common.Deferrable
}

func NewExporter() (*Exporter, error) {
    return &Exporter{
        queue: common.NewDeferrable(0),
    }, nil
}

// ExportSpan writes a SpanData in json format to io_writer.
func (e *Exporter) ExportSpan(ctx context.Context, data *export.SpanData) {
    entry := common.ExtractEntry(ctx, data)

    mutex.Lock()
    defer mutex.Unlock()

    if err := e.connectIf(ctx); err != nil {
        e.deferOne(entry)
    }

    e.postOne(ctx, entry)
}

func (e *Exporter) connectIf(ctx context.Context) error {
    if !configSet || reconnectInProgress {
        return errors.New("not yet ready")
    }

    return nil
}

func (e *Exporter) flushDeferred(ctx context.Context) error {
    err := e.queue.Flush(ctx, func (ctx context.Context, entry *pbl.Entry) error {
        _, err := client.Append(ctx, &pb.AppendRequest{ Entry: entry })

        return err
    })

    if err != nil {
        client = nil
        _ = conn.Close()
        conn = nil

        if err2 := e.scheduleReconnect(); err2 != nil {
            log.Fatalf("unable to reschedule connect, fatal error.  err=%v", err2)
        }
    }

    return nil
}

func (e *Exporter) scheduleReconnect() error {
    if reconnectInProgress {
        return nil
    }

    return nil
}

func (e *Exporter) deferOne(data *pbl.Entry) {
    if err := e.queue.Defer(data); err != nil {
        log.Fatalf("unable to defer an entry, fatal error.  err=%v", err)
    }

    if err2 := e.scheduleReconnect(); err2 != nil {
        log.Fatalf("unable to reschedule connect, fatal error.  err=%v", err2)
    }
}

func (e *Exporter) postOne(ctx context.Context, data *pbl.Entry) {
    _, err := client.Append(ctx, &pb.AppendRequest{Entry: data})
    if err != nil {
        client = nil

        if conn != nil {
            _ = conn.Close()
            conn = nil
        }

        e.deferOne(data)
    }
}
