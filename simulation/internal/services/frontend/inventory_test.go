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

	pb "github.com/Jim3Things/CloudChamber/simulation/pkg/protos/inventory"
)

type InventoryTestSuite struct {
	testSuiteCore
}

func (ts *InventoryTestSuite) racksPath() string             { return ts.baseURI + "/api/racks/" }
func (ts *InventoryTestSuite) rackInPath(rack string) string { return ts.racksPath() + rack + "/" }
func (ts *InventoryTestSuite) bladesInPath(rack string) string {
	return ts.rackInPath(rack) + "blades/"
}
func (ts *InventoryTestSuite) bladeInPath(rack string, bladeID int) string {
	return fmt.Sprintf("%s%d", ts.bladesInPath(rack), bladeID)
}

func (ts *InventoryTestSuite) SetupSuite() {
	ts.testSuiteCore.SetupSuite()
}

func (ts *InventoryTestSuite) SetupTest() {
	_ = ts.utf.Open(ts.T())
}

func (ts *InventoryTestSuite) TearDownTest() {
	ts.utf.Close()
}

// First DBInventory unit test
func (ts *InventoryTestSuite) TestListRacks() {
	assert := ts.Assert()
	require := ts.Require()

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	request := httptest.NewRequest("GET", ts.racksPath(), nil)
	response = ts.doHTTP(request, response.Cookies())
	assert.HTTPStatusOK(response)

	list := &pb.External_ZoneSummary{}
	assert.NoError(ts.getJSONBody(response, list))

	assert.Equal(int64(8), list.MaxBladeCount)
	assert.Equal(int64(32), list.MaxCapacity.Cores)
	assert.Equal(int64(16384), list.MaxCapacity.MemoryInMb)
	assert.Equal(int64(240), list.MaxCapacity.DiskInGb)
	assert.Equal(int64(2*1024), list.MaxCapacity.NetworkBandwidthInMbps)

	require.NotNil(list.Racks)
	assert.Equal(8, len(list.Racks))

	r, ok := list.Racks["rack1"]
	assert.True(ok)
	assert.Equal(ts.rackInPath("rack1"), r.Uri)

	r, ok = list.Racks["rack2"]
	assert.True(ok)
	assert.Equal(ts.rackInPath("rack2"), r.Uri)

	ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())
}

// Inventory rack read test
func (ts *InventoryTestSuite) TestRackRead() {
	require := ts.Require()

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	request := httptest.NewRequest("GET", ts.rackInPath("Rack1"), nil)
	request.Header.Set("Content-Type", "application/json")

	response = ts.doHTTP(request, response.Cookies())

	require.HTTPStatusOK(response)

	rack := &pb.External_Rack{}
	require.NoError(ts.getJSONBody(response, rack))

	require.True(rack.GetDetails().GetEnabled())
	require.Equal(pb.Condition_operational, rack.GetDetails().GetCondition())
	require.Equal("Pacific NW, rack 1", rack.GetDetails().GetLocation())
	require.Equal("rack definition, 1 pdu, 1 tor, 8 blades", rack.GetDetails().GetNotes())

	require.NotNil(rack.Blades)
	require.Equal(8, len(rack.Blades))
	_, ok := rack.Blades[1]
	require.True(ok, "Blade 1 not found")
	_, ok = rack.Blades[2]
	require.True(ok, "Blade 2 not found")

	ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())
}

// Reading a rack that does not exist - should get status not found error
func (ts *InventoryTestSuite) TestUnknownRack() {
	assert := ts.Assert()

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	request := httptest.NewRequest("GET", ts.rackInPath("Rack9"), nil)
	request.Header.Set("Content-Type", "application/json")

	response = ts.doHTTP(request, response.Cookies())
	assert.Equal(http.StatusNotFound, response.StatusCode, "Handler returned the expected error: %v", response.StatusCode)

	ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())
}

func (ts *InventoryTestSuite) TestListBlades() {
	assert := ts.Assert()

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	request := httptest.NewRequest("GET", ts.bladesInPath("rack1"), nil)
	response = ts.doHTTP(request, response.Cookies())
	assert.HTTPStatusOK(response)

	body, err := ts.getBody(response)
	assert.Equal("text/plain; charset=utf-8", strings.ToLower(response.Header.Get("Content-Type")))

	var splits = strings.Split(string(body), "\n") // Created an array per line

	expected := []string{
		ts.bladeInPath("rack1", 1),
		ts.bladeInPath("rack1", 2),
		ts.bladeInPath("rack1", 3),
		ts.bladeInPath("rack1", 4),
		ts.bladeInPath("rack1", 5),
		ts.bladeInPath("rack1", 6),
		ts.bladeInPath("rack1", 7),
		ts.bladeInPath("rack1", 8),
		"",
	}

	assert.Equal(splits[0], "Blades in \"rack1\" (List)")
	assert.ElementsMatch(expected, splits[1:])

	assert.NoError(err)

	ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())
}

func (ts *InventoryTestSuite) TestUnknownBlade() {
	assert := ts.Assert()

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	request := httptest.NewRequest("GET", ts.bladeInPath("rack1", 9), nil)
	request.Header.Set("Content-Type", "application/json")

	response = ts.doHTTP(request, response.Cookies())
	assert.Equal(http.StatusNotFound, response.StatusCode, "Handler returned the expected error: %v", response.StatusCode)

	ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())
}

func (ts *InventoryTestSuite) TestNegativeBlade() {
	assert := ts.Assert()

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	request := httptest.NewRequest("GET", ts.bladeInPath("rack1", -1), nil)
	request.Header.Set("Content-Type", "application/json")

	response = ts.doHTTP(request, response.Cookies())
	assert.Equal(http.StatusNotFound, response.StatusCode, "Handler returned the expected error: %v", response.StatusCode)

	ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())
}

func (ts *InventoryTestSuite) TestZeroBlade() {
	assert := ts.Assert()

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	request := httptest.NewRequest("GET", ts.bladeInPath("rack1", 0), nil)
	request.Header.Set("Content-Type", "application/json")

	response = ts.doHTTP(request, response.Cookies())
	assert.Equal(http.StatusNotFound, response.StatusCode, "Handler returned the expected error: %v", response.StatusCode)

	ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())
}

func (ts *InventoryTestSuite) TestStringBlade() {
	assert := ts.Assert()

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	request := httptest.NewRequest("GET", ts.bladesInPath("rack1")+"Jeff", nil)
	request.Header.Set("Content-Type", "application/json")
	response = ts.doHTTP(request, response.Cookies())

	assert.Equal(
		http.StatusBadRequest,
		response.StatusCode,
		"Handler returned the expected error: %d", response.StatusCode)

	ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())
}

func (ts *InventoryTestSuite) TestBadRackBlade() {
	assert := ts.Assert()

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	request := httptest.NewRequest("GET", ts.bladeInPath("rack9", 2), nil)
	request.Header.Set("Content-Type", "application/json")

	response = ts.doHTTP(request, response.Cookies())
	assert.Equal(http.StatusNotFound, response.StatusCode, "Handler returned the expected error: %v", response.StatusCode)

	ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())
}

func (ts *InventoryTestSuite) TestBladeRead() {
	assert := ts.Assert()

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	request := httptest.NewRequest("GET", ts.bladeInPath("rack1", 1), nil)
	request.Header.Set("Content-Type", "application/json")

	response = ts.doHTTP(request, response.Cookies())
	assert.HTTPStatusOK(response)

	blade := &pb.BladeCapacity{}
	assert.NoError(ts.getJSONBody(response, blade))

	assert.Equal(int64(16), blade.Cores)
	assert.Equal(int64(16384), blade.MemoryInMb)
	assert.Equal("X64", blade.Arch)
	assert.Equal(int64(240), blade.DiskInGb)
	assert.Equal(int64(2048), blade.NetworkBandwidthInMbps)
	assert.Equal(0, len(blade.Accelerators))

	ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())
}

func (ts *InventoryTestSuite) TestBlade2Read() {
	assert := ts.Assert()

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	request := httptest.NewRequest("GET", ts.bladeInPath("rack1", 2), nil)
	request.Header.Set("Content-Type", "application/json")

	response = ts.doHTTP(request, response.Cookies())
	assert.HTTPStatusOK(response)

	blade := &pb.BladeCapacity{}
	assert.NoError(ts.getJSONBody(response, blade))

	assert.Equal(int64(32), blade.Cores)
	assert.Equal(int64(16384), blade.MemoryInMb)
	assert.Equal("X64", blade.Arch)
	assert.Equal(int64(120), blade.DiskInGb)
	assert.Equal(int64(2048), blade.NetworkBandwidthInMbps)
	assert.Equal(0, len(blade.Accelerators))

	ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())
}

// The purpose of this test is to check that the Inventory function get
// executed in a valid & established http session only
func (ts *InventoryTestSuite) TestNoSession() {
	assert := ts.Assert()

	request := httptest.NewRequest("GET", ts.racksPath(), nil)
	response := ts.doHTTP(request, nil)

	assert.Equal(http.StatusForbidden, response.StatusCode,
		"Handler returned %v, rather than %v", response.StatusCode, http.StatusForbidden)
}

func (ts *InventoryTestSuite) TestNoSessionRack() {
	assert := ts.Assert()

	request := httptest.NewRequest("GET", ts.rackInPath("rack1"), nil)
	response := ts.doHTTP(request, nil)

	assert.Equal(http.StatusForbidden, response.StatusCode,
		"Handler returned %v, rather than %v", response.StatusCode, http.StatusForbidden)
}

func (ts *InventoryTestSuite) TestNoSessionBlade() {
	assert := ts.Assert()

	request := httptest.NewRequest("GET", ts.bladeInPath("rack1", 1), nil)
	response := ts.doHTTP(request, nil)

	assert.Equal(http.StatusForbidden, response.StatusCode,
		"Handler returned %v, rather than %v", response.StatusCode, http.StatusForbidden)
}

func TestInventoryTestSuite(t *testing.T) {
	suite.Run(t, new(InventoryTestSuite))
}
