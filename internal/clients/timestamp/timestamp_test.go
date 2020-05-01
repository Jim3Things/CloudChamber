package clients

import (
	"context"
	"log"
	"net"
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes/duration"
	"github.com/stretchr/testify/assert"

	"github.com/Jim3Things/CloudChamber/internal/services/stepper_actor"
	ctrc "github.com/Jim3Things/CloudChamber/internal/tracing/client"
	"github.com/Jim3Things/CloudChamber/internal/tracing/exporters"
	"github.com/Jim3Things/CloudChamber/internal/tracing/exporters/unit_test"
	strc "github.com/Jim3Things/CloudChamber/internal/tracing/server"
	"github.com/Jim3Things/CloudChamber/internal/tracing/setup"
	pb "github.com/Jim3Things/CloudChamber/pkg/protos/Stepper"
	ct "github.com/Jim3Things/CloudChamber/pkg/protos/common"

	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

var lis *bufconn.Listener

func init() {
	setup.Init(exporters.UnitTest)

	lis = bufconn.Listen(bufSize)
	s := grpc.NewServer(grpc.UnaryInterceptor(strc.Interceptor))
	stepper.Register(s)

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
	err := Reset()
	assert.Nilf(t, err, "Reset failed")

	unit_test.SetTesting(t)

	InitTimestamp("bufnet",
		grpc.WithContextDialer(bufDialer),
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(ctrc.Interceptor))
}

func TestNow(t *testing.T) {
	commonSetup(t)
	assert.Nil(t, SetPolicy(pb.StepperPolicy_Manual, &duration.Duration{Seconds: 0}))

	now, err := Now()
	assert.Nil(t, err)
	assert.Equal(t, int64(0), now.Ticks)

	now, err = Now()
	assert.Nil(t, err)
	assert.Equal(t, int64(0), now.Ticks)

	assert.Nil(t, Advance())
	now, err = Now()
	assert.Nil(t, err)
	assert.Equal(t, int64(1), now.Ticks)

	now, err = Now()
	assert.Nil(t, err)
	assert.Equal(t, int64(1), now.Ticks)
}

func TestTimestamp_After(t *testing.T) {
	commonSetup(t)
	assert.Nil(t, SetPolicy(pb.StepperPolicy_Manual, &duration.Duration{Seconds: 0}))

	now, err := Now()
	assert.Nil(t, err)
	assert.Equal(t, int64(0), now.Ticks)
	afterHit := false

	go func(deadline int64) {
		ch, err := After(&ct.Timestamp{Ticks: deadline})
		assert.Nil(t, err)

		data := <-ch
		afterHit = true

		assert.Nil(t, data.err)
		assert.GreaterOrEqual(t, deadline, data.time.Ticks)
	}(3)

	assert.Nil(t, Advance())
	time.Sleep(time.Duration(2) * time.Second)
	assert.False(t, afterHit)

	assert.Nil(t, Advance())
	time.Sleep(time.Duration(2) * time.Second)
	assert.False(t, afterHit)

	assert.Nil(t, Advance())
	time.Sleep(time.Duration(2) * time.Second)
	assert.True(t, afterHit)
}
