// Unit tests for the web service frontend.
//
// Borrows heavily from the gorilla mux test package.
//
package frontend

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Jim3Things/CloudChamber/simulation/pkg/errors"
	pb "github.com/Jim3Things/CloudChamber/simulation/pkg/protos/admin"
)

type UserTestSuite struct {
	testSuiteCore
}

func (ts *UserTestSuite) SetupSuite() {
	ts.testSuiteCore.SetupSuite()
}

func (ts *UserTestSuite) userRead(path string, cookies []*http.Cookie) (*http.Response, *pb.UserPublic) {
	require := ts.Require()

	request := httptest.NewRequest("GET", path, nil)
	request.Header.Set("Content-Type", "application/json")

	response := ts.doHTTP(request, cookies)

	require.HTTPRSuccess(response)

	user := &pb.UserPublic{}
	require.NoError(ts.getJSONBody(response, user))

	return response, user
}

func (ts *UserTestSuite) setPassword(
	name string,
	upd *pb.UserPassword,
	rev int64,
	cookies []*http.Cookie) (*http.Response, int64) {
	require := ts.Require()

	r, err := ts.toJSONReader(upd)
	require.NoError(err)

	path := ts.userPath() + name + "?password"

	request := httptest.NewRequest("PUT", path, r)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("If-Match", formatAsEtag(rev))

	response := ts.doHTTP(request, cookies)

	match, err := parseETag(response.Header.Get("ETag"))
	require.NoError(err)

	return response, match
}

func (ts *UserTestSuite) userUpdate(
	path string,
	upd *pb.UserUpdate,
	rev int64,
	cookies []*http.Cookie) (*http.Response, int64) {
	require := ts.Require()

	r, err := ts.toJSONReader(upd)
	require.NoError(err, "Failed to format UserUpdate, err = %v", err)

	request := httptest.NewRequest("PUT", path, r)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("If-Match", formatAsEtag(rev))

	response := ts.doHTTP(request, cookies)
	require.HTTPRHasCookiesExact(response, sessionCookieName)
	require.HTTPRSuccess(response)

	tag, err := parseETag(response.Header.Get("ETag"))
	require.NoError(err)

	return response, tag
}

// --- Helper functions

// The individual unit tests follow here.  They are grouped by the operation
// they are testing, starting with a simple happy path case, followed by
// repeating sequences (optionally), and then by failure cases.
//
// Each group is demarcated by comment lines.  The group starts with one that
// that has "+++", and ends with one that has "==="

// +++ Login tests

func (ts *UserTestSuite) TestLoginSessionSimple() {
	require := ts.Require()

	// login for the first time, should succeed
	request := httptest.NewRequest("PUT", fmt.Sprintf("%s?op=login", ts.admin()), strings.NewReader(ts.adminPassword()))

	response := ts.doHTTP(request, nil)
	require.HTTPRHasCookiesExact(response, sessionCookieName)
	require.HTTPRSuccess(response)
	require.HTTPRContentTypeEqual("text/plain; charset=utf-8", response)

	body := ts.getBody(response)
	require.Equal("User \"admin\" logged in\n", string(body))

	// ... and logout, which should succeed
	//     (note that this also checks that the username match is case insensitive)
	//
	request = httptest.NewRequest("PUT", fmt.Sprintf("%s?op=logout", strings.ToUpper(ts.admin())), nil)

	response = ts.doHTTP(request, response.Cookies())
	require.HTTPRHasCookiesExact(response, sessionCookieName)
	require.HTTPRSuccess(response)
	require.HTTPRContentTypeEqual("text/plain; charset=utf-8", response)

	body = ts.getBody(response)
	require.Equal("User \"admin\" logged out\n", string(body))
}

func (ts *UserTestSuite) TestLoginSessionRepeat() {
	// login for the first time, should succeed
	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	// ... and logout, which should succeed
	//
	response = ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())

	// login for the second iteration, should succeed
	response = ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), response.Cookies())

	// ... and logout, which should succeed
	//
	response = ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())
}

func (ts *UserTestSuite) TestLoginDupLogins() {
	require := ts.Require()

	// login for the first time, should succeed
	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	// now repeat the attempt to login again, which should fail
	request := httptest.NewRequest("PUT", fmt.Sprintf("%s?op=login", ts.admin()), strings.NewReader(ts.adminPassword()))

	response = ts.doHTTP(request, response.Cookies())
	require.HTTPRHasCookiesExact(response, sessionCookieName)
	require.HTTPRStatusEqual(http.StatusBadRequest, response)
	require.HTTPRContentTypeEqual("text/plain; charset=utf-8", response)

	body := ts.getBody(response)

	require.Equal(
		fmt.Sprintf("%s\n", errors.ErrUserAlreadyLoggedIn.Error()), string(body),
		"Handler returned unexpected response body: %v", string(body))

	// .. and let's just try with another user, which should also fail
	request = httptest.NewRequest("PUT", fmt.Sprintf("%s?op=login", ts.bob()), strings.NewReader("test2"))

	response = ts.doHTTP(request, response.Cookies())
	require.HTTPRHasCookiesExact(response, sessionCookieName)
	require.HTTPRStatusEqual(http.StatusBadRequest, response)
	require.HTTPRContentTypeEqual("text/plain; charset=utf-8", response)

	body = ts.getBody(response)

	require.Equal(
		fmt.Sprintf("%s\n", errors.ErrUserAlreadyLoggedIn.Error()), string(body),
		"Handler returned unexpected response body: %v", string(body))

	// ... and logout, which should succeed
	//
	ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())
}

func (ts *UserTestSuite) TestLoginLogoutDiffAccounts() {
	require := ts.Require()

	// login for the first time, should succeed
	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	// ... next we need a second account that we're sure exists
	_, cookies := ts.ensureAccount(ts.aliceName(), ts.aliceDef, response.Cookies())

	// ... and now try to logout from it, which should not succeed
	//
	request := httptest.NewRequest("PUT", fmt.Sprintf("%s?op=logout", ts.alice()), nil)

	response = ts.doHTTP(request, cookies)
	require.HTTPRHasCookiesExact(response, sessionCookieName)
	require.HTTPRStatusEqual(http.StatusBadRequest, response)
	require.HTTPRContentTypeEqual("text/plain; charset=utf-8", response)

	body := ts.getBody(response)
	require.Equal(
		"CloudChamber: user \"alice\" not logged into this session\n",
		string(body))

	// ... and logout, which should succeed
	//
	response = ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())
}

func (ts *UserTestSuite) TestDoubleLogout() {
	require := ts.Require()

	// login for the first time, should succeed
	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	// ... logout, which should succeed
	//
	response = ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())

	// ... logout again, which should fail
	//
	request := httptest.NewRequest("PUT", fmt.Sprintf("%s?op=logout", ts.admin()), nil)

	response = ts.doHTTP(request, response.Cookies())
	require.HTTPRHasCookiesExact(response, sessionCookieName)
	require.HTTPRStatusEqual(http.StatusBadRequest, response)
	require.HTTPRContentTypeEqual("text/plain; charset=utf-8", response)

	body := ts.getBody(response)
	require.Equal(
		"CloudChamber: user \"admin\" not logged into this session\n",
		string(body))
}

func (ts *UserTestSuite) TestLoginSessionBadPassword() {
	require := ts.Require()

	// login for the first time, should succeed
	request := httptest.NewRequest(
		"PUT",
		fmt.Sprintf("%s?op=login", ts.admin()),
		strings.NewReader(ts.adminPassword()+"rubbish"))
	response := ts.doHTTP(request, nil)
	require.HTTPRHasCookiesExact(response, sessionCookieName)
	require.HTTPRStatusEqual(http.StatusForbidden, response)
	require.HTTPRContentTypeEqual("text/plain; charset=utf-8", response)

	body := ts.getBody(response)
	require.Equal(
		"CloudChamber: authentication failed, invalid user name or password\n",
		string(body))

	// Now just validate that there really isn't an active session here.
	response = ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), response.Cookies())

	ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())
}

func (ts *UserTestSuite) TestLoginSessionNoUser() {
	require := ts.Require()

	// login for the first time, the http call should succeed, but fail the login
	request := httptest.NewRequest(
		"PUT",
		fmt.Sprintf("%s%s?op=login", ts.admin(), "Bogus"),
		strings.NewReader(ts.adminPassword()))
	response := ts.doHTTP(request, nil)
	require.HTTPRHasCookiesExact(response, sessionCookieName)
	require.HTTPRStatusEqual(http.StatusNotFound, response)
	require.HTTPRContentTypeEqual("text/plain; charset=utf-8", response)

	body := ts.getBody(response)
	require.Equal(
		"CloudChamber: authentication failed, invalid user name or password\n",
		string(body))

	// Now just validate that there really isn't an active session here.
	response = ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), response.Cookies())

	ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())
}

// --- Login tests

// +++ User creation tests

func (ts *UserTestSuite) TestCreate() {
	require := ts.Require()

	path := ts.alice() + "2"

	r, err := ts.toJSONReader(ts.aliceDef)
	require.NoError(err, "Failed to format UserDefinition, err = %v", err)

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	request := httptest.NewRequest("POST", path, r)
	request.Header.Set("Content-Type", "application/json")

	response = ts.doHTTP(request, response.Cookies())
	require.HTTPRSuccess(response)
	require.HTTPRHasCookiesExact(response, sessionCookieName)
	require.HTTPRContentTypeEqual("text/plain; charset=utf-8", response)

	body := ts.getBody(response)
	require.Equal(
		"User \"Alice2\" created, enabled: true, rights: ",
		string(body))

	ts.knownNames[path] = path
	ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())
}

func (ts *UserTestSuite) TestCreateDup() {
	require := ts.Require()

	r, err := ts.toJSONReader(ts.aliceDef)
	require.NoError(err, "Failed to format UserDefinition, err = %v", err)

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	_, cookies := ts.ensureAccount(ts.aliceName(), ts.aliceDef, response.Cookies())

	request := httptest.NewRequest("POST", ts.alice(), r)
	request.Header.Set("Content-Type", "application/json")

	response = ts.doHTTP(request, cookies)
	require.HTTPRStatusEqual(http.StatusBadRequest, response)
	require.HTTPRHasCookiesExact(response, sessionCookieName)
	require.HTTPRContentTypeEqual("text/plain; charset=utf-8", response)

	body := ts.getBody(response)
	require.Equal(
		"CloudChamber: user \"Alice\" already exists\n", string(body))

	ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())
}

func (ts *UserTestSuite) TestCreateBadData() {
	require := ts.Require()

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	request := httptest.NewRequest(
		"POST",
		ts.alice()+"2",
		strings.NewReader("{\"enabled\":123,\"manageAccounts\":false, \"password\":\"test\"}"))
	request.Header.Set("Content-Type", "application/json")

	response = ts.doHTTP(request, response.Cookies())
	require.HTTPRStatusEqual(http.StatusBadRequest, response)
	require.HTTPRHasCookiesExact(response, sessionCookieName)
	require.HTTPRContentTypeEqual("text/plain; charset=utf-8", response)

	body := ts.getBody(response)
	require.Equal(
		"json: cannot unmarshal number into Go value of type bool\n",
		string(body))

	ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())
}

func (ts *UserTestSuite) TestCreateNoPrivilege() {
	require := ts.Require()

	r, err := ts.toJSONReader(ts.bobDef)
	require.NoError(err, "Failed to format UserDefinition, err = %v", err)

	response := ts.doLogin(ts.aliceName(), ts.alicePassword(), nil)

	request := httptest.NewRequest("POST", ts.bob(), r)
	request.Header.Set("Content-Type", "application/json")

	response = ts.doHTTP(request, response.Cookies())
	require.HTTPRStatusEqual(http.StatusForbidden, response)
	require.HTTPRHasCookiesExact(response, sessionCookieName)
	require.HTTPRContentTypeEqual("text/plain; charset=utf-8", response)

	body := ts.getBody(response)
	require.Equal("CloudChamber: permission denied\n", string(body))

	ts.doLogout(ts.aliceName(), response.Cookies())
}

func (ts *UserTestSuite) TestCreateNoSession() {
	require := ts.Require()

	path := ts.alice() + "2"

	r, err := ts.toJSONReader(ts.aliceDef)
	require.NoError(err, "Failed to format UserDefinition, err = %v", err)

	request := httptest.NewRequest("POST", path, r)
	request.Header.Set("Content-Type", "application/json")

	response := ts.doHTTP(request, nil)
	require.HTTPRStatusEqual(http.StatusForbidden, response)
	require.HTTPRHasCookiesExact(response, sessionCookieName)
	require.HTTPRContentTypeEqual("text/plain; charset=utf-8", response)

	body := ts.getBody(response)
	require.Equal("CloudChamber: permission denied\n", string(body))
}

// --- User creation tests

// +++ Known users list tests

func (ts *UserTestSuite) TestList() {
	require := ts.Require()

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	request := httptest.NewRequest("GET", ts.userPath(), nil)
	request.Header.Set("Content-Type", "application/json")

	response = ts.doHTTP(request, response.Cookies())
	require.HTTPRSuccess(response)

	users := &pb.UserList{}
	require.NoError(ts.getJSONBody(response, users))

	// Now verify that the list of names matches our expectations.
	// First, form an array of names from the returned structure
	addresses := make([]string, 0, len(users.Users))
	for _, entry := range users.Users {
		require.True(strings.HasSuffix(entry.Uri, entry.Name))
		if strings.EqualFold(entry.Name, ts.adminAccountName()) {
			require.True(entry.Protected)
		}

		addresses = append(addresses, entry.Uri)
	}

	// .. and then verify that all following lines correctly consist of all the expected names
	match := ts.knownNames
	match[ts.admin()] = ts.admin()

	// .. this involves converting the set of keys to an array for matching
	keys := make([]string, 0, len(match))
	for k := range match {
		keys = append(keys, k)
	}

	require.ElementsMatch(keys, addresses)

	ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())
}

func (ts *UserTestSuite) TestListNoPrivilege() {
	require := ts.Require()

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)
	_, cookies := ts.ensureAccount(ts.aliceName(), ts.aliceDef, response.Cookies())
	response = ts.doLogout(ts.randomCase(ts.adminAccountName()), cookies)

	response = ts.doLogin(ts.aliceName(), ts.alicePassword(), response.Cookies())

	request := httptest.NewRequest("GET", ts.userPath(), nil)

	response = ts.doHTTP(request, response.Cookies())
	require.HTTPRStatusEqual(http.StatusForbidden, response)
	require.HTTPRHasCookiesExact(response, sessionCookieName)
	require.HTTPRContentTypeEqual("text/plain; charset=utf-8", response)

	_ = ts.getBody(response)

	ts.doLogout("alice", response.Cookies())
}

// --- Known user list tests

// +++ Get user details tests

func (ts *UserTestSuite) TestRead() {
	require := ts.Require()

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	request := httptest.NewRequest("GET", fmt.Sprintf("%s%s", ts.userPath(), ts.randomCase(ts.adminAccountName())), nil)
	request.Header.Set("Content-Type", "application/json")

	response = ts.doHTTP(request, response.Cookies())
	require.HTTPRSuccess(response)

	user := &pb.UserPublic{}
	require.NoError(ts.getJSONBody(response, user))

	match, err := parseETag(response.Header.Get("ETag"))
	require.NoError(err, "failed to convert the ETag to valid int64")
	require.Less(int64(1), match)

	require.True(user.Enabled)
	require.True(user.Rights.CanManageAccounts)
	require.True(user.NeverDelete)

	ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())
}

func (ts *UserTestSuite) TestReadUnknownUser() {
	require := ts.Require()

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	request := httptest.NewRequest("GET", fmt.Sprintf("%s%s", ts.userPath(), "BadUser"), nil)
	request.Header.Set("Content-Type", "application/json")

	response = ts.doHTTP(request, response.Cookies())
	require.HTTPRHasCookiesExact(response, sessionCookieName)
	require.HTTPRContentTypeEqual("text/plain; charset=utf-8", response)
	require.HTTPRStatusEqual(http.StatusNotFound, response)

	body := ts.getBody(response)
	require.Equal("CloudChamber: user \"baduser\" not found\n", string(body))

	ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())
}

func (ts *UserTestSuite) TestReadNoPrivilege() {
	require := ts.Require()

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)
	_, cookies := ts.ensureAccount(ts.aliceName(), ts.aliceDef, response.Cookies())
	response = ts.doLogout(ts.randomCase(ts.adminAccountName()), cookies)

	response = ts.doLogin(ts.aliceName(), ts.alicePassword(), response.Cookies())

	request := httptest.NewRequest("GET", fmt.Sprintf("%s%s", ts.userPath(), "BadUser"), nil)
	request.Header.Set("Content-Type", "application/json")

	response = ts.doHTTP(request, response.Cookies())
	require.HTTPRHasCookiesExact(response, sessionCookieName)
	require.HTTPRContentTypeEqual("text/plain; charset=utf-8", response)
	require.HTTPRStatusEqual(http.StatusForbidden, response)

	body := ts.getBody(response)
	require.Equal("CloudChamber: permission denied\n", string(body))

	ts.doLogout("alice", response.Cookies())
}

// --- Get user details tests

// +++ User operation (?op=) tests

func (ts *UserTestSuite) TestOperationIllegal() {
	require := ts.Require()

	// Verify a bunch of failure cases. Specifically,
	// - that a naked op fails
	// - that an invalid op fails

	// Case 1, check that naked op fails
	//
	request := httptest.NewRequest("PUT", fmt.Sprintf("%s?op", ts.alice()), nil)
	response := ts.doHTTP(request, nil)

	response = ts.doHTTP(request, response.Cookies())
	require.HTTPRStatusEqual(http.StatusBadRequest, response)
	require.HTTPRHasCookiesExact(response, sessionCookieName)
	require.HTTPRContentTypeEqual("text/plain; charset=utf-8", response)

	body := ts.getBody(response)
	require.Equal(
		"CloudChamber: invalid user operation requested (?op=) for user \"alice\"\n",
		string(body))

	// Case 2, check that an invalid op fails
	//
	request = httptest.NewRequest("PUT", fmt.Sprintf("%s?op=testInvalid", ts.alice()), nil)
	response = ts.doHTTP(request, response.Cookies())
	require.HTTPRStatusEqual(http.StatusBadRequest, response)
	require.HTTPRHasCookiesExact(response, sessionCookieName)
	require.HTTPRContentTypeEqual("text/plain; charset=utf-8", response)

	body = ts.getBody(response)
	require.Equal(
		"CloudChamber: invalid user operation requested (?op=testInvalid) for user \"alice\"\n",
		string(body))
}

// --- User operation (?op=) tests

// +++ Update user tests

func (ts *UserTestSuite) TestUpdateSuccess() {
	require := ts.Require()

	aliceUpd := &pb.UserUpdate{
		Enabled: true,
		Rights: &pb.Rights{
			CanManageAccounts:  true,
			CanStepTime:        false,
			CanModifyWorkloads: false,
			CanModifyInventory: false,
			CanInjectFaults:    false,
			CanPerformRepairs:  false,
		},
	}

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	r, err := ts.toJSONReader(aliceUpd)
	require.NoError(err, "Failed to format UserDefinition, err = %v", err)

	rev, cookies := ts.ensureAccount(ts.aliceName(), ts.aliceDef, response.Cookies())
	request := httptest.NewRequest("PUT", ts.alice(), r)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("If-Match", formatAsEtag(rev))

	response = ts.doHTTP(request, cookies)

	user := &pb.UserPublic{}
	require.NoError(ts.getJSONBody(response, user))

	match, err := parseETag(response.Header.Get("ETag"))
	require.NoError(err, "failed to convert the ETag to valid int64")

	// Note: since ensureAccount() will attempt to re-use an existing account, all we know is
	// that by the time it returns there will be an account, and the returned revision is the
	// revision at the time the account was created, whether then, or earlier. Since for the
	// store, the revision is per-store, and NOT per-key, we cannot assume anything about the
	// exact relationship or "distance" between revisions that are not equal.
	//
	// So a "rev + 1" style test is not appropriate.
	//
	require.Less(rev, match)

	require.True(user.Enabled)
	require.Equal(aliceUpd.Rights, user.Rights)
	require.False(user.NeverDelete)

	require.HTTPRSuccess(response)

	ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())
}

func (ts *UserTestSuite) TestUpdateBadData() {
	require := ts.Require()

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	rev, cookies := ts.ensureAccount(ts.aliceName(), ts.aliceDef, response.Cookies())

	request := httptest.NewRequest(
		"PUT",
		ts.alice(),
		strings.NewReader("{\"enabled\":123,\"manageAccounts\":false, \"password\":\"test\"}"))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("If-Match", formatAsEtag(rev))

	response = ts.doHTTP(request, cookies)
	require.HTTPRStatusEqual(http.StatusBadRequest, response)
	require.HTTPRHasCookiesExact(response, sessionCookieName)
	require.HTTPRContentTypeEqual("text/plain; charset=utf-8", response)

	body := ts.getBody(response)
	require.Equal(
		"json: cannot unmarshal number into Go value of type bool\n",
		string(body))

	ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())
}

func (ts *UserTestSuite) TestUpdateBadMatch() {
	require := ts.Require()

	aliceUpd := &pb.UserUpdate{
		Enabled: true,
		Rights:  &pb.Rights{CanManageAccounts: true},
	}

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	r, err := ts.toJSONReader(aliceUpd)
	require.NoError(err, "Failed to format UserDefinition, err = %v", err)

	rev, cookies := ts.ensureAccount(ts.aliceName(), ts.aliceDef, response.Cookies())
	request := httptest.NewRequest("PUT", ts.alice(), r)

	// Poison the revision
	rev += 10

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("If-Match", formatAsEtag(rev))

	response = ts.doHTTP(request, cookies)
	require.HTTPRStatusEqual(http.StatusConflict, response)
	require.HTTPRHasCookiesExact(response, sessionCookieName)
	require.HTTPRContentTypeEqual("text/plain; charset=utf-8", response)

	body := ts.getBody(response)
	require.Equal(
		"CloudChamber: user \"alice\" has a newer version than expected\n",
		string(body))

	ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())
}

func (ts *UserTestSuite) TestUpdateBadMatchSyntax() {
	require := ts.Require()

	aliceUpd := &pb.UserDefinition{
		Password: ts.alicePassword(),
		Enabled:  true,
		Rights:   &pb.Rights{CanManageAccounts: true},
	}

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	r, err := ts.toJSONReader(aliceUpd)
	require.NoError(err, "Failed to format UserDefinition, err = %v", err)

	_, cookies := ts.ensureAccount(ts.aliceName(), ts.aliceDef, response.Cookies())
	request := httptest.NewRequest("PUT", ts.alice(), r)

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("If-Match", "\"abc\"")

	response = ts.doHTTP(request, cookies)
	require.HTTPRStatusEqual(http.StatusBadRequest, response)
	require.HTTPRHasCookiesExact(response, sessionCookieName)
	require.HTTPRContentTypeEqual("text/plain; charset=utf-8", response)

	body := ts.getBody(response)
	require.Equal(
		"CloudChamber: match value \"\\\"abc\\\"\" is not recognized as a valid integer\n",
		string(body))

	ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())
}

func (ts *UserTestSuite) TestUpdateNoUser() {
	require := ts.Require()

	upd := &pb.UserUpdate{
		Enabled: true,
		Rights:  &pb.Rights{CanManageAccounts: true},
	}

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	r, err := ts.toJSONReader(upd)
	require.NoError(err, "Failed to format UserDefinition, err = %v", err)

	request := httptest.NewRequest("PUT", fmt.Sprintf("%s%s", ts.userPath(), "BadUser"), r)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("If-Match", formatAsEtag(1))

	response = ts.doHTTP(request, response.Cookies())
	require.HTTPRStatusEqual(http.StatusNotFound, response)
	require.HTTPRHasCookiesExact(response, sessionCookieName)
	require.HTTPRContentTypeEqual("text/plain; charset=utf-8", response)

	body := ts.getBody(response)
	require.Equal("CloudChamber: user \"baduser\" not found\n", string(body))

	ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())
}

func (ts *UserTestSuite) TestUpdateNoPrivilege() {
	require := ts.Require()

	aliceUpd := &pb.UserUpdate{
		Enabled: true,
		Rights:  &pb.Rights{CanManageAccounts: true},
	}

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)
	_, cookies := ts.ensureAccount(ts.aliceName(), ts.aliceDef, response.Cookies())
	rev, cookies := ts.ensureAccount(ts.bobName(), ts.bobDef, cookies)
	response = ts.doLogout(ts.randomCase(ts.adminAccountName()), cookies)

	response = ts.doLogin(ts.bobName(), ts.bobPassword(), response.Cookies())

	r, err := ts.toJSONReader(aliceUpd)
	require.NoError(err, "Failed to format UserDefinition, err = %v", err)

	request := httptest.NewRequest("PUT", ts.alice(), r)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("If-Match", formatAsEtag(rev))

	response = ts.doHTTP(request, response.Cookies())
	require.HTTPRStatusEqual(http.StatusForbidden, response)
	require.HTTPRHasCookiesExact(response, sessionCookieName)
	require.HTTPRContentTypeEqual("text/plain; charset=utf-8", response)

	body := ts.getBody(response)
	require.Equal("CloudChamber: permission denied\n", string(body))

	ts.doLogout(ts.bobName(), response.Cookies())
}

func (ts *UserTestSuite) TestUpdateExpandRights() {
	require := ts.Require()

	aliceUpd := &pb.UserUpdate{
		Enabled: true,
		Rights: &pb.Rights{
			CanStepTime: true,
		},
	}

	aliceOriginal := &pb.UserUpdate{
		Enabled: true,
		Rights: &pb.Rights{
			CanManageAccounts:  false,
			CanModifyInventory: true,
		},
	}

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)
	rev, cookies := ts.ensureAccount(ts.aliceName(), ts.aliceDef, response.Cookies())

	response, rev = ts.userUpdate(ts.alice(), aliceOriginal, rev, cookies)
	_ = ts.getBody(response)

	response = ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())

	response = ts.doLogin(ts.aliceName(), ts.alicePassword(), response.Cookies())

	r, err := ts.toJSONReader(aliceUpd)
	require.NoError(err, "Failed to format UserUpdate, err = %v", err)

	request := httptest.NewRequest("PUT", ts.alice(), r)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("If-Match", formatAsEtag(rev))

	response = ts.doHTTP(request, response.Cookies())
	require.HTTPRStatusEqual(http.StatusForbidden, response)

	body := ts.getBody(response)
	require.NoError(err)
	require.Equal("CloudChamber: permission denied\n", string(body))

	// Now verify that the entry has not been changed
	response, user := ts.userRead(ts.alice(), response.Cookies())
	require.Equal(aliceOriginal.Rights, user.Rights)
	require.True(user.Enabled)
	require.False(user.NeverDelete)

	ts.doLogout(ts.aliceName(), response.Cookies())
}

// --- Update user tests

// +++ Delete user tests

func (ts *UserTestSuite) TestDelete() {
	require := ts.Require()

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	_, cookies := ts.ensureAccount(ts.aliceName(), ts.aliceDef, response.Cookies())
	request := httptest.NewRequest("DELETE", ts.alice(), nil)
	request.Header.Set("Content-Type", "application/json")

	response = ts.doHTTP(request, cookies)
	require.HTTPRSuccess(response)
	require.HTTPRHasCookiesExact(response, sessionCookieName)
	require.HTTPRContentTypeEqual("text/plain; charset=utf-8", response)

	body := ts.getBody(response)
	require.Equal("User \"alice\" deleted.", string(body))

	delete(ts.knownNames, ts.aliceName())

	// Now verify the deletion by trying to get the user

	request = httptest.NewRequest("GET", ts.alice(), nil)

	response = ts.doHTTP(request, response.Cookies())
	require.HTTPRStatusEqual(http.StatusNotFound, response)
	require.HTTPRHasCookiesExact(response, sessionCookieName)
	require.HTTPRContentTypeEqual("text/plain; charset=utf-8", response)

	body = ts.getBody(response)
	require.Equal("CloudChamber: user \"alice\" not found\n", string(body))

	ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())
}

func (ts *UserTestSuite) TestDeleteNoUser() {
	require := ts.Require()

	path := ts.alice() + "Bogus"

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	request := httptest.NewRequest("DELETE", path, nil)
	request.Header.Set("Content-Type", "application/json")

	response = ts.doHTTP(request, response.Cookies())
	require.HTTPRStatusEqual(http.StatusNotFound, response)
	require.HTTPRHasCookiesExact(response, sessionCookieName)
	require.HTTPRContentTypeEqual("text/plain; charset=utf-8", response)

	body := ts.getBody(response)
	require.Equal("CloudChamber: user \"alicebogus\" not found\n", string(body))

	ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())
}

func (ts *UserTestSuite) TestDeleteNoPrivilege() {
	require := ts.Require()

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)
	_, cookies := ts.ensureAccount(ts.aliceName(), ts.aliceDef, response.Cookies())
	_, cookies = ts.ensureAccount(ts.bobName(), ts.bobDef, cookies)
	response = ts.doLogout(ts.randomCase(ts.adminAccountName()), cookies)

	response = ts.doLogin(ts.bobName(), ts.bobPassword(), response.Cookies())

	request := httptest.NewRequest("DELETE", ts.alice(), nil)

	response = ts.doHTTP(request, response.Cookies())
	require.HTTPRStatusEqual(http.StatusForbidden, response)
	require.HTTPRHasCookiesExact(response, sessionCookieName)
	require.HTTPRContentTypeEqual("text/plain; charset=utf-8", response)

	body := ts.getBody(response)
	require.Equal("CloudChamber: permission denied\n", string(body))

	ts.doLogout(ts.bobName(), response.Cookies())
}

func (ts *UserTestSuite) TestDeleteProtected() {
	require := ts.Require()

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	request := httptest.NewRequest("DELETE", ts.admin(), nil)

	response = ts.doHTTP(request, response.Cookies())
	require.HTTPRStatusEqual(http.StatusForbidden, response)
	require.HTTPRHasCookiesExact(response, sessionCookieName)
	require.HTTPRContentTypeEqual("text/plain; charset=utf-8", response)

	body := ts.getBody(response)
	require.Equal(
		"CloudChamber: user \"admin\" is protected and cannot be deleted\n",
		string(body))

	ts.doLogout(ts.adminAccountName(), response.Cookies())
}

// --- Delete user tests

// +++ SetPassword user tests

func (ts *UserTestSuite) TestSetPassword() {
	require := ts.Require()

	aliceNewPassword := ts.alicePassword() + "xxx"

	aliceUpd := &pb.UserPassword{
		OldPassword: ts.alicePassword(),
		NewPassword: aliceNewPassword,
		Force:       false,
	}

	aliceRevert := &pb.UserPassword{
		OldPassword: aliceNewPassword,
		NewPassword: ts.alicePassword(),
		Force:       true,
	}

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	rev, cookies := ts.ensureAccount(ts.aliceName(), ts.aliceDef, response.Cookies())
	response, match := ts.setPassword(ts.aliceName(), aliceUpd, rev, cookies)
	require.HTTPRSuccess(response)

	// Note: since ensureAccount() will attempt to re-use an existing account, all we know is
	// that by the time it returns there will be an account, and the returned revision is the
	// revision at the time the account was created, whether then, or earlier. Since for the
	// store, the revision is per-store, and NOT per-key, we cannot assume anything about the
	// exact relationship or "distance" between revisions that are not equal.
	//
	// So a "rev + 1" style test is not appropriate.
	//
	require.Less(rev, match)

	response = ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())

	// Now verify that the password was changed, by trying to log in again
	response = ts.doLogin(ts.aliceName(), aliceNewPassword, response.Cookies())

	// Now set the password back
	response, _ = ts.setPassword(ts.aliceName(), aliceRevert, match, response.Cookies())
	require.HTTPRSuccess(response)

	ts.doLogout(ts.aliceName(), response.Cookies())
}

func (ts *UserTestSuite) TestSetPasswordForce() {
	require := ts.Require()

	aliceNewPassword := ts.alicePassword() + "xxx"

	aliceUpd := &pb.UserPassword{
		OldPassword: "bogus",
		NewPassword: aliceNewPassword,
		Force:       true,
	}

	aliceRevert := &pb.UserPassword{
		OldPassword: aliceNewPassword,
		NewPassword: ts.alicePassword(),
		Force:       true,
	}

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	rev, cookies := ts.ensureAccount(ts.aliceName(), ts.aliceDef, response.Cookies())
	response, match := ts.setPassword(ts.aliceName(), aliceUpd, rev, cookies)
	require.HTTPRSuccess(response)

	// Note: since ensureAccount() will attempt to re-use an existing account, all we know is
	// that by the time it returns there will be an account, and the returned revision is the
	// revision at the time the account was created, whether then, or earlier. Since for the
	// store, the revision is per-store, and NOT per-key, we cannot assume anything about the
	// exact relationship or "distance" between revisions that are not equal.
	//
	// So a "rev + 1" style test is not appropriate.
	//
	require.Less(rev, match)

	response = ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())

	// Now verify that the password was changed, by trying to log in again
	response = ts.doLogin(ts.aliceName(), aliceNewPassword, response.Cookies())

	// Now set the password back
	response, _ = ts.setPassword(ts.aliceName(), aliceRevert, match, response.Cookies())
	require.HTTPRSuccess(response)

	ts.doLogout(ts.aliceName(), response.Cookies())
}

func (ts *UserTestSuite) TestSetPasswordBadPassword() {
	require := ts.Require()

	aliceNewPassword := ts.alicePassword() + "xxx"

	aliceUpd := &pb.UserPassword{
		OldPassword: "bogus",
		NewPassword: aliceNewPassword,
		Force:       false,
	}

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	rev, cookies := ts.ensureAccount(ts.aliceName(), ts.aliceDef, response.Cookies())

	r, err := ts.toJSONReader(aliceUpd)
	require.NoError(err, "Failed to format UserPassword, err = %v", err)

	path := ts.alice() + "?password"

	request := httptest.NewRequest("PUT", path, r)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("If-Match", formatAsEtag(rev))

	response = ts.doHTTP(request, cookies)

	require.HTTPRStatusEqual(http.StatusForbidden, response)

	response = ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())

	// Now verify that the password was not changed, by trying to log in again
	response = ts.doLogin(ts.aliceName(), ts.alicePassword(), response.Cookies())

	ts.doLogout(ts.aliceName(), response.Cookies())
}

func (ts *UserTestSuite) TestSetPasswordNoPrivilege() {
	require := ts.Require()

	adminNewPassword := ts.adminPassword() + "xxx"

	adminUpd := &pb.UserPassword{
		OldPassword: ts.adminPassword(),
		NewPassword: adminNewPassword,
		Force:       false,
	}

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)
	_, cookies := ts.ensureAccount(ts.aliceName(), ts.aliceDef, response.Cookies())
	response = ts.doLogout("Admin", cookies)

	response = ts.doLogin(ts.aliceName(), ts.alicePassword(), response.Cookies())

	r, err := ts.toJSONReader(adminUpd)
	require.NoError(err, "Failed to format UserPassword, err = %v", err)

	path := ts.admin() + "?password"

	request := httptest.NewRequest("PUT", path, r)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("If-Match", formatAsEtag(-1))

	response = ts.doHTTP(request, response.Cookies())
	require.HTTPRStatusEqual(http.StatusForbidden, response)
	require.HTTPRHasCookiesExact(response, sessionCookieName)
	require.HTTPRContentTypeEqual("text/plain; charset=utf-8", response)

	response = ts.doLogout(ts.aliceName(), response.Cookies())

	// Now verify that the password was not changed, by trying to log in again
	response = ts.doLogin("Admin", ts.adminPassword(), response.Cookies())

	ts.doLogout("Admin", response.Cookies())
}

// --- SetPassword user tests

func (ts *UserTestSuite) TestSetRights() {
	require := ts.Require()

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)
	_, cookies := ts.ensureAccount(ts.aliceName(), ts.aliceDef, response.Cookies())

	response, _ = ts.userRead(ts.alice(), cookies)

	rev, err := parseETag(response.Header.Get("ETag"))
	require.NoError(err)

	newRights := &pb.Rights{
		CanManageAccounts:  true,
		CanStepTime:        false,
		CanModifyWorkloads: true,
		CanModifyInventory: false,
		CanInjectFaults:    true,
		CanPerformRepairs:  false,
	}

	upd := &pb.UserUpdate{
		Enabled: true,
		Rights:  newRights,
	}

	response, match := ts.userUpdate(ts.alice(), upd, rev, response.Cookies())

	user := &pb.UserPublic{}
	require.NoError(ts.getJSONBody(response, user))

	require.Less(rev, match)
	require.Equal(newRights, user.Rights)

	rev = match

	newRights = &pb.Rights{
		CanManageAccounts:  false,
		CanStepTime:        true,
		CanModifyWorkloads: false,
		CanModifyInventory: true,
		CanInjectFaults:    false,
		CanPerformRepairs:  true,
	}

	upd.Rights = newRights

	response, match = ts.userUpdate(ts.alice(), upd, rev, response.Cookies())

	user = &pb.UserPublic{}
	require.NoError(ts.getJSONBody(response, user))

	require.Less(rev, match)
	require.Equal(newRights, user.Rights)

	ts.doLogout(ts.adminAccountName(), response.Cookies())
}

func TestUserTestSuite(t *testing.T) {
	suite.Run(t, new(UserTestSuite))
}
