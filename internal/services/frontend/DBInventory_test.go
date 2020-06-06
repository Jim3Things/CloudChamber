// Unit tests for the web service frontend.
//
// Borrows heavily from the gorilla mux test package.
//
package frontend

import (
	"log"
	"net/http/httptest"
	"testing"

	"github.com/Jim3Things/CloudChamber/internal/tracing/exporters/unit_test"
)

// // First DBInventory unit test
func TestListRacks(t *testing.T) {
	unit_test.SetTesting(t)

	request := httptest.NewRequest("GET", "/api/racks/", nil)
	response := doHTTP(request, nil)
	body, err := getBody(response)

	log.Printf("ListRacks: SC=%v, Content-Type='%v'\n", response.StatusCode, response.Header.Get("Content-Type"))
	log.Println(string(body))
	log.Printf("error: %v", err)
}

// List racks
