package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"

	"github.com/Jim3Things/CloudChamber/internal/config"
	"github.com/Jim3Things/CloudChamber/internal/services/monitor"
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

	setup.Init(exporters.IoWriter)
	version.Trace()

	cfg, err := config.ReadGlobalConfig(*cfgPath)
	if err != nil {
		log.Fatalf("failed to process the global configuration: %v", err)
	}

	if err = setup.SetFileWriter(cfg.Controller.TraceFile); err != nil {
		log.Fatalf("failed to set up the trace logger, err=%v", err)
	}

	if err = setup.SetEndpoint(cfg.SimSupport.EP.String()); err != nil {
		log.Fatalf("failed to set the trace sink endpoint, err=%v", err)
	}

	if *showConfig {
		fmt.Println(cfg)
		os.Exit(0)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Controller.EP.Port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer(grpc.UnaryInterceptor(server.Interceptor))

	monitor.Register(s)

	if err = s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
