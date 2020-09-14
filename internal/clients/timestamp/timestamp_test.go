package clients

import (
    "context"
    "log"
    "net"
    "testing"
    "time"

    "github.com/golang/protobuf/ptypes/duration"
    "github.com/stretchr/testify/assert"

    "github.com/Jim3Things/CloudChamber/internal/common"
    "github.com/Jim3Things/CloudChamber/internal/services/stepper_actor"
    ctrc "github.com/Jim3Things/CloudChamber/internal/tracing/client"
    "github.com/Jim3Things/CloudChamber/internal/tracing/exporters"
    strc "github.com/Jim3Things/CloudChamber/internal/tracing/server"
    ct "github.com/Jim3Things/CloudChamber/pkg/protos/common"
    pb "github.com/Jim3Things/CloudChamber/pkg/protos/services"

    "google.golang.org/grpc"
    "google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

var (
	lis *bufconn.Listener

	utf *exporters.Exporter
)

func init() {
	utf = exporters.NewExporter(exporters.NewUTForwarder())
	exporters.ConnectToProvider(utf)

	lis = bufconn.Listen(bufSize)
	s := grpc.NewServer(grpc.UnaryInterceptor(strc.Interceptor))

	if err := stepper.Register(s, pb.StepperPolicy_Invalid); err != nil {
		log.Fatalf("Failed to register stepper actor: %v", err)
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

func commonSetup(t *testing.T) {
	InitTimestamp("bufnet",
		grpc.WithContextDialer(bufDialer),
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(ctrc.Interceptor))

	err := Reset(context.Background())
	assert.Nilf(t, err, "Reset failed")
}

func TestNow(t *testing.T) {
	_ = utf.Open(t)
	defer utf.Close()

	ctx := context.Background()

	commonSetup(t)
	assert.Nil(t, SetPolicy(ctx, pb.StepperPolicy_Manual, &duration.Duration{Seconds: 0}, -1))

	now, err := Now(ctx)
	assert.Nil(t, err)
	assert.Equal(t, int64(0), now.Ticks)

	now, err = Now(ctx)
	assert.Nil(t, err)
	assert.Equal(t, int64(0), now.Ticks)

	assert.Nil(t, Advance(ctx))
	now, err = Now(ctx)
	assert.Nil(t, err)
	assert.Equal(t, int64(1), now.Ticks)

	now, err = Now(ctx)
	assert.Nil(t, err)
	assert.Equal(t, int64(1), now.Ticks)
}

func TestTimestamp_After(t *testing.T) {
	_ = utf.Open(t)
	defer utf.Close()

	ctx := context.Background()

	commonSetup(t)
	assert.Nil(t, SetPolicy(ctx, pb.StepperPolicy_Manual, &duration.Duration{Seconds: 0}, -1))

	now, err := Now(ctx)
	assert.Nil(t, err)
	assert.Equal(t, int64(0), now.Ticks)

	ch := make(chan bool)

	go func(deadline int64, res chan<- bool) {
		data := <-After(ctx, &ct.Timestamp{Ticks: deadline})

		assert.Nil(t, data.Err)
		assert.GreaterOrEqual(t, deadline, data.Time.Ticks)
		res <- true
	}(3, ch)

	assert.Nil(t, Advance(ctx))
	assert.True(t, common.DoNotCompleteWithin(ch, time.Duration(2)*time.Second))

	assert.Nil(t, Advance(ctx))
	assert.True(t, common.DoNotCompleteWithin(ch, time.Duration(2)*time.Second))

	assert.Nil(t, Advance(ctx))
	assert.True(t, common.CompleteWithin(ch, time.Duration(2)*time.Second))
}
