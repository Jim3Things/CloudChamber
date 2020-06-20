// Unit tests for the web service frontend.
//
// Borrows heavily from the gorilla mux test package.
//
package frontend

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Jim3Things/CloudChamber/internal/tracing/exporters/unit_test"
	pb "github.com/Jim3Things/CloudChamber/pkg/protos/inventory"
	"github.com/stretchr/testify/assert"
)

// // First DBInventory unit test
func TestInventoryListRacks(t *testing.T) {
	unit_test.SetTesting(t)

	request := httptest.NewRequest("GET", "/api/racks/", nil)
	response := doHTTP(request, nil)
	body, err := getBody(response)

	log.Printf("ListRacks: SC=%v, Content-Type='%v'\n", response.StatusCode, response.Header.Get("Content-Type"))

	s := string(body)                   // Converted into a string
	var splits = strings.Split(s, "\n") // Created an array per line
	fmt.Println(splits)                 // just a print statement
	expected := []string{"/api/racks/rack1", "/api/racks/rack2", ""}

	assert.Equal(t, splits[0], "Racks (List)")
	assert.ElementsMatch(t, expected, splits[1:])

	log.Printf("error: %v", err)
}

func TestInventoryListRead(t *testing.T) {
	unit_test.SetTesting(t)

	// response := doLogin(t, randomCase(adminAccountName), adminPassword, nil)

	request := httptest.NewRequest("GET", fmt.Sprintf("%s%s", baseURI, "/api/racks/Rack1"), nil)
	request.Header.Set("Content-Type", "application/json")

	response := doHTTP(request, nil)

	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	rack := &pb.ExternalRack{}
	err := getJsonBody(response, rack)
	assert.Nilf(t, err, "Failed to convert body to valid json.  err: %v", err)

	assert.Equal(t, "application/json", strings.ToLower(response.Header.Get("Content-Type")))
	//assert.Equal(t, "1", response.Header.Get("ETag"))
	assert.Equal(t, 2, len(rack.Blades))
	_, ok := rack.Blades[1]
	assert.True(t, ok, "Blade 1 not found")
	_, ok = rack.Blades[2]
	assert.True(t, ok, "Blade 2 not found")

	// doLogout(t, randomCase(adminAccountName), response.Cookies())
}
