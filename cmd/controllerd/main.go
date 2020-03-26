package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	"github.com/Jim3Things/CloudChamber/internal/config"
	"github.com/Jim3Things/CloudChamber/internal/services/monitor"
	"github.com/Jim3Things/CloudChamber/internal/tracing/exporters"
	"github.com/Jim3Things/CloudChamber/internal/tracing/server"
	"github.com/Jim3Things/CloudChamber/internal/tracing/setup"
)

func main() {
	setup.Init(exporters.StdOut)

	cfgPath := flag.String("config", ".", "path to the configuration file")
	flag.Parse()

	cfg, err := config.ReadGlobalConfig(*cfgPath)
	if err != nil {
		panic(err)
	}

	fmt.Print(config.ToString(cfg))

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Controller.Port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer(grpc.UnaryInterceptor(server.Interceptor))

	monitor.Register(s)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
