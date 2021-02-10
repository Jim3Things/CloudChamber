package exporters

import (
	"context"
	"reflect"
	"time"

	"google.golang.org/grpc"

	"github.com/Jim3Things/CloudChamber/simulation/pkg/errors"
	pbl "github.com/Jim3Things/CloudChamber/simulation/pkg/protos/log"
	pb "github.com/Jim3Things/CloudChamber/simulation/pkg/protos/services"
)

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

// SinkForwarder implements a forwarder to the Cloud Chamber trace sink service
type SinkForwarder struct {
	endpoint string

	opts   []grpc.DialOption
	conn   *grpc.ClientConn
	client pb.TraceSinkClient

	connectCount int
	active       bool
}

// NewSinkForwarder creates a new, closed, forwarder instance.
func NewSinkForwarder(dialOptions ...grpc.DialOption) *SinkForwarder {
	return &SinkForwarder{
		endpoint:     "",
		opts:         dialOptions,
		conn:         nil,
		client:       nil,
		connectCount: 0,
		active:       false,
	}
}

// Open supplies the endpoint information to be able to connect to the trace
// sink service.
func (sf *SinkForwarder) Open(attrs interface{}) error {
	if attrs == nil {
		return errors.ErrOpenAttrsNil
	}

	ep, ok := attrs.(string)

	if !ok {
		return errors.ErrInvalidOpenAttrsType{
			Expected: reflect.TypeOf(sf.endpoint).String(),
			Actual:   reflect.TypeOf(ep).String(),
		}
	}

	sf.endpoint = ep
	return nil
}

// Close detaches from the trace sink service
func (sf *SinkForwarder) Close() {
	sf.closeConnection()
}

// Forward emits the log entries to the trace sink service.  It automatically
// handle reconnecting when there is an error while forwarding an entry.
func (sf *SinkForwarder) Forward(entry *pbl.Entry, early bool) error {
	for done := false; !done; {
		sf.waitForConnection()

		if err := sf.post(entry, early); err != nil {
			sf.closeConnection()
		} else {
			done = true
		}
	}

	return nil
}

// closeConnection closes any active connection, and cleans up from a failed
// client operation
func (sf *SinkForwarder) closeConnection() {
	sf.client = nil
	if sf.conn != nil {
		_ = sf.conn.Close()
		sf.conn = nil
	}

	sf.active = false
}

// waitForConnection establishes a connection to the trace sink service,
// including handling retries with increasing backoff intervals.
func (sf *SinkForwarder) waitForConnection() {
	var err error

	for !sf.active {
		if sf.conn, err = grpc.Dial(sf.endpoint, sf.opts...); err != nil {
			sf.conn = nil

			ms := reconnectBackoff * sf.connectCount
			if ms > maxBackoff {
				ms = maxBackoff
			}

			time.Sleep(time.Duration(ms) * time.Millisecond)
			sf.connectCount++
		} else {
			sf.client = pb.NewTraceSinkClient(sf.conn)

			// connect
			sf.active = true
		}
	}
}

// post performs the GRPC call to send the log entry to the trace sink service.
func (sf *SinkForwarder) post(entry *pbl.Entry, _ bool) error {
	_, err := sf.client.Append(context.Background(), &pb.AppendRequest{Entry: entry})

	if err == nil {
		sf.connectCount = 0
	}

	return err
}
