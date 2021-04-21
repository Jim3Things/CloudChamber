package frontend

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/Jim3Things/CloudChamber/simulation/internal/clients/timestamp"
	"github.com/Jim3Things/CloudChamber/simulation/internal/config"
	"github.com/Jim3Things/CloudChamber/simulation/internal/tracing/exporters"
	"github.com/Jim3Things/CloudChamber/simulation/pkg/protos/admin"
	pb "github.com/Jim3Things/CloudChamber/simulation/pkg/protos/services"
	"github.com/Jim3Things/CloudChamber/simulation/test"
	"github.com/Jim3Things/CloudChamber/simulation/test/setup"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/duration"
)

// The constants and global variables here are limited to items that needed by
// these common functions.  Anything specific to a subset of the tests should
// be put into the test file where they are needed.  Also, no specific test
// file should redefine the values set here.

var (
	initServiceDone = false
)

type testSuiteCore struct {
	test.Suite

	baseURI string

	utf *exporters.Exporter

	cfg *config.GlobalConfig

	aliceDef *admin.UserDefinition
	bobDef   *admin.UserDefinition

	knownNames map[string]string
}

func (ts *testSuiteCore) adminAccountName() string { return "Admin" }
func (ts *testSuiteCore) adminPassword() string    { return "AdminPassword" }
func (ts *testSuiteCore) userPath() string         { return ts.baseURI + "/api/users/" }
func (ts *testSuiteCore) admin() string            { return ts.userPath() + ts.adminAccountName() }
func (ts *testSuiteCore) aliceName() string        { return "Alice" }
func (ts *testSuiteCore) alice() string            { return ts.userPath() + ts.aliceName() }
func (ts *testSuiteCore) bobName() string          { return "Bob" }
func (ts *testSuiteCore) bob() string              { return ts.userPath() + ts.bobName() }
func (ts *testSuiteCore) alicePassword() string    { return "test" }
func (ts *testSuiteCore) bobPassword() string      { return "test2" }

func (ts *testSuiteCore) SetupSuite() {
	require := ts.Require()

	// The user URLs that have been added and not deleted during the test run.
	// Note that this does not include any predefined users, such as Admin.
	ts.knownNames = make(map[string]string)

	ts.aliceDef = &admin.UserDefinition{
		Password: ts.alicePassword(),
		Enabled:  true,
		Rights:   &admin.Rights{CanManageAccounts: false},
	}

	ts.bobDef = &admin.UserDefinition{
		Password: ts.bobPassword(),
		Enabled:  true,
		Rights:   &admin.Rights{CanManageAccounts: false},
	}

	ts.utf = exporters.NewExporter(exporters.NewUTForwarder())
	exporters.ConnectToProvider(ts.utf)

	c, err := setup.StartSimSupportServices()
	require.NoError(err)
	ts.cfg = c

	ts.ensureServicesStarted()
}

func (ts *testSuiteCore) SetupTest() {
	require := ts.Require()

	_ = ts.utf.Open(ts.T())

	ts.baseURI = fmt.Sprintf("http://localhost:%d", server.port)

	ctx := context.Background()

	require.NoError(timestamp.Reset(ctx))
	_, err := timestamp.SetPolicy(ctx, pb.StepperPolicy_Manual, &duration.Duration{Seconds: 0}, -1)
	require.NoError(err)
}

func (ts *testSuiteCore) TearDownTest() {
	ts.utf.Close()
}

// Convert a proto message into a reader with json-formatted contents
func (ts *testSuiteCore) toJSONReader(v proto.Message) (io.Reader, error) {
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
func (ts *testSuiteCore) doHTTP(req *http.Request, cookies []*http.Cookie) *http.Response {
	for _, c := range cookies {
		req.AddCookie(c)
	}

	w := httptest.NewRecorder()

	server.handler.ServeHTTP(w, req)

	return w.Result()
}

// Get the body of a response, and close it
func (ts *testSuiteCore) getBody(resp *http.Response) ([]byte, error) {
	defer func() { _ = resp.Body.Close() }()
	return ioutil.ReadAll(resp.Body)
}

// Get the body of a response, unmarshalled into the supplied message structure
func (ts *testSuiteCore) getJSONBody(resp *http.Response, v proto.Message) error {
	require := ts.Require()

	defer func() { _ = resp.Body.Close() }()

	require.HTTPRContentTypeJson(resp)

	return jsonpb.Unmarshal(resp.Body, v)
}

// Take a string and randomly return
//  a) that string unchanged
//  b) that string fully upper-cased
//  c) that string fully lower-cased
//
// This allows validation that case insensitive string handling is functioning
// correctly.
func (ts *testSuiteCore) randomCase(val string) string {
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
func (ts *testSuiteCore) doLogin(user string, password string, cookies []*http.Cookie) *http.Response {
	require := ts.Require()
	logf := ts.T().Logf

	path := fmt.Sprintf("%s%s?op=login", ts.userPath(), user)
	logf("[login as %q (%q)]", user, path)

	request := httptest.NewRequest("PUT", path, strings.NewReader(password))
	response := ts.doHTTP(request, cookies)

	require.HTTPRSuccess(response)
	require.HTTPRHasCookie(sessionCookieName, response)

	_, err := ts.getBody(response)

	require.NoError(err, "Failed to read body returned from call to handler for route %q: %v", path, err)

	return response
}

// Log the specified user out of CloudChamber
func (ts *testSuiteCore) doLogout(user string, cookies []*http.Cookie) *http.Response {
	require := ts.Require()
	logf := ts.T().Logf

	path := fmt.Sprintf("%s%s?op=logout", ts.userPath(), user)
	logf("[logout from %q (%q)]", user, path)

	request := httptest.NewRequest("PUT", path, nil)
	response := ts.doHTTP(request, cookies)

	require.HTTPRSuccess(response)
	require.HTTPRHasCookie(sessionCookieName, response)

	_, err := ts.getBody(response)

	require.NoError(err, "Failed to read body returned from call to handler for route %v: %v", user, err)

	return response
}

// ensureServicesStarted handles the various components that can only be set or
// initialized once.
func (ts *testSuiteCore) ensureServicesStarted() {
	if !initServiceDone {
		require := ts.Require()

		ctx := context.Background()

		_ = ts.utf.Open(ts.T())
		defer ts.utf.Close()

		// Start the test web service, which all tests will use
		require.NoError(initService(ts.cfg))
		initServiceDone = true

		// Load the standard inventory into the store which all tests will use
		require.NoError(
			dbInventory.inventory.UpdateInventoryDefinition(
				ctx,
				ts.cfg.Inventory.InventoryDefinition))

		// Need to reload the actual inventory after loading the store.
		require.NoError(dbInventory.LoadInventoryActual(ctx, true))
	}
}

// Ensure that the specified account exists.  This function first checks if it
// is already known, returning that account's current revision if it is.  If it
// is not, then the account is created using the supplied definition, again
// returning the revision number.
//
// Note that this is mostly used by unit tests in order to support running any
// unit test in isolation from the overall flow.
func (ts *testSuiteCore) ensureAccount(
	user string,
	u *admin.UserDefinition,
	cookies []*http.Cookie) (int64, []*http.Cookie) {
	assert := ts.Assert()
	logf := ts.T().Logf

	path := ts.userPath() + user

	req := httptest.NewRequest("GET", path, nil)
	req.Header.Set("Content-Type", "application/json")
	response := ts.doHTTP(req, cookies)
	_ = response.Body.Close()

	// If we found the user, just return the existing revision and cookies
	if response.StatusCode == http.StatusOK {
		logf("Found existing user %q.", user)

		var rev int64
		tagString := response.Header.Get("ETag")
		rev, err := parseETag(tagString)
		assert.NoError(err, "Error parsing ETag. tag = %q, err = %v", tagString, err)

		return rev, response.Cookies()
	}

	// Didn't find the user, create a new incarnation of it.
	logf("Did not find user %q.  Creating it from scratch.", user)

	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)

	p := jsonpb.Marshaler{}
	err := p.Marshal(w, u)
	assert.NoError(err)
	_ = w.Flush()
	r := bufio.NewReader(&buf)

	req = httptest.NewRequest("POST", path, r)
	req.Header.Set("Content-Type", "application/json")

	response = ts.doHTTP(req, response.Cookies())
	assert.HTTPRSuccess(response)

	ts.knownNames[path] = path

	tagString := response.Header.Get("ETag")
	tag, err := parseETag(tagString)
	assert.NoError(err, "Error parsing ETag. tag = %q, err = %v", tagString, err)

	return tag, response.Cookies()
}
