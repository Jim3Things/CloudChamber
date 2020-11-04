package utilities

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/golang/protobuf/ptypes/duration"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"

	"github.com/Jim3Things/CloudChamber/internal/clients/timestamp"
	"github.com/Jim3Things/CloudChamber/internal/clients/trace_sink"
	"github.com/Jim3Things/CloudChamber/internal/config"
	stepper "github.com/Jim3Things/CloudChamber/internal/services/stepper_actor"
	"github.com/Jim3Things/CloudChamber/internal/services/tracing_sink"
	ct "github.com/Jim3Things/CloudChamber/internal/tracing/client"
	st "github.com/Jim3Things/CloudChamber/internal/tracing/server"
	"github.com/Jim3Things/CloudChamber/pkg/protos/services"
)

const (
	bufSize = 1024 * 1024
)

var (
	s   *grpc.Server = nil
	lis *bufconn.Listener
	ep string
	dialOpts []grpc.DialOption

	cfg *config.GlobalConfig
)

func bufDialer(_ context.Context, _ string) (net.Conn, error) {
	return lis.Dial()
}

// StartSimSupportServices does the one-time-only initialization of the
// in-process test configuration for the time and log sink services.  In
// order to do so, it includes the flag parsing and reading the global
// configuration.
func StartSimSupportServices() (*config.GlobalConfig, error) {
	if s == nil {
		ep = "test_channel"
		dialOpts = []grpc.DialOption{
			grpc.WithContextDialer(bufDialer),
			grpc.WithInsecure(),
			grpc.WithUnaryInterceptor(ct.Interceptor),
		}

		configPath := flag.String("config", "./testdata", "path to the configuration file")
		flag.Parse()

		c, err := config.ReadGlobalConfig(*configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to process the global configuration: %v", err)
		}

		cfg = c

		timestamp.InitTimestamp(ep, dialOpts...)
		trace_sink.InitSinkClient(ep, dialOpts...)

		lis = bufconn.Listen(bufSize)
		s = grpc.NewServer(grpc.UnaryInterceptor(st.Interceptor))

		if err = stepper.Register(s, services.StepperPolicy_Invalid); err != nil {
			return nil, fmt.Errorf("failed to register stepper actor: %v", err)
		}

		if _, err = tracing_sink.Register(s, cfg.SimSupport.TraceRetentionLimit); err != nil {
			return nil, fmt.Errorf("failed to register tracing sink: %v", err)
		}

		go func() {
			if err = s.Serve(lis); err != nil {
				log.Fatalf("Server exited with error: %v", err)
			}
		}()

		// Force the initial state to manual so that the setup tracing works
		// correctly (and does not produce spurious trace errors)
		if err = timestamp.SetPolicy(
			context.Background(),
				services.StepperPolicy_Manual,
				&duration.Duration{Seconds: 0}, -1); err != nil {
			return nil, err
		}
	}

	return cfg, nil
}
