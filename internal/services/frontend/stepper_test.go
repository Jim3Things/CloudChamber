package frontend

import (
    "fmt"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/stretchr/testify/assert"

    "github.com/Jim3Things/CloudChamber/internal/tracing/exporters/unit_test"
    "github.com/Jim3Things/CloudChamber/pkg/protos/common"
)

func TestStepperGetNow(t *testing.T) {
    unit_test.SetTesting(t)
    defer unit_test.SetTesting(nil)

    request := httptest.NewRequest("GET", fmt.Sprintf("%s%s", baseURI, "/api/stepper/now"), nil)
    response := doHTTP(request, nil)

    assert.Equal(t, http.StatusOK, response.StatusCode)

    res := &common.Timestamp{}
    err := getJsonBody(response, res)

    assert.Nilf(t, err, "Unexpected error, err: %v", err)
    t.Log(res.Ticks)

    request = httptest.NewRequest("GET", fmt.Sprintf("%s%s", baseURI, "/api/stepper/now"), nil)
    response = doHTTP(request, nil)

    assert.Equal(t, http.StatusOK, response.StatusCode)

    res2 := &common.Timestamp{}
    err = getJsonBody(response, res2)

    assert.Nilf(t, err, "Unexpected error, err: %v", err)
    assert.Equal(t, res.Ticks, res2.Ticks, "times with no advance do not match, %v != %v", res.Ticks, res2.Ticks)
}
