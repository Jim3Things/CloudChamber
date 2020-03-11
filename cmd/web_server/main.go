package main

import (
    "flag"
    "log"


    "github.com/Jim3Things/CloudChamber/internal/services/frontend"
)

const (
    defaultPort = 8080
    defaultRootPath = "C:\\Chamber"
)

func main() {
    port := flag.Int(
        "port",
        defaultPort,
        "port used by the web service")

    rootPath := flag.String(
        "path",
        defaultRootPath,
        "directory path holding the cloud chamber web service data")


    flag.Parse()

    if err := frontend.StartService(port, rootPath); err != nil {
        log.Fatalf("Error running service: %v", err)
    }

}
