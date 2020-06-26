// This module contains the common support functions used by all frontend
// tests, as well as the single 'TestMain' starting point for the test
// suite.

package frontend

import (
    "bufio"
    "bytes"
    "context"
    "fmt"
    "io"
    "io/ioutil"
    "log"
    "math/rand"
    "net"
    "net/http"
    "net/http/httptest"
    "os"
    "strings"
    "testing"

    "github.com/golang/protobuf/jsonpb"
    "github.com/golang/protobuf/proto"
    "github.com/stretchr/testify/assert"
    "google.golang.org/grpc"
    "google.golang.org/grpc/test/bufconn"

    ts "github.com/Jim3Things/CloudChamber/internal/clients/timestamp"
    "github.com/Jim3Things/CloudChamber/internal/config"
    stepper "github.com/Jim3Things/CloudChamber/internal/services/stepper_actor"
    ctrc "github.com/Jim3Things/CloudChamber/internal/tracing/client"
    "github.com/Jim3Things/CloudChamber/internal/tracing/exporters"
    strc "github.com/Jim3Things/CloudChamber/internal/tracing/server"
    "github.com/Jim3Things/CloudChamber/internal/tracing/setup"
    pb "github.com/Jim3Things/CloudChamber/pkg/protos/Stepper"
)

// The constants and global variables here are limited to items that needed by
// these common functions.  Anything specific to a subset of the tests should
// be put into the test file where they are needed.  Also, no specific test
// file should redefine the values set here.

const (
    adminAccountName = "Admin"
    adminPassword = "AdminPassword"
    bufSize = 1024 * 1024
)

var (
    baseURI     string
    initialized bool
    lis *bufconn.Listener
)


// Common test startup method.  This is the _only_ Test* function in this
// file.
func TestMain(m *testing.M) {
    commonSetup()

    os.Exit(m.Run())
}

// Establish the test environment, including starting a test frontend service
// over a faked http connection.
func commonSetup() {
    if initialized {
        log.Fatalf("Error initializing service for second or subsequent time")
    }

    setup.Init(exporters.UnitTest)

    // Set up the internal test deployment of the stepper service, in order to
    // support the stepper frontend unit tests
    lis = bufconn.Listen(bufSize)
    s := grpc.NewServer(grpc.UnaryInterceptor(strc.Interceptor))

    if err := stepper.Register(s, pb.StepperPolicy_Manual); err != nil {
        log.Fatalf("Failed to register stepper actor: %v", err)
    }

    go func() {
        if err := s.Serve(lis); err != nil {
            log.Fatalf("Server exited with error: %v", err)
        }
    }()

    ts.InitTimestamp("bufnet",
        grpc.WithContextDialer(bufDialer),
        grpc.WithInsecure(),
        grpc.WithUnaryInterceptor(ctrc.Interceptor))

    // Finally, start the test web service, which all tests will use
    if err := initService(&config.GlobalConfig{
        Controller: config.ControllerType{},
        Inventory:  config.InventoryType{},
        SimSupport: config.SimSupportType{
            EP: config.Endpoint{
                Hostname: "localhost",
                Port:     8083,
            },
            StepperPolicy: "manual",
        },
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

// Simple dialer for the test in-memory message grpc transport
func bufDialer(_ context.Context, _ string) (net.Conn, error) {
    return lis.Dial()
}

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

// Log the specified user into CloudChamber
func doLogin(t *testing.T, user string, password string, cookies []*http.Cookie) *http.Response {
    path := fmt.Sprintf("%s%s%s?op=login", baseURI, userURI, user)
    t.Logf("[login as %q (%q)]", user, path)

    request := httptest.NewRequest("PUT", path, strings.NewReader(password))
    response := doHTTP(request, cookies)
    _, err := getBody(response)

    assert.Nilf(t, err, "Failed to read body returned from call to handler for route %q: %v", path, err)
    assert.Equal(t, 1, len(response.Cookies()), "Unexpected number of cookies found")
    assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

    return response
}

// Log the specified user out of CloudChamber
func doLogout(t *testing.T, user string, cookies []*http.Cookie) *http.Response {
    path := fmt.Sprintf("%s%s%s?op=logout", baseURI, userURI, user)
    t.Logf("[logout from %q (%q)]", user, path)

    request := httptest.NewRequest("PUT", path, nil)
    response := doHTTP(request, cookies)
    _, err := getBody(response)

    assert.Nilf(t, err, "Failed to read body returned from call to handler for route %v: %v", alice, err)
    assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

    return response
}