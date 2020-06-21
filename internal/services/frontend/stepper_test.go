package frontend

import (
    "fmt"
    "net/http/httptest"
    "testing"

    "github.com/Jim3Things/CloudChamber/internal/tracing/exporters/unit_test"
)

func TestStepperGetNow(t *testing.T) {
    unit_test.SetTesting(t)

    request := httptest.NewRequest("GET", fmt.Sprintf("%s%s", baseURI, "/api/stepper/now"), nil)
    response := doHTTP(request, nil)
    body, err := getBody(response)

    t.Log(string(body))
    t.Logf("Err=%v", err)
}
