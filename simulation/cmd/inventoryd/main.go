package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"

	"github.com/Jim3Things/CloudChamber/simulation/internal/clients/limits"
	"github.com/Jim3Things/CloudChamber/simulation/internal/config"
	"github.com/Jim3Things/CloudChamber/simulation/internal/services/inventory"
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
	sink := exporters.NewExporter(exporters.NewSinkForwarder(
		grpc.WithInsecure(),
		grpc.WithConnectParams(limits.BackoffSettings)))
	exporters.ConnectToProvider(iow, sink)

	version.Trace()

	cfg, err := config.ReadGlobalConfig(*cfgPath)
	if err != nil {
		log.Fatalf("failed to process the global configuration: %v", err)
	}

	if *showConfig {
		fmt.Println(cfg)
		os.Exit(0)
	}

	if err = sink.Open(cfg.SimSupport.EP.String()); err != nil {
		log.Fatalf("failed to set the trace sink endpoint, err=%v", err)
	}

	section := cfg.Inventory

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

	if err = inventory.Register(s, cfg); err != nil {
		log.Fatalf("failed to register the inventory service: %v", err)
	}

	if err = s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
