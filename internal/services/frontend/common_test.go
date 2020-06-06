// This module contains the common support functions used by all frontend
// tests, as well as the single 'TestMain' starting point for the test
// suite.

package frontend

import (
    "bufio"
    "bytes"
    "fmt"
    "io"
    "io/ioutil"
    "log"
    "math/rand"
    "net/http"
    "net/http/httptest"
    "os"
    "strings"
    "testing"

    "github.com/golang/protobuf/jsonpb"
    "github.com/golang/protobuf/proto"

    "github.com/Jim3Things/CloudChamber/internal/config"
    "github.com/Jim3Things/CloudChamber/internal/tracing/exporters"
    "github.com/Jim3Things/CloudChamber/internal/tracing/setup"
)

// The constants and global variables here are limited to items that needed by
// these common functions.  Anything specific to a subset of the tests should
// be put into the test file where they are needed.  Also, no specific test
// file should redefine the values set here.

const (
    adminAccountName = "Admin"
    adminPassword = "AdminPassword"
)

var (
    baseURI     string
    initialized bool
)

// Common test startup method.  This is the _only_ Test* function in this
// file.
func TestMain(m *testing.M) {
    commonSetup()

    os.Exit(m.Run())
}

// Estalish the test environment, including starting a test frontend service
// over a faked http connection.
func commonSetup() {

    if initialized {
        log.Fatalf("Error initializing service for second or subsequent time")
    }

    setup.Init(exporters.UnitTest)

    if err := initService(&config.GlobalConfig{
        Controller: config.ControllerType{},
        Inventory:  config.InventoryType{},
        SimSupport: config.SimSupportType{},
        WebServer: config.WebServerType{
            RootFilePath:          "C:\\CloudChamber",
            SystemAccount:         adminAccountName,
            SystemAccountPassword: adminPassword,
            FE: config.Endpoint{
                Hostname: "localhost",
                Port:     8080,
            },
            BE: config.Endpoint{},
        },
    }); err != nil {
        log.Fatalf("Error initializing service: %v", err)
    }

    baseURI = fmt.Sprintf("http://localhost:%d", server.port)

    initialized = true
}

// +++ Helper functions

// Convert a proto message into a reader with json-formatted contents
func toJsonReader(v proto.Message) (io.Reader, error) {
    var buf bytes.Buffer
    w := bufio.NewWriter(&buf)
    p := jsonpb.Marshaler{}

    if err := p.Marshal(w, v); err != nil {
        return nil, err
    }

    if err := w.Flush(); err != nil {
        return nil, err
    }

    return bufio.NewReader(&buf), nil
}

// Execute an http request/response sequence
func doHTTP(req *http.Request, cookies []*http.Cookie) *http.Response {
    for _, c := range cookies {
        req.AddCookie(c)
    }

    w := httptest.NewRecorder()

    server.handler.ServeHTTP(w, req)

    return w.Result()
}

// Get the body of a response, and close it
func getBody(resp *http.Response) ([]byte, error) {
    defer func() { _ = resp.Body.Close() }()
    return ioutil.ReadAll(resp.Body)
}

// Get the body of a response, unmarshaled into the supplied message structure
func getJsonBody(resp *http.Response, v proto.Message) error {
    defer func() { _ = resp.Body.Close() }()
    return jsonpb.Unmarshal(resp.Body, v)
}

// Take a string and randomly return
//  a) that string unchanged
//  b) that string fully upper-cased
//  c) that string fully lower-cased
//
// This allows validation that case insensitive string handling is.
func randomCase(val string) string {
    switch rand.Intn(3) {
    case 0:
        return val

    case 1:
        return strings.ToUpper(val)

    default:
        return strings.ToLower(val)
    }
}

// --- Helper functions