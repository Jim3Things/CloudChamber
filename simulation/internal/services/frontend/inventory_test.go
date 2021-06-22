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
	require := ts.Require()

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	request := httptest.NewRequest("GET", ts.racksPath(), nil)
	response = ts.doHTTP(request, response.Cookies())
	require.HTTPRSuccess(response)

	list := &pb.External_ZoneSummary{}
	require.NoError(ts.getJSONBody(response, list))

	require.EqualValues(1, list.MaxTorCount)
	require.EqualValues(1, list.MaxPduCount)
	require.EqualValues(9, list.MaxConnectors)
	require.EqualValues(8, list.MaxBladeCount)

	require.Equal(int64(32), list.MaxCapacity.Cores)
	require.Equal(int64(16384), list.MaxCapacity.MemoryInMb)
	require.Equal(int64(240), list.MaxCapacity.DiskInGb)
	require.Equal(int64(2*1024), list.MaxCapacity.NetworkBandwidthInMbps)

	require.NotNil(list.Racks)
	require.Equal(8, len(list.Racks))

	r, ok := list.Racks["rack1"]
	require.True(ok)
	require.Equal(ts.rackInPath("rack1"), r.Uri)

	r, ok = list.Racks["rack2"]
	require.True(ok)
	require.Equal(ts.rackInPath("rack2"), r.Uri)

	ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())
}

// Inventory rack read test
func (ts *InventoryTestSuite) TestRackRead() {
	require := ts.Require()

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	request := httptest.NewRequest("GET", ts.rackInPath("Rack1"), nil)
	request.Header.Set("Content-Type", "application/json")

	response = ts.doHTTP(request, response.Cookies())

	require.HTTPRSuccess(response)

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
	require := ts.Require()

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	request := httptest.NewRequest("GET", ts.rackInPath("Rack9"), nil)
	request.Header.Set("Content-Type", "application/json")

	response = ts.doHTTP(request, response.Cookies())
	require.HTTPRStatusEqual(http.StatusNotFound, response)

	ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())
}

func (ts *InventoryTestSuite) TestListBlades() {
	require := ts.Require()

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	request := httptest.NewRequest("GET", ts.bladesInPath("rack1"), nil)
	response = ts.doHTTP(request, response.Cookies())
	require.HTTPRSuccess(response)

	body := ts.getBody(response)
	require.HTTPRContentTypeEqual("text/plain; charset=utf-8", response)

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

	require.Equal(splits[0], "Blades in \"rack1\" (List)")
	require.ElementsMatch(expected, splits[1:])

	ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())
}

func (ts *InventoryTestSuite) TestUnknownBlade() {
	require := ts.Require()

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	request := httptest.NewRequest("GET", ts.bladeInPath("rack1", 9), nil)
	request.Header.Set("Content-Type", "application/json")

	response = ts.doHTTP(request, response.Cookies())
	require.HTTPRStatusEqual(http.StatusNotFound, response)

	ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())
}

func (ts *InventoryTestSuite) TestNegativeBlade() {
	require := ts.Require()

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	request := httptest.NewRequest("GET", ts.bladeInPath("rack1", -1), nil)
	request.Header.Set("Content-Type", "application/json")

	response = ts.doHTTP(request, response.Cookies())
	require.HTTPRStatusEqual(http.StatusNotFound, response)

	ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())
}

func (ts *InventoryTestSuite) TestZeroBlade() {
	require := ts.Require()

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	request := httptest.NewRequest("GET", ts.bladeInPath("rack1", 0), nil)
	request.Header.Set("Content-Type", "application/json")

	response = ts.doHTTP(request, response.Cookies())
	require.HTTPRStatusEqual(http.StatusNotFound, response)

	ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())
}

func (ts *InventoryTestSuite) TestStringBlade() {
	require := ts.Require()

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	request := httptest.NewRequest("GET", ts.bladesInPath("rack1")+"Jeff", nil)
	request.Header.Set("Content-Type", "application/json")
	response = ts.doHTTP(request, response.Cookies())

	require.HTTPRStatusEqual(http.StatusBadRequest, response)

	ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())
}

func (ts *InventoryTestSuite) TestBadRackBlade() {
	require := ts.Require()

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	request := httptest.NewRequest("GET", ts.bladeInPath("rack9", 2), nil)
	request.Header.Set("Content-Type", "application/json")

	response = ts.doHTTP(request, response.Cookies())
	require.HTTPRStatusEqual(http.StatusNotFound, response)

	ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())
}

func (ts *InventoryTestSuite) TestBladeRead() {
	require := ts.Require()

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	request := httptest.NewRequest("GET", ts.bladeInPath("rack1", 1), nil)
	request.Header.Set("Content-Type", "application/json")

	response = ts.doHTTP(request, response.Cookies())
	require.HTTPRSuccess(response)

	blade := &pb.BladeCapacity{}
	require.NoError(ts.getJSONBody(response, blade))

	require.Equal(int64(16), blade.Cores)
	require.Equal(int64(16384), blade.MemoryInMb)
	require.Equal("X64", blade.Arch)
	require.Equal(int64(240), blade.DiskInGb)
	require.Equal(int64(2048), blade.NetworkBandwidthInMbps)
	require.Equal(0, len(blade.Accelerators))

	ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())
}

func (ts *InventoryTestSuite) TestBlade2Read() {
	require := ts.Require()

	response := ts.doLogin(ts.randomCase(ts.adminAccountName()), ts.adminPassword(), nil)

	request := httptest.NewRequest("GET", ts.bladeInPath("rack1", 2), nil)
	request.Header.Set("Content-Type", "application/json")

	response = ts.doHTTP(request, response.Cookies())
	require.HTTPRSuccess(response)

	blade := &pb.BladeCapacity{}
	require.NoError(ts.getJSONBody(response, blade))

	require.Equal(int64(32), blade.Cores)
	require.Equal(int64(16384), blade.MemoryInMb)
	require.Equal("X64", blade.Arch)
	require.Equal(int64(120), blade.DiskInGb)
	require.Equal(int64(2048), blade.NetworkBandwidthInMbps)
	require.Equal(0, len(blade.Accelerators))

	ts.doLogout(ts.randomCase(ts.adminAccountName()), response.Cookies())
}

// The purpose of this test is to check that the Inventory function get
// executed in a valid & established http session only
func (ts *InventoryTestSuite) TestNoSession() {
	require := ts.Require()

	request := httptest.NewRequest("GET", ts.racksPath(), nil)
	response := ts.doHTTP(request, nil)

	require.HTTPRStatusEqual(http.StatusForbidden, response)
}

func (ts *InventoryTestSuite) TestNoSessionRack() {
	require := ts.Require()

	request := httptest.NewRequest("GET", ts.rackInPath("rack1"), nil)
	response := ts.doHTTP(request, nil)

	require.HTTPRStatusEqual(http.StatusForbidden, response)
}

func (ts *InventoryTestSuite) TestNoSessionBlade() {
	require := ts.Require()

	request := httptest.NewRequest("GET", ts.bladeInPath("rack1", 1), nil)
	response := ts.doHTTP(request, nil)

	require.HTTPRStatusEqual(http.StatusForbidden, response)
}

func TestInventoryTestSuite(t *testing.T) {
	suite.Run(t, new(InventoryTestSuite))
}
