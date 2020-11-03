package inventory

import (
	"context"
	"flag"
	"log"
	"net"
	"time"

	"github.com/golang/protobuf/ptypes/duration"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"

	"github.com/Jim3Things/CloudChamber/internal/clients/timestamp"
	"github.com/Jim3Things/CloudChamber/internal/clients/trace_sink"
	"github.com/Jim3Things/CloudChamber/internal/common"
	"github.com/Jim3Things/CloudChamber/internal/config"
	stepper "github.com/Jim3Things/CloudChamber/internal/services/stepper_actor"
	"github.com/Jim3Things/CloudChamber/internal/services/tracing_sink"
	"github.com/Jim3Things/CloudChamber/internal/sm"
	"github.com/Jim3Things/CloudChamber/internal/tracing"
	ct "github.com/Jim3Things/CloudChamber/internal/tracing/client"
	"github.com/Jim3Things/CloudChamber/internal/tracing/exporters"
	st "github.com/Jim3Things/CloudChamber/internal/tracing/server"
	pbc "github.com/Jim3Things/CloudChamber/pkg/protos/common"
	pb "github.com/Jim3Things/CloudChamber/pkg/protos/inventory"
	"github.com/Jim3Things/CloudChamber/pkg/protos/services"
)

const (
	bufSize = 1024 * 1024
)

var (
	s   *grpc.Server = nil
	lis *bufconn.Listener
)

type testSuiteCore struct {
	suite.Suite

	utf *exporters.Exporter

	ep       string
	dialOpts []grpc.DialOption

	cfg *config.GlobalConfig
}

func (ts *testSuiteCore) rackName() string { return "rack1" }

func (ts *testSuiteCore) SetupSuite() {
	ts.utf = exporters.NewExporter(exporters.NewUTForwarder())
	exporters.ConnectToProvider(ts.utf)

	ts.ensureServicesStarted()
}

func (ts *testSuiteCore) SetupTest() {
	_ = ts.utf.Open(ts.T())

	err := timestamp.Reset(context.Background())
	ts.Require().Nilf(err, "Reset failed")

	ctx := context.Background()

	ts.Require().Nil(timestamp.SetPolicy(ctx, services.StepperPolicy_Manual, &duration.Duration{Seconds: 0}, -1))
}

func (ts *testSuiteCore) TearDownTest() {
	ts.utf.Close()
}

func (ts *testSuiteCore) createDummyRack(bladeCount int) *pb.ExternalRack {
	rackDef := &pb.ExternalRack{
		Pdu:    &pb.ExternalPdu{},
		Tor:    &pb.ExternalTor{},
		Blades: make(map[int64]*pbc.BladeCapacity),
	}

	for i := 0; i < bladeCount; i++ {
		rackDef.Blades[int64(i)] = &pbc.BladeCapacity{}
	}

	return rackDef
}

func (ts *testSuiteCore) completeWithin(ch <-chan *sm.Response, delay time.Duration) *sm.Response {
	select {
	case res := <-ch:
		return res
	case <-time.After(delay):
		return nil
	}
}

func (ts *testSuiteCore) execute(
	ctx context.Context,
	msg sm.Envelope,
	action func(ctx2 context.Context, envelope sm.Envelope)) {
	go func(tick int64, msg sm.Envelope) {
		c2, span := tracing.StartSpan(context.Background(),
			tracing.WithName("Executing simulated inventory operation"),
			tracing.WithNewRoot(),
			tracing.WithLink(msg.GetSpanContext(), msg.GetLinkID()))

		c2 = common.ContextWithTick(c2, tick)

		action(c2, msg)

		span.End()
	}(common.TickFromContext(ctx), msg)
}

func (ts *testSuiteCore) bufDialer(_ context.Context, _ string) (net.Conn, error) {
	return lis.Dial()
}

// ensureServicesStarted handles the various components that can only be set or
// initialized once.  This includes flag parsing and support service startup.
func (ts *testSuiteCore) ensureServicesStarted() {
	if s == nil {
		require := ts.Require()

		ts.ep = "bufnet"
		ts.dialOpts = []grpc.DialOption{
			grpc.WithContextDialer(ts.bufDialer),
			grpc.WithInsecure(),
			grpc.WithUnaryInterceptor(ct.Interceptor),
		}

		configPath := flag.String("config", "./testdata", "path to the configuration file")
		flag.Parse()

		cfg, err := config.ReadGlobalConfig(*configPath)
		if err != nil {
			log.Fatalf("failed to process the global configuration: %v", err)
		}

		ts.cfg = cfg

		timestamp.InitTimestamp(ts.ep, ts.dialOpts...)
		trace_sink.InitSinkClient(ts.ep, ts.dialOpts...)

		lis = bufconn.Listen(bufSize)
		s = grpc.NewServer(grpc.UnaryInterceptor(st.Interceptor))

		if err = stepper.Register(s, services.StepperPolicy_Invalid); err != nil {
			log.Fatalf("Failed to register stepper actor: %v", err)
			return
		}

		if _, err = tracing_sink.Register(s, cfg.SimSupport.TraceRetentionLimit); err != nil {
			log.Fatalf("Failed to register tracing sink: %v", err)
		}

		go func() {
			if err = s.Serve(lis); err != nil {
				log.Fatalf("Server exited with error: %v", err)
			}
		}()

		// Force the initial state to manual so that the setup tracing works
		// correctly (and does not produce spurious trace errors)
		require.NoError(
			timestamp.SetPolicy(
				context.Background(),
				services.StepperPolicy_Manual,
				&duration.Duration{Seconds: 0}, -1))
	}
}
