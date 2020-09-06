package production

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"go.opentelemetry.io/otel/sdk/export/trace"
	"google.golang.org/grpc"

	"github.com/Jim3Things/CloudChamber/internal/tracing/exporters/common"
	pbl "github.com/Jim3Things/CloudChamber/pkg/protos/log"
	pb "github.com/Jim3Things/CloudChamber/pkg/protos/services"
)

// Note well: This exporter type assumes that the exporter is a singleton.  That is
// sufficient for Cloud Chamber at this time, but is something that will need to be
// cleaned up in a future update.

// Exporter is an implementation of trace.Exporter that writes spans to io_writer.
type Exporter struct {
}

const (
	// reconnectBackoff is the number of milliseconds to use as a unit for
	// waiting for to attempt to reconnect to the trace sink.  Note that the
	// actual waiting time uses the number of sequential failures to determine
	// how many backoff intervals to wait.
	reconnectBackoff = 100

	// maxBackoff is the maximum period in milliseconds allowed when waiting to
	// attempt to connect to the trace sink service.
	maxBackoff = 5000
)

const (
	// noExporterRequested is the starting state, indicating that no caller has
	// yet tried to set up this exporter type
	noExporterRequested = iota

	// notInitialized indicates that an exporter has been requested, but the
	// target GRPC endpoint has not yet been established
	notInitialized

	// disconnected indicates that there is no active connection to the trace
	// sink service, and no attempt is in progress to establish one
	disconnected

	// awaitingConnection indicates that there is no active connection and that
	// there is an active attempt to establish a connection in progress
	awaitingConnection

	// active is the fully operational state - there is an active exporter and
	// it has a working connection to the trace sink service to send events
	active
)

var (
	// alreadyInitialized indicates that an attempt has been made to change the
	// defined trace sink endpoint.  This is not supported.
	alreadyInitialized = errors.New("trace endpoint already set")

	// mutex controls access to the common entries here
	mutex = sync.Mutex{}

	// state contains the current execution state as defined above
	state = noExporterRequested

	// endpoint contains the host and port for the trace sink service
	endpoint string

	// opts contains the GRPC options to use when connecting to the trace sink
	opts []grpc.DialOption

	// conn holds the active connection to the trace sink, or nil, if none is
	// currently established
	conn *grpc.ClientConn = nil

	// client holds the GRPC client facade for the trace sink service
	client pb.TraceSinkClient

	// connectCount tracks the number of times in a given sequence where we
	// attempt to connect to the trace server and fail.  This is used to
	// determine the backoff interval to use
	connectCount = 0

	// queue contains any trace entries that arrived while there was no active
	// connection to the trace sink.  Entries are placed here, and then sent
	// to the sink service once a connection is established
	queue = common.NewDeferrable(0)
)

// SetEndpoint configures the endpoint and GRPC options, and then starts the
// process to connect to the trace sink service
func SetEndpoint(name string, dialOptions ...grpc.DialOption) error {
	mutex.Lock()
	defer mutex.Unlock()

	if state == noExporterRequested {
		// The production tracer was not chosen.  Ignore this call.
		return nil
	}

	if state != notInitialized {
		return alreadyInitialized
	}

	endpoint = name
	opts = append(opts, dialOptions...)
	state = disconnected

	attemptConnection()

	return nil
}

// NewExporter creates a new trace exporter that sends trace entries to the
// Cloud Chamber trace sink service
func NewExporter() (*Exporter, error) {
	mutex.Lock()
	defer mutex.Unlock()

	if state == noExporterRequested {
		state = notInitialized
	}

	return &Exporter{}, nil
}

// ExportSpan writes a SpanData in json format to io_writer.
func (e *Exporter) ExportSpan(ctx context.Context, data *trace.SpanData) {
	entry := common.ExtractEntry(ctx, data)

	mutex.Lock()
	defer mutex.Unlock()

	if state != active {
		e.deferOne(entry)
	} else {
		e.postOne(ctx, entry)
	}
}

// deferOne puts the trace entry in to the deferred queue.  Note that a failure
// to defer an entry is fatal.
func (e *Exporter) deferOne(data *pbl.Entry) {
	if err := queue.Defer(data); err != nil {
		log.Fatalf("unable to defer an entry, fatal error.  err=%v", err)
	}
}

// postOne attempts to send the trace entry to the trace sink service.  If that
// fails, it will defer the entry and initiate a reconnection attempt.
func (e *Exporter) postOne(ctx context.Context, data *pbl.Entry) {
	_, err := client.Append(ctx, &pb.AppendRequest{Entry: data})
	if err != nil {
		client = nil

		if conn != nil {
			_ = conn.Close()
			conn = nil
		}

		e.deferOne(data)

		state = disconnected
		scheduleReconnect()
	}
}

// attemptConnection tries to connect to the trace sink service.  If it
// succeeds, then all deferred entries are flushed.  If it fails, or if the
// flush operation fails at any point, then a future reconnect is scheduled.
func attemptConnection() {
	var err error

	if state != disconnected {
		return
	}

	connectCount++
	conn, err = grpc.Dial(endpoint, opts...)

	if err == nil {
		client = pb.NewTraceSinkClient(conn)

		connectCount = 0
		state = active

		err = queue.Flush(context.Background(), func(ctx context.Context, entry *pbl.Entry) error {
			_, err = client.Append(ctx, &pb.AppendRequest{Entry: entry})
			return err
		})
	}

	if err != nil {
		state = disconnected
		scheduleReconnect()
		return
	}
}

// scheduleReconnect waits for a backoff interval and restarts the connection
// attempt
func scheduleReconnect() {
	if state != disconnected {
		return
	}

	ms := reconnectBackoff * connectCount
	if ms > maxBackoff {
		ms = maxBackoff
	}

	state = awaitingConnection
	_ = time.AfterFunc(time.Duration(ms)*time.Millisecond, func() {
		mutex.Lock()
		defer mutex.Unlock()

		if state != awaitingConnection {
			log.Fatalf("unexpected change in state while waiting for the backoff interval.  State is %v", state)
		}

		state = disconnected
		attemptConnection()
	})
}
