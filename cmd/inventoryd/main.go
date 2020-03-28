package main

import (
    "flag"
    "fmt"
    "log"
    "os"

    "github.com/Jim3Things/CloudChamber/internal/config"
    "github.com/Jim3Things/CloudChamber/internal/tracing/exporters"
    "github.com/Jim3Things/CloudChamber/internal/tracing/setup"
)

func main() {
    setup.Init(exporters.StdOut)

    cfgPath := flag.String("config", ".", "path to the configuration file")
    showConfig := flag.Bool("showConfig", false, "display the current configuration settings")
    flag.Parse()

    cfg, err := config.ReadGlobalConfig(*cfgPath)
    if err != nil {
        log.Fatalf("failed to process the global configuration: %v", err)
    }

    if *showConfig {
        fmt.Println(config.ToString(cfg))
        os.Exit(0)
    }

}
