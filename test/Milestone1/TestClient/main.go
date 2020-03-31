package main

import (
    "bufio"
    "flag"
    "fmt"
    "io"
    "log"
    "net/http"
    "os"

    "github.com/Jim3Things/CloudChamber/internal/config"
    "github.com/Jim3Things/CloudChamber/internal/tracing/exporters"
    "github.com/Jim3Things/CloudChamber/internal/tracing/setup"
)

/*
PUT api/Users/admin?op=login  <-- No password or creds at this point
--- Log cookies
GET api/Racks
--- Add login cookie
--- Display output
GET api/Racks/{id} <-- Use the first link returned above, and for all other rack designators below
--- Display output
GET api/Racks/{id}/Blades
GET api/Racks/{id}/Blades/{bladeId} <-- Use the first link returned from the previous call
PUT api/Users/admin?op=logout
 */

// Command line: M1S -config=<global config file>
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

    baseAddress := fmt.Sprintf("http://%s:%d/api", cfg.WebServer.FE.Hostname, cfg.WebServer.FE.Port)
    client := &http.Client{}

    target := fmt.Sprintf("%s/users/admin?op=login", baseAddress)
    resp, err := put(client, target, nil)
    if err != nil {
        panic(err)
    }

    dumpResponse(resp, err)
}

func put(client *http.Client, uri string, body io.Reader) (*http.Response, error) {
    fmt.Printf("PUT to %s\n", uri)
    req, err := http.NewRequest("PUT", uri, body)

    if err != nil {
        return nil, err
    }

    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }

    return resp, nil
}

func dumpResponse(resp *http.Response, err error) {
    fmt.Printf("resp: %v\nerr: %v", resp, err)

    defer resp.Body.Close()
    scanner := bufio.NewScanner(resp.Body)
    scanner.Split(bufio.ScanBytes)
    for scanner.Scan() {
        fmt.Print(scanner.Text())
    }
}
