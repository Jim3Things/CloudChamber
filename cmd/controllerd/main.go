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
)

var (
	// declare a bunch of variables whose values given here will be replaced by some
	// build-time defined values. If for some reason those build-time replacements
	// do not take place then the -version flag will dump these values as-is.
	//
	version         = "development"
	buildDate       = "buildDate"
	buildBranch     = "branch"
	buildBranchHash = "branch-hash"
	buildBranchDate = "branch-date"
)

func main() {
	cfgPath := flag.String("config", ".", "path to the configuration file")
	showConfig := flag.Bool("showConfig", false, "display the current configuration settings")
	showVersion := flag.Bool("version", false, "display the current version of the program")
	flag.Parse()

	if *showVersion {
		fmt.Printf("Version: %v\nBuildDate: %v\nBranch: %v (%v \\ %v)\n", version, buildDate, buildBranch, buildBranchDate, buildBranchHash)
		os.Exit(0)
	}

	setup.Init(exporters.StdOut)

	cfg, err := config.ReadGlobalConfig(*cfgPath)
	if err != nil {
		log.Fatalf("failed to process the global configuration: %v", err)
	}

	if *showConfig {
		fmt.Println(config.ToString(cfg))
		os.Exit(0)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Controller.EP.Port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer(grpc.UnaryInterceptor(server.Interceptor))

	monitor.Register(s)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
