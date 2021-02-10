package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"

	"github.com/Jim3Things/CloudChamber/simulation/internal/config"
	"github.com/Jim3Things/CloudChamber/simulation/internal/services/stepper"
	"github.com/Jim3Things/CloudChamber/simulation/internal/services/tracing_sink"
	"github.com/Jim3Things/CloudChamber/simulation/internal/tracing/exporters"
	"github.com/Jim3Things/CloudChamber/simulation/internal/tracing/server"
	"github.com/Jim3Things/CloudChamber/simulation/pkg/version"
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

	section := cfg.SimSupport

	var writer io.Writer
	if writer, err = exporters.NameToWriter(section.TraceFile); err != nil {
		log.Fatalf("failed to open name %q, err=%v", section.TraceFile, err)
	}

	if err = iow.Open(writer); err != nil {
		log.Fatalf("failed to set up the trace logger, err=%v", err)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", section.EP.Port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer(grpc.UnaryInterceptor(server.Interceptor))

	if _, err = tracing_sink.Register(s, section.TraceRetentionLimit); err != nil {
		log.Fatalf(
			"failed to register the tracing sink.  Err: %v", err)
	}

	if err = stepper.Register(context.Background(), s, section.GetPolicyType()); err != nil {
		log.Fatalf(
			"failed to register the stepper actor.  default policy: %v, err: %v",
			section.GetPolicyType(),
			err)
	}

	if err = s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
