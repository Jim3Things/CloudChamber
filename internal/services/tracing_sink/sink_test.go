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

	"github.com/Jim3Things/CloudChamber/internal/common"
	ct "github.com/Jim3Things/CloudChamber/internal/tracing/client"
	"github.com/Jim3Things/CloudChamber/internal/tracing/exporters"
	st "github.com/Jim3Things/CloudChamber/internal/tracing/server"
	log2 "github.com/Jim3Things/CloudChamber/pkg/protos/log"
	pb "github.com/Jim3Things/CloudChamber/pkg/protos/services"
)

const bufSize = 1024 * 1024

var (
	lis    *bufconn.Listener
	client pb.TraceSinkClient

	utf *exporters.Exporter
)

func init() {
	utf = exporters.NewExporter(exporters.NewUTForwarder())
	exporters.ConnectToProvider(utf)

	lis = bufconn.Listen(bufSize)
	s := grpc.NewServer(grpc.UnaryInterceptor(st.Interceptor))
	if _, err := Register(s, 100); err != nil {
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

func createNonZeroRand64() int64 {
	return rand.Int63n(0x7FFF_FFFF_FFFF_FFFE) + 1
}

func createEntryBase(name string, spanID int64, parentID int64, traceID [2]int64) *log2.Entry {
	return &log2.Entry{
		Name:           name,
		SpanID:         fmt.Sprintf("%016x", spanID),
		ParentID:       fmt.Sprintf("%016x", parentID),
		TraceID:        fmt.Sprintf("%016x%016x", traceID[0], traceID[1]),
		Infrastructure: false,
		Status:         "ok",
		StackTrace:     fmt.Sprintf("xxxx%d", rand.Int63()),
		Event:          []*log2.Event{},
	}

}
func createEntry(events int) *log2.Entry {
	entry := createEntryBase(
		fmt.Sprintf("test-%d", rand.Int63()),
		createNonZeroRand64(),
		createNonZeroRand64(),
		[2]int64{createNonZeroRand64(), createNonZeroRand64()})

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

func createRootEntry(events int) *log2.Entry {
	entry := createEntryBase(
		fmt.Sprintf("test-%d", rand.Int63()),
		createNonZeroRand64(),
		0,
		[2]int64{createNonZeroRand64(), createNonZeroRand64()})

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
		assertEventMatches(t, expected.Event[i], actual.Event[i])
	}
}

func assertEventMatches(t *testing.T, expected *log2.Event, actual *log2.Event) {
	assert.Equal(t, expected.Name, actual.Name)
	assert.Equal(t, expected.Tick, actual.Tick)
	assert.Equal(t, expected.StackTrace, actual.StackTrace)
	assert.Equal(t, expected.Severity, actual.Severity)
	assert.Equal(t, expected.Text, actual.Text)
	assert.Equal(t, len(expected.Impacted), len(actual.Impacted))
}

func TestGetAfterNoEntries(t *testing.T) {
	_ = utf.Open(t)
	defer utf.Close()

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
	_ = utf.Open(t)
	defer utf.Close()

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
	_ = utf.Open(t)
	defer utf.Close()

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
	_ = utf.Open(t)
	defer utf.Close()

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
		"rpc error: code = Unknown desc = the field \"MaxEntries\" "+
			"must be greater than or equal to 10.  It is 1, which is invalid",
		err.Error(),
		"unexpected error: %v",
		err)
	assert.Nilf(t, res, "go an unexpected result: %v", res)
}

func TestGetAfterMultipleEntries(t *testing.T) {
	_ = utf.Open(t)
	defer utf.Close()

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

func TestGetAfterMissed(t *testing.T) {
	_ = utf.Open(t)
	defer utf.Close()

	ctx, conn := commonSetup(t)
	defer func() { _ = conn.Close() }()

	entries := createEntries(105, 1)
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
	assert.Equal(t, int64(14), res.LastId)
	assert.True(t, res.Missed)

	require.Equal(t, 10, len(res.Entries))

	for i := 0; i < len(res.Entries); i++ {
		id := res.Entries[i].Id
		assertEntryMatches(t, entries[id], res.Entries[i].Entry)
	}
}

func TestGetAfterRepeatedNewAppends(t *testing.T) {
	_ = utf.Open(t)
	defer utf.Close()

	ctx, conn := commonSetup(t)
	defer func() { _ = conn.Close() }()

	entry := createRootEntry(1)
	_, err := client.Append(ctx, &pb.AppendRequest{Entry: entry})
	require.Nilf(t, err, "unexpected error: %v", err)

	entries := createEntries(4, 1)
	for i := 0; i < len(entries); i++ {
		_, err = client.Append(ctx, &pb.AppendRequest{Entry: entries[i]})
		require.Nilf(t, err, "unexpected error: %v", err)
	}

	entries = append([]*log2.Entry{entry}, entries...)

	policy, err := client.GetPolicy(ctx, &pb.GetPolicyRequest{})
	require.Nilf(t, err, "unexpected error: %v", err)

	assert.Equal(t, int64(-1), policy.FirstId)
	assert.Equal(t, int64(100), policy.MaxEntriesHeld)

	res, err := client.GetAfter(ctx, &pb.GetAfterRequest{
		Id:         policy.FirstId,
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

	entries = createEntries(5, 1)
	for i := 0; i < len(entries); i++ {
		_, err = client.Append(ctx, &pb.AppendRequest{Entry: entries[i]})
		require.Nilf(t, err, "unexpected error: %v", err)
	}

	res, err = client.GetAfter(ctx, &pb.GetAfterRequest{
		Id:         res.LastId,
		MaxEntries: 10,
		Wait:       false,
	})

	assert.Nilf(t, err, "unexpected error: %v", err)

	require.NotNil(t, res)
	assert.Equal(t, int64(9), res.LastId)
	assert.False(t, res.Missed)

	require.Equal(t, 5, len(res.Entries))

	for i := 0; i < len(res.Entries); i++ {
		assert.Equal(t, int64(i+5), res.Entries[i].Id)
		assertEntryMatches(t, entries[i], res.Entries[i].Entry)
	}
}

func TestGetPolicyOverflow(t *testing.T) {
	_ = utf.Open(t)
	defer utf.Close()

	ctx, conn := commonSetup(t)
	defer func() { _ = conn.Close() }()

	policy, err := client.GetPolicy(ctx, &pb.GetPolicyRequest{})
	require.Nilf(t, err, "unexpected error: %v", err)

	assert.Equal(t, int64(-1), policy.FirstId)
	assert.Equal(t, int64(100), policy.MaxEntriesHeld)

	entries := createEntries(105, 1)
	for i := 0; i < len(entries); i++ {
		_, err = client.Append(ctx, &pb.AppendRequest{Entry: entries[i]})
		require.Nilf(t, err, "unexpected error: %v", err)
	}

	policy, err = client.GetPolicy(ctx, &pb.GetPolicyRequest{})
	require.Nilf(t, err, "unexpected error: %v", err)

	assert.Equal(t, int64(4), policy.FirstId)
	assert.Equal(t, int64(100), policy.MaxEntriesHeld)

	res, err := client.GetAfter(ctx, &pb.GetAfterRequest{
		Id:         policy.FirstId,
		MaxEntries: 10,
		Wait:       false,
	})

	assert.Nilf(t, err, "unexpected error: %v", err)

	require.NotNil(t, res)
	assert.Equal(t, int64(14), res.LastId)
	assert.False(t, res.Missed)

	require.Equal(t, 10, len(res.Entries))

	for i := 0; i < len(res.Entries); i++ {
		id := int64(i) + policy.FirstId + 1
		assert.Equal(t, id, res.Entries[i].Id)
		assertEntryMatches(t, entries[id], res.Entries[i].Entry)
	}
}

func TestGetAfterWaitImmediateOneEntry(t *testing.T) {
	_ = utf.Open(t)
	defer utf.Close()

	ctx, conn := commonSetup(t)
	defer func() { _ = conn.Close() }()

	entry := createEntry(0)

	_, err := client.Append(ctx, &pb.AppendRequest{Entry: entry})
	require.Nilf(t, err, "unexpected error: %v", err)

	res, err := client.GetAfter(ctx, &pb.GetAfterRequest{
		Id:         -1,
		MaxEntries: 10,
		Wait:       true,
	})

	assert.Nilf(t, err, "unexpected error: %v", err)

	require.NotNil(t, res)
	assert.Equal(t, int64(0), res.LastId)
	assert.False(t, res.Missed)

	require.Equal(t, 1, len(res.Entries))

	assert.Equal(t, int64(0), res.Entries[0].Id)
	assertEntryMatches(t, entry, res.Entries[0].Entry)
}

func TestGetAfterWaitOneEntry(t *testing.T) {
	_ = utf.Open(t)
	defer utf.Close()

	ctx, conn := commonSetup(t)
	defer func() { _ = conn.Close() }()

	entry := createEntry(0)

	ch := make(chan bool)

	go func(ch chan<- bool) {
		res, err := client.GetAfter(ctx, &pb.GetAfterRequest{
			Id:         -1,
			MaxEntries: 10,
			Wait:       true,
		})

		assert.Nilf(t, err, "unexpected error: %v", err)

		require.NotNil(t, res)
		assert.Equal(t, int64(0), res.LastId)
		assert.False(t, res.Missed)

		require.Equal(t, 1, len(res.Entries))

		assert.Equal(t, int64(0), res.Entries[0].Id)
		assertEntryMatches(t, entry, res.Entries[0].Entry)
		ch <- true
	}(ch)

	assert.True(t, common.DoNotCompleteWithin(ch, time.Duration(100)*time.Millisecond))

	_, err := client.Append(ctx, &pb.AppendRequest{Entry: entry})
	require.Nilf(t, err, "unexpected error: %v", err)

	assert.True(t, common.CompleteWithin(ch, time.Duration(1)*time.Second))
}

func TestGetAfterWaitAfterInfraEntry(t *testing.T) {
	_ = utf.Open(t)
	defer utf.Close()

	ctx, conn := commonSetup(t)
	defer func() { _ = conn.Close() }()

	entry := createEntry(0)
	entry.Infrastructure = true

	_, err := client.Append(ctx, &pb.AppendRequest{Entry: entry})
	require.Nilf(t, err, "unexpected error: %v", err)

	entry = createEntry(0)

	ch := make(chan bool)

	go func(ch chan<- bool) {
		res, err2 := client.GetAfter(ctx, &pb.GetAfterRequest{
			Id:         -1,
			MaxEntries: 10,
			Wait:       true,
		})

		assert.Nilf(t, err2, "unexpected error: %v", err2)

		require.NotNil(t, res)
		assert.Equal(t, int64(1), res.LastId)
		assert.False(t, res.Missed)

		require.Equal(t, 2, len(res.Entries))

		assert.Equal(t, int64(0), res.Entries[0].Id)
		assertEntryMatches(t, entry, res.Entries[1].Entry)

		ch <- true
	}(ch)

	assert.True(t, common.DoNotCompleteWithin(ch, time.Duration(100)*time.Millisecond))

	_, err = client.Append(ctx, &pb.AppendRequest{Entry: entry})
	require.Nilf(t, err, "unexpected error: %v", err)

	assert.True(t, common.CompleteWithin(ch, time.Duration(1)*time.Second))
}
