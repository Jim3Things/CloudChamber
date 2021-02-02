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

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/duration"
	"github.com/stretchr/testify/suite"

	"github.com/Jim3Things/CloudChamber/internal/clients/timestamp"
	"github.com/Jim3Things/CloudChamber/internal/config"
	"github.com/Jim3Things/CloudChamber/internal/tracing/exporters"
	"github.com/Jim3Things/CloudChamber/pkg/protos/admin"
	pb "github.com/Jim3Things/CloudChamber/pkg/protos/services"
	"github.com/Jim3Things/CloudChamber/test/setup"
)

// The constants and global variables here are limited to items that needed by
// these common functions.  Anything specific to a subset of the tests should
// be put into the test file where they are needed.  Also, no specific test
// file should redefine the values set here.

var (
	initServiceDone = false
)

type testSuiteCore struct {
	suite.Suite

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
func (ts *testSuiteCore) alice() string            { return ts.userPath() + "Alice" }
func (ts *testSuiteCore) bob() string              { return ts.userPath() + "Bob" }
func (ts *testSuiteCore) alicePassword() string    { return "test" }
func (ts *testSuiteCore) bobPassword() string      { return "test2" }

func (ts *testSuiteCore) SetupSuite() {
	require := ts.Require()

	// The user URLs that have been added and not deleted during the test run.
	// Note that this does not include any predefined users, such as Admin.
	ts.knownNames = make(map[string]string)

	ts.aliceDef = &admin.UserDefinition{
		Password:          ts.alicePassword(),
		Enabled:           true,
		CanManageAccounts: false,
	}

	ts.bobDef = &admin.UserDefinition{
		Password:          ts.bobPassword(),
		Enabled:           true,
		CanManageAccounts: false,
	}

	ts.utf = exporters.NewExporter(exporters.NewUTForwarder())
	exporters.ConnectToProvider(ts.utf)

	c, err := setup.StartSimSupportServices()
	require.NoError(err)
	ts.cfg = c

	ts.ensureServicesStarted()
}

func (ts *testSuiteCore) SetupTest() {
	_ = ts.utf.Open(ts.T())

	ts.baseURI = fmt.Sprintf("http://localhost:%d", server.port)

	err := timestamp.Reset(context.Background())
	ts.Require().Nilf(err, "Reset failed")

	ctx := context.Background()

	ts.Require().Nil(timestamp.SetPolicy(ctx, pb.StepperPolicy_Manual, &duration.Duration{Seconds: 0}, -1))
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
	defer func() { _ = resp.Body.Close() }()
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
	_, err := ts.getBody(response)

	require.NoError(err, "Failed to read body returned from call to handler for route %q: %v", path, err)
	require.Equal(1, len(response.Cookies()), "Unexpected number of cookies found")
	require.Equal(http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	return response
}

// Log the specified user out of CloudChamber
func (ts *testSuiteCore) doLogout(user string, cookies []*http.Cookie) *http.Response {
	assert := ts.Assert()
	logf := ts.T().Logf

	path := fmt.Sprintf("%s%s?op=logout", ts.userPath(), user)
	logf("[logout from %q (%q)]", user, path)

	request := httptest.NewRequest("PUT", path, nil)
	response := ts.doHTTP(request, cookies)
	_, err := ts.getBody(response)

	assert.Nilf(err, "Failed to read body returned from call to handler for route %v: %v", user, err)
	assert.Equal(http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	return response
}

// ensureServicesStarted handles the various components that can only be set or
// initialized once.
func (ts *testSuiteCore) ensureServicesStarted() {
	if !initServiceDone {
		require := ts.Require()

		// Start the test web service, which all tests will use
		require.NoError(initService(ts.cfg))
		initServiceDone = true
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
	assert.Equal(http.StatusOK, response.StatusCode)

	ts.knownNames[path] = path

	tagString := response.Header.Get("ETag")
	tag, err := parseETag(tagString)
	assert.NoError(err, "Error parsing ETag. tag = %q, err = %v", tagString, err)

	return tag, response.Cookies()
}
