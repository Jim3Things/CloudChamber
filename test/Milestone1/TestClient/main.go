package main

import (
    "bufio"
    "flag"
    "fmt"
    "io"
    "log"
    "net/http"
    "os"
    "strings"

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
    setup.Init(exporters.IoWriter)

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

    // 1: try to login
    target := fmt.Sprintf("%s/users/admin?op=login", baseAddress)
    resp, err := put(client, target, nil, strings.NewReader(cfg.WebServer.SystemAccountPassword))
    if err != nil {
        panic(err)
    }

    dumpResponse(resp, err)

    // 0: get list of known users
    target = fmt.Sprintf("%s/users", baseAddress)
    resp, err = get(client, target, resp.Cookies(), nil)
    if err != nil {
        panic(err)
    }

    dumpResponse(resp, err)

    // 2: get the list of racks
    // TBD

    // last: and now logout
    target = fmt.Sprintf("%s/users/admin?op=logout", baseAddress)
    resp, err = put(client, target, resp.Cookies(), nil)
    if err != nil {
        panic(err)
    }

    dumpResponse(resp, err)
}

func get(client *http.Client, uri string, cookies []*http.Cookie, body io.Reader) (*http.Response, error) {
    fmt.Printf("GET to %s\n", uri)
    req, err := http.NewRequest("GET", uri, body)
    if err != nil {
        return nil, err
    }

    for _, c := range cookies {
        req.AddCookie(c)
    }

    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }

    return resp, nil
}

func put(client *http.Client, uri string, cookies []*http.Cookie, body io.Reader) (*http.Response, error) {
    fmt.Printf("PUT to %s\n", uri)
    req, err := http.NewRequest("PUT", uri, body)

    if err != nil {
        return nil, err
    }

    for _, c := range cookies {
        req.AddCookie(c)
    }

    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }

    return resp, nil
}

func dumpResponse(resp *http.Response, err error) {
    fmt.Printf("resp: %v\nerr: %v\n", resp, err)

    for _, cookie := range resp.Cookies() {
        fmt.Println(cookie.Name)
    }

    defer resp.Body.Close()
    scanner := bufio.NewScanner(resp.Body)
    scanner.Split(bufio.ScanBytes)
    for scanner.Scan() {
        fmt.Print(scanner.Text())
    }

    fmt.Println("")
}
