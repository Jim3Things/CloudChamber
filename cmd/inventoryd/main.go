package main

import (
    "flag"
    "log"

    "github.com/Jim3Things/CloudChamber/internal/config"
    "github.com/Jim3Things/CloudChamber/internal/tracing/exporters"
    "github.com/Jim3Things/CloudChamber/internal/tracing/setup"
)

func main() {
    setup.Init(exporters.StdOut)

    cfgPath := flag.String("config", ".", "path to the configuration file")
    flag.Parse()

    _, err := config.ReadGlobalConfig(*cfgPath)
    if err != nil {
        log.Fatalf("failed to process the global configuration: %v", err)
    }
}
