package timestamp

import (
	"context"
	"log"
	"net"

	"github.com/golang/protobuf/ptypes/duration"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"

	"github.com/Jim3Things/CloudChamber/simulation/internal/clients/limits"
	"github.com/Jim3Things/CloudChamber/simulation/internal/services/stepper"
	ctrc "github.com/Jim3Things/CloudChamber/simulation/internal/tracing/client"
	"github.com/Jim3Things/CloudChamber/simulation/internal/tracing/exporters"
	strc "github.com/Jim3Things/CloudChamber/simulation/internal/tracing/server"
	pb "github.com/Jim3Things/CloudChamber/simulation/pkg/protos/services"
)

const bufSize = 1024 * 1024

var (
	s   *grpc.Server = nil
	lis *bufconn.Listener
)

type stepperTestClientSuite struct {
	suite.Suite

	lis *bufconn.Listener
	utf *exporters.Exporter

	ep       string
	dialOpts []grpc.DialOption

	s *grpc.Server
}

func (ts *stepperTestClientSuite) SetupSuite() {
	ts.utf = exporters.NewExporter(exporters.NewUTForwarder())
	exporters.ConnectToProvider(ts.utf)

	ts.ep = "bufnet"
	ts.dialOpts = []grpc.DialOption{
		grpc.WithContextDialer(ts.bufDialer),
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(ctrc.Interceptor),
		grpc.WithConnectParams(limits.BackoffSettings),
	}

	ts.ensureStepperStarted()
	_ = InitTimestamp(ts.ep, ts.dialOpts...)
}

func (ts *stepperTestClientSuite) SetupTest() {
	require := ts.Require()

	_ = ts.utf.Open(ts.T())

	ctx := context.Background()

	require.NoError(Reset(ctx))
	_, err := SetPolicy(ctx, pb.StepperPolicy_Manual, &duration.Duration{Seconds: 0}, -1)
	require.NoError(err)
}

func (ts *stepperTestClientSuite) TearDownTest() {
	ts.utf.Close()
}

func (ts *stepperTestClientSuite) bufDialer(_ context.Context, _ string) (net.Conn, error) {
	return ts.lis.Dial()
}

// ensureStepperStarted handles the requirement that there be only one stepper
// service active.  This starts one if one has not yet been started, otherwise
// it reuses the connection information.
func (ts *stepperTestClientSuite) ensureStepperStarted() {
	if s == nil {
		lis = bufconn.Listen(bufSize)
		s = grpc.NewServer(grpc.UnaryInterceptor(strc.Interceptor))

		if err := stepper.Register(context.Background(), s, pb.StepperPolicy_Invalid); err != nil {
			log.Fatalf("Failed to register stepper actor: %v", err)
			return
		}

		go func() {
			if err := s.Serve(lis); err != nil {
				log.Fatalf("Server exited with error: %v", err)
			}
		}()
	}

	ts.lis = lis
	ts.s = s
}
