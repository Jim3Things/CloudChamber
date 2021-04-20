package test

import (
	"net/http"
	"strings"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type SuiteX struct {
	suite.Suite
}

func (ts *SuiteX) Assert() *Assert {
	return &Assert{
		Assertions: ts.Suite.Assert(),
	}
}

func (ts *SuiteX) Require() *Require {
	return &Require{
		Assertions: ts.Suite.Require(),
	}
}

type Assert struct {
	*assert.Assertions
}

func (a *Assert) HTTPContentTypeEqual(match string, r *http.Response) bool {
	return a.Equal(match, getContentType(r))
}

func (a *Assert) HTTPContentTypeNotEqual(match string, r *http.Response) bool {
	return a.NotEqual(match, getContentType(r))
}

func (a *Assert) HTTPContentTypeJson(r *http.Response) bool {
	return a.Equal("application/json", getContentType(r))
}

func (a *Assert) HTTPContentTypeNotJson(r *http.Response) bool {
	return a.NotEqual("application/json", getContentType(r))
}

func (a *Assert) HTTPStatusOK(r *http.Response) bool {
	return a.Equal(http.StatusOK, r.StatusCode)
}

func (a *Assert) HTTPStatusEqual(status int, r *http.Response) bool {
	return a.Equal(status, r.StatusCode)
}

type Require struct {
	*require.Assertions
}

func (rq *Require) HTTPContentTypeEqual(match string, r *http.Response) {
	rq.Equal(match, getContentType(r))
}

func (rq *Require) HTTPContentTypeNotEqual(match string, r *http.Response) {
	rq.NotEqual(match, getContentType(r))
}

func (rq *Require) HTTPContentTypeJson(r *http.Response) {
	rq.Equal("application/json", getContentType(r))
}

func (rq *Require) HTTPContentTypeNotJson(r *http.Response) {
	rq.NotEqual("application/json", getContentType(r))
}

func (rq *Require) HTTPStatusOK(r *http.Response) {
	rq.Equal(http.StatusOK, r.StatusCode)
}

func (rq *Require) HTTPStatusEqual(status int, r *http.Response) {
	rq.Equal(status, r.StatusCode)
}

func (rq *Require) Bogus() {
	rq.Fail("We hit Bogus!")
}

func getContentType(r *http.Response) string {
	return strings.ToLower(r.Header.Get("Content-Type"))
}
