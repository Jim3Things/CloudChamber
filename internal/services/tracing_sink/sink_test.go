package tracing_sink

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/test/bufconn"

	ct "github.com/Jim3Things/CloudChamber/internal/tracing/client"
	"github.com/Jim3Things/CloudChamber/internal/tracing/exporters"
	"github.com/Jim3Things/CloudChamber/internal/tracing/exporters/unit_test"
	st "github.com/Jim3Things/CloudChamber/internal/tracing/server"
	"github.com/Jim3Things/CloudChamber/internal/tracing/setup"
	log2 "github.com/Jim3Things/CloudChamber/pkg/protos/log"
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
		grpc.WithUnaryInterceptor(ct.Interceptor))
	assert.Nilf(t, err, "Failed to dial bufnet: %v", err)

	md := metadata.Pairs(
		"timestamp", time.Now().Format(time.StampNano),
		"client-id", "web-api-client-us-east-1",
		"user-id", "some-test-user-id",
	)
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	client = pb.NewTraceSinkClient(conn)

	_, err = client.Reset(ctx, &pb.ResetRequest{})
	require.Nilf(t, err, "unexpected error: %v", err)

	return ctx, conn
}

func createEntry(events int) *log2.Entry {
	tag := rand.Int()

	entry := &log2.Entry{
		Name:       fmt.Sprintf("test-%d", tag),
		SpanID:     fmt.Sprintf("0000%d", tag),
		ParentID:   fmt.Sprintf("0001%d", tag),
		Status:     "ok",
		StackTrace: fmt.Sprintf("xxxx%d", tag),
		Event:      []*log2.Event{},
	}

	for i := 0; i < events; i++ {
		entry.Event = append(entry.Event, &log2.Event{
			Tick:       0,
			Severity:   0,
			Name:       fmt.Sprintf("Event-%d", i),
			Text:       "xxxx",
			StackTrace: fmt.Sprintf("xxxx%d", i),
			Impacted:   nil,
		})
	}

	return entry
}

func createEntries(entries int, eventsPer int) []*log2.Entry {
	var items = make([]*log2.Entry, 0)

	for i := 0; i < entries; i++ {
		items = append(items, createEntry(eventsPer))
	}

	return items
}

func assertEntryMatches(t *testing.T, expected *log2.Entry, actual *log2.Entry) {
	require.NotNil(t, actual)
	assert.Equal(t, expected.Name, actual.Name)
	assert.Equal(t, expected.SpanID, actual.SpanID)
	assert.Equal(t, expected.ParentID, actual.ParentID)
	assert.Equal(t, expected.StackTrace, actual.StackTrace)
	assert.Equal(t, expected.Status, actual.Status)
	assert.Equal(t, len(expected.Event), len(actual.Event))

	for i := 0; i < len(expected.Event); i++ {
		assert.Equal(t, expected.Event[i].Name, actual.Event[i].Name)
		assert.Equal(t, expected.Event[i].Tick, actual.Event[i].Tick)
		assert.Equal(t, expected.Event[i].StackTrace, actual.Event[i].StackTrace)
		assert.Equal(t, expected.Event[i].Severity, actual.Event[i].Severity)
	}
}

func TestGetAfterNoEntries(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	ctx, conn := commonSetup(t)
	defer func() { _ = conn.Close() }()

	res, err := client.GetAfter(ctx, &pb.GetAfterRequest{
		Id:         -1,
		MaxEntries: 10,
		Wait:       false,
	})
	require.Nilf(t, err, "unexpected error: %v", err)

	require.NotNil(t, res)
	assert.Equal(t, int64(-1), res.LastId)
	assert.Equal(t, 0, len(res.Entries))
	assert.False(t, res.Missed)
}

func TestGetAfterOneEntry(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	ctx, conn := commonSetup(t)
	defer func() { _ = conn.Close() }()

	entry := createEntry(0)

	_, err := client.Append(ctx, &pb.AppendRequest{Entry: entry})
	require.Nilf(t, err, "unexpected error: %v", err)

	res, err := client.GetAfter(ctx, &pb.GetAfterRequest{
		Id:         -1,
		MaxEntries: 10,
		Wait:       false,
	})

	assert.Nilf(t, err, "unexpected error: %v", err)

	require.NotNil(t, res)
	assert.Equal(t, int64(0), res.LastId)
	assert.False(t, res.Missed)

	require.Equal(t, 1, len(res.Entries))

	assert.Equal(t, int64(0), res.Entries[0].Id)
	assertEntryMatches(t, entry, res.Entries[0].Entry)
}

func TestGetAfterEndOfEntries(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	ctx, conn := commonSetup(t)
	defer func() { _ = conn.Close() }()

	entry := createEntry(0)

	_, err := client.Append(ctx, &pb.AppendRequest{Entry: entry})
	require.Nilf(t, err, "unexpected error: %v", err)

	res, err := client.GetAfter(ctx, &pb.GetAfterRequest{
		Id:         -1,
		MaxEntries: 10,
		Wait:       false,
	})

	assert.Nilf(t, err, "unexpected error: %v", err)

	require.NotNil(t, res)
	assert.Equal(t, int64(0), res.LastId)
	assert.False(t, res.Missed)

	require.Equal(t, 1, len(res.Entries))

	item := res.Entries[0].Entry
	assert.Equal(t, int64(0), res.Entries[0].Id)

	require.NotNil(t, item)
	assert.Equal(t, entry.Name, item.Name)

	startID := res.LastId

	res, err = client.GetAfter(ctx, &pb.GetAfterRequest{
		Id:         startID,
		MaxEntries: 10,
		Wait:       false,
	})

	assert.Nilf(t, err, "unexpected error: %v", err)

	require.NotNil(t, res)
	assert.Equal(t, startID, res.LastId)
	assert.False(t, res.Missed)

	require.Equal(t, 0, len(res.Entries))
}

func TestGetAfterMaxEntriesTooSmall(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	ctx, conn := commonSetup(t)
	defer func() { _ = conn.Close() }()

	entry := createEntry(0)

	_, err := client.Append(ctx, &pb.AppendRequest{Entry: entry})
	require.Nilf(t, err, "unexpected error: %v", err)

	res, err := client.GetAfter(ctx, &pb.GetAfterRequest{
		Id:         -1,
		MaxEntries: 1,
		Wait:       false,
	})

	require.NotNil(t, err)
	assert.Equal(
		t,
		"rpc error: code = Unknown desc = the field \"MaxEntries\" " +
			"must be greater than or equal to 10.  It is 1, which is invalid",
		err.Error(),
		"unexpected error: %v",
		err)
	assert.Nilf(t, res, "go an unexpected result: %v", res)
}

func TestGetAfterMultipleEntries(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	ctx, conn := commonSetup(t)
	defer func() { _ = conn.Close() }()

	entries := createEntries(5, 1)
	for i := 0; i < len(entries); i++ {
		_, err := client.Append(ctx, &pb.AppendRequest{Entry: entries[i]})
		require.Nilf(t, err, "unexpected error: %v", err)
	}

	res, err := client.GetAfter(ctx, &pb.GetAfterRequest{
		Id:         -1,
		MaxEntries: 10,
		Wait:       false,
	})

	assert.Nilf(t, err, "unexpected error: %v", err)

	require.NotNil(t, res)
	assert.Equal(t, int64(4), res.LastId)
	assert.False(t, res.Missed)

	require.Equal(t, 5, len(res.Entries))

	for i := 0; i < len(res.Entries); i++ {
		assert.Equal(t, int64(i), res.Entries[i].Id)
		assertEntryMatches(t, entries[i], res.Entries[i].Entry)
	}
}

// Repeated append/getafter sequence

// getpolicy
// append *105, getpolicy
// getpolicy to drive getafter after some appends


