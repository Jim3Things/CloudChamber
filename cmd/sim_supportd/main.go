package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	"github.com/Jim3Things/CloudChamber/internal/services/stepper"
	"github.com/Jim3Things/CloudChamber/internal/tracing/exporters"
	"github.com/Jim3Things/CloudChamber/internal/tracing/server"
	"github.com/Jim3Things/CloudChamber/internal/tracing/setup"
)

const (
	defaultPort = 8080
)

func main() {
	port := flag.Int(
		"port",
		defaultPort,
		"port to listen on for simulation utility functions")

	flag.Parse()

	setup.Init(exporters.StdOut)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer(grpc.UnaryInterceptor(server.Interceptor))

	stepper.Register(s)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
