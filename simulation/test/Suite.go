package test

import (
	"net/http"
	"strings"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// Suite contains extensions to the testify Suite package.
type Suite struct {
	suite.Suite
}

// Require returns the extended set of assertions, which include both those
// provided by the testify package as well as those provided in this module.
func (ts *Suite) Require() *Require {
	return &Require{
		Assertions: ts.Suite.Require(),
	}
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

// HTTPRHasCookiesExact verifies that the supplied response carries the cookies
// with the expected names, and no other.
func (rq *Require) HTTPRHasCookiesExact(r *http.Response, matches ...string) {
	const prefix = "\n    "

	var names []string

	for _, cookie := range r.Cookies() {
		names = append(names, cookie.Name)
	}

	sought := strings.Join(matches, prefix)
	contents := strings.Join(names, prefix)

	rq.ElementsMatchf(
		matches, names,
		"Cookies did not match: Expected to see \n[%s%s\n], but found \n[%s%s\n]",
		prefix, sought, prefix, contents)
}

func getContentType(r *http.Response) string {
	return strings.ToLower(r.Header.Get("Content-Type"))
}
