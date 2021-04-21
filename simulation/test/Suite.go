package test

import (
	"net/http"
	"strings"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// Suite contains extensions to the testify Suite package.
type Suite struct {
	suite.Suite
}

// Assert returns the extended set of assertions, which include both those
// provided by the testify package as well as those provided in this module.
func (ts *Suite) Assert() *Assert {
	return &Assert{
		Assertions: ts.Suite.Assert(),
	}
}

// Require returns the extended set of assertions, which include both those
// provided by the testify package as well as those provided in this module.
func (ts *Suite) Require() *Require {
	return &Require{
		Assertions: ts.Suite.Require(),
	}
}

type Assert struct {
	*assert.Assertions
}

// HTTPRContentTypeEqual verifies that the content type attribute in the
// supplied response object matches the expected string.
func (a *Assert) HTTPRContentTypeEqual(match string, r *http.Response) bool {
	return a.Equal(match, getContentType(r))
}

// HTTPRContentTypeNotEqual verifies that the content type attribute in the
// supplied response object does not match the expected string.
func (a *Assert) HTTPRContentTypeNotEqual(match string, r *http.Response) bool {
	return a.NotEqual(match, getContentType(r))
}

// HTTPRContentTypeJson verifies that the content type attribute in the
// supplied response object matches the json specifier ('application/json')
func (a *Assert) HTTPRContentTypeJson(r *http.Response) bool {
	return a.Equal("application/json", getContentType(r))
}

// HTTPRContentTypeNotJson verifies that the content type attribute in the
// supplied response object does not match the json specifier ('application/json')
func (a *Assert) HTTPRContentTypeNotJson(r *http.Response) bool {
	return a.NotEqual("application/json", getContentType(r))
}

// HTTPRSuccess verifies that the HTTP status code in the response is OK (200).
func (a *Assert) HTTPRSuccess(r *http.Response) bool {
	return a.Equal(http.StatusOK, r.StatusCode)
}

// HTTPRStatusEqual verifies that the HTTP status code in the response matches
// the expected status.
func (a *Assert) HTTPRStatusEqual(status int, r *http.Response) bool {
	return a.Equal(status, r.StatusCode)
}

// HTTPRHasCookie verifies that the supplied response carries a cookie with the
// expected name.
func (a *Assert) HTTPRHasCookie(name string, r *http.Response) bool {
	cookies := r.Cookies()
	var names []string

	for _, cookie := range cookies {
		names = append(names, cookie.Name)
		if cookie.Name == name {
			return true
		}
	}

	contents := strings.Join(names, "\n    ")
	return a.Failf("Failed to find the expected cookie",
		"Expected to find a cookie named %q, cookies names present are \n[\n    %s\n]",
		name,
		contents)
}

type Require struct {
	*require.Assertions
}

// HTTPRContentTypeEqual verifies that the content type attribute in the
// supplied response object matches the expected string.
func (rq *Require) HTTPRContentTypeEqual(match string, r *http.Response) {
	rq.Equal(match, getContentType(r))
}

// HTTPRContentTypeNotEqual verifies that the content type attribute in the
// supplied response object does not match the expected string.
func (rq *Require) HTTPRContentTypeNotEqual(match string, r *http.Response) {
	rq.NotEqual(match, getContentType(r))
}

// HTTPRContentTypeJson verifies that the content type attribute in the
// supplied response object matches the json specifier ('application/json')
func (rq *Require) HTTPRContentTypeJson(r *http.Response) {
	rq.Equal("application/json", getContentType(r))
}

// HTTPRContentTypeNotJson verifies that the content type attribute in the
// supplied response object does not match the json specifier ('application/json')
func (rq *Require) HTTPRContentTypeNotJson(r *http.Response) {
	rq.NotEqual("application/json", getContentType(r))
}

// HTTPRSuccess verifies that the HTTP status code in the response is OK (200).
func (rq *Require) HTTPRSuccess(r *http.Response) {
	rq.Equal(http.StatusOK, r.StatusCode)
}

// HTTPRStatusEqual verifies that the HTTP status code in the response matches
// the expected status.
func (rq *Require) HTTPRStatusEqual(status int, r *http.Response) {
	rq.Equal(status, r.StatusCode)
}

// HTTPRHasCookie verifies that the supplied response carries a cookie with the
// expected name.
func (rq *Require) HTTPRHasCookie(name string, r *http.Response) {
	cookies := r.Cookies()
	var names []string

	for _, cookie := range cookies {
		names = append(names, cookie.Name)
		if cookie.Name == name {
			return
		}
	}

	contents := strings.Join(names, "\n    ")
	rq.Failf("Failed to find the expected cookie",
		"Expected to find a cookie named %q, cookies names present are \n[\n    %s\n]",
		name,
		contents)
}

func getContentType(r *http.Response) string {
	return strings.ToLower(r.Header.Get("Content-Type"))
}
