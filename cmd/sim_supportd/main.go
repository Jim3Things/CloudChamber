package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"

	"github.com/Jim3Things/CloudChamber/internal/config"
	"github.com/Jim3Things/CloudChamber/internal/services/stepper_actor"
	"github.com/Jim3Things/CloudChamber/internal/services/tracing_sink"
	"github.com/Jim3Things/CloudChamber/internal/tracing/exporters"
	"github.com/Jim3Things/CloudChamber/internal/tracing/server"
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
	iow := exporters.NewExporter(exporters.NewIOWForwarder())
	exporters.ConnectToProvider(iow)

	version.Trace()

	cfg, err := config.ReadGlobalConfig(*cfgPath)
	if err != nil {
		log.Fatalf("failed to process the global configuration: %v", err)
	}

	if *showConfig {
		fmt.Println(cfg)
		os.Exit(0)
	}

	var writer io.Writer
	if writer, err = exporters.NameToWriter(cfg.SimSupport.TraceFile); err != nil {
		log.Fatalf("failed to open name %q, err=%v", cfg.WebServer.TraceFile, err)
	}

	if err = iow.Open(writer); err != nil {
		log.Fatalf("failed to set up the trace logger, err=%v", err)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.SimSupport.EP.Port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
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
