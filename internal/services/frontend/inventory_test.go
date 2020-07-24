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

	"github.com/Jim3Things/CloudChamber/internal/tracing/exporters/unit_test"
	"github.com/Jim3Things/CloudChamber/pkg/protos/common"
	pb "github.com/Jim3Things/CloudChamber/pkg/protos/inventory"
	"github.com/stretchr/testify/assert"
)

// First DBInventory unit test
func TestInventoryListRacks(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	request := httptest.NewRequest("GET", "/api/racks/", nil)
	response := doHTTP(request, nil)
	body, err := getBody(response)

	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)
	assert.Equal(t, "text/plain; charset=utf-8", strings.ToLower(response.Header.Get("Content-Type")))
	s := string(body)                   // Converted into a string
	var splits = strings.Split(s, "\n") // Created an array per line

	expected := []string{"/api/racks/rack1", "/api/racks/rack2", ""}

	assert.Equal(t, splits[0], "Racks (List)")
	assert.ElementsMatch(t, expected, splits[1:])

	assert.Nil(t, err)
}

// Inventory rack read test
func TestInventoryRackRead(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	request := httptest.NewRequest("GET", fmt.Sprintf("%s%s", baseURI, "/api/racks/Rack1"), nil)
	request.Header.Set("Content-Type", "application/json")

	response := doHTTP(request, nil)

	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)

	rack := &pb.ExternalRack{}
	err := getJSONBody(response, rack)
	assert.Nilf(t, err, "Failed to convert body to valid json.  err: %v", err)

	assert.Equal(t, "application/json", strings.ToLower(response.Header.Get("Content-Type")))
	assert.Equal(t, 2, len(rack.Blades))
	_, ok := rack.Blades[1]
	assert.True(t, ok, "Blade 1 not found")
	_, ok = rack.Blades[2]
	assert.True(t, ok, "Blade 2 not found")
}

//Reading a rack that do not exist - should get status not found error
func TestInventoryUnknownRack(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	request := httptest.NewRequest("GET", fmt.Sprintf("%s%s", baseURI, "/api/racks/Rack3"), nil)
	request.Header.Set("Content-Type", "application/json")

	response := doHTTP(request, nil)
	assert.Equal(t, http.StatusNotFound, response.StatusCode, "Handler returned the expected error: %v", response.StatusCode)

}

func TestInventoryListBlades(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	request := httptest.NewRequest("GET", "/api/racks/rack1/blades", nil)
	response := doHTTP(request, nil)
	body, err := getBody(response)
	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned unexpected error: %v", response.StatusCode)
	assert.Equal(t, "text/plain; charset=utf-8", strings.ToLower(response.Header.Get("Content-Type")))
	s := string(body)                   // Converted into a string
	var splits = strings.Split(s, "\n") // Created an array per line

	expected := []string{"/api/racks/rack1/blades/1", "/api/racks/rack1/blades/2", ""}

	assert.Equal(t, splits[0], "Blades in \"rack1\" (List)")
	assert.ElementsMatch(t, expected, splits[1:])

	assert.Nil(t, err)
}
func TestInventoryUnknownBlade(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	request := httptest.NewRequest("GET", fmt.Sprintf("%s%s", baseURI, "/api/racks/rack1/blades/3"), nil)
	request.Header.Set("Content-Type", "application/json")

	response := doHTTP(request, nil)
	assert.Equal(t, http.StatusNotFound, response.StatusCode, "Handler returned the expected error: %v", response.StatusCode)
}
func TestInventoryNegativeBlade(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	request := httptest.NewRequest("GET", fmt.Sprintf("%s%s", baseURI, "/api/racks/rack1/blades/-1"), nil)
	request.Header.Set("Content-Type", "application/json")

	response := doHTTP(request, nil)
	assert.Equal(t, http.StatusNotFound, response.StatusCode, "Handler returned the expected error: %v", response.StatusCode)
}
func TestInventoryZeroBlade(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	request := httptest.NewRequest("GET", fmt.Sprintf("%s%s", baseURI, "/api/racks/rack1/blades/0"), nil)
	request.Header.Set("Content-Type", "application/json")

	response := doHTTP(request, nil)
	assert.Equal(t, http.StatusNotFound, response.StatusCode, "Handler returned the expected error: %v", response.StatusCode)
}
func TestInventoryStringBlade(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	request := httptest.NewRequest("GET", fmt.Sprintf("%s%s", baseURI, "/api/racks/rack1/blades/Jeff"), nil)
	request.Header.Set("Content-Type", "application/json")
	response := doHTTP(request, nil)

	assert.Equal(t, http.StatusBadRequest, response.StatusCode, "Handler returned the expected error: %d", response.StatusCode)
}
func TestInventoryBadRackBlade(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	request := httptest.NewRequest("GET", fmt.Sprintf("%s%s", baseURI, "/api/racks/rack3/blades/2"), nil)
	request.Header.Set("Content-Type", "application/json")

	response := doHTTP(request, nil)
	assert.Equal(t, http.StatusNotFound, response.StatusCode, "Handler returned the expected error: %v", response.StatusCode)
}

func TestInventoryBladeRead(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	request := httptest.NewRequest("GET", fmt.Sprintf("%s%s", baseURI, "/api/racks/rack1/blades/1"), nil)
	request.Header.Set("Content-Type", "application/json")

	response := doHTTP(request, nil)
	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned the Blade: %v", response.StatusCode)

	blade := &common.BladeCapacity{}
	err := getJSONBody(response, blade)
	assert.Nilf(t, err, "Failed to convert body to valid json.  err: %v", err)

	assert.Equal(t, "application/json", strings.ToLower(response.Header.Get("Content-Type")))
	assert.Equal(t, int64(8), blade.Cores)
	assert.Equal(t, int64(16384), blade.MemoryInMb)
	assert.Equal(t, "X64", blade.Arch)
	assert.Equal(t, int64(120), blade.DiskInGb)
	assert.Equal(t, int64(1024), blade.NetworkBandwidthInMbps)
	assert.Equal(t, 0, len(blade.Accelerators))

}
func TestInventoryBlade2Read(t *testing.T) {
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	request := httptest.NewRequest("GET", fmt.Sprintf("%s%s", baseURI, "/api/racks/rack1/blades/2"), nil)
	request.Header.Set("Content-Type", "application/json")

	response := doHTTP(request, nil)
	assert.Equal(t, http.StatusOK, response.StatusCode, "Handler returned the Blade: %v", response.StatusCode)

	blade := &common.BladeCapacity{}
	err := getJSONBody(response, blade)
	assert.Nilf(t, err, "Failed to convert body to valid json.  err: %v", err)

	assert.Equal(t, "application/json", strings.ToLower(response.Header.Get("Content-Type")))
	assert.Equal(t, int64(16), blade.Cores)
	assert.Equal(t, int64(16384), blade.MemoryInMb)
	assert.Equal(t, "X64", blade.Arch)
	assert.Equal(t, int64(240), blade.DiskInGb)
	assert.Equal(t, int64(2048), blade.NetworkBandwidthInMbps)
	assert.Equal(t, 0, len(blade.Accelerators))

}
