package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"

	"github.com/Jim3Things/CloudChamber/internal/config"
	"github.com/Jim3Things/CloudChamber/internal/services/stepper_actor"
	"github.com/Jim3Things/CloudChamber/internal/services/tracing_sink"
	"github.com/Jim3Things/CloudChamber/internal/tracing/exporters"
	"github.com/Jim3Things/CloudChamber/internal/tracing/server"
	"github.com/Jim3Things/CloudChamber/internal/tracing/setup"
	"github.com/Jim3Things/CloudChamber/pkg/version"
)

func main() {
	cfgPath := flag.String("config", ".", "path to the configuration file")
	showConfig := flag.Bool("showConfig", false, "display the current configuration settings")
	showVersion := flag.Bool("version", false, "display the current version of the program")
	flag.Parse()

	if *showVersion {
		version.Show()
		os.Exit(0)
	}

	setup.Init(exporters.IoWriter, exporters.Production)

	version.Trace()

	cfg, err := config.ReadGlobalConfig(*cfgPath)
	if err != nil {
		log.Fatalf("failed to process the global configuration: %v", err)
	}

	if *showConfig {
		fmt.Println(config.ToString(cfg))
		os.Exit(0)
	}

	if err = setup.SetFileWriter(cfg.SimSupport.TraceFile); err != nil {
		log.Fatalf("failed to set up the trace logger, err=%v", err)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.SimSupport.EP.Port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// This will fail initially, as the grpc listener is not active yet.  But
	// a failure to connect is a planned condition and normal reconnection
	// handling will get this channel set up.
	if err = setup.SetEndpoint(cfg.SimSupport.EP.String()); err != nil {
		log.Fatalf("failed to set the trace sink endpoint, err=%v", err)
	}

	s := grpc.NewServer(grpc.UnaryInterceptor(server.Interceptor))

	if err = tracing_sink.Register(s); err != nil {
		log.Fatalf(
			"failed to register the tracing sink.  Err: %v", err)
	}

	if err = stepper.Register(s, cfg.SimSupport.GetPolicyType()); err != nil {
		log.Fatalf(
			"failed to register the stepper actor.  default policy: %v, err: %v",
			cfg.SimSupport.GetPolicyType(),
			err)
	}

	if err = s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
