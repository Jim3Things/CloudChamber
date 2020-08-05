package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/Jim3Things/CloudChamber/internal/config"
	"github.com/Jim3Things/CloudChamber/internal/tracing/exporters"
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

	if err = setup.SetFileWriter(cfg.Inventory.TraceFile); err != nil {
		log.Fatalf("failed to set up the trace logger, err=%v", err)
	}

	if err = setup.SetEndpoint(cfg.SimSupport.EP.Hostname, cfg.SimSupport.EP.Port); err != nil {
		log.Fatalf("failed to set the trace sink endpoint, err=%v", err)
	}

	if *showConfig {
		fmt.Println(config.ToString(cfg))
		os.Exit(0)
	}

}
