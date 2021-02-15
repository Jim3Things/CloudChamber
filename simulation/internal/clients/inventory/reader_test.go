// Test to parse the Inventory.Yaml file
package inventory

import (
	"context"
	"fmt"
	"testing"

	"github.com/spf13/viper"

	"github.com/stretchr/testify/suite"

	pb "github.com/Jim3Things/CloudChamber/simulation/pkg/protos/inventory"
)

type readerTestSuite struct {
	testSuiteCore	
}

func (ts *readerTestSuite) SetupSuite() {
	ts.testSuiteCore.SetupSuite()
}

func (ts *readerTestSuite) SetupTest() {
	require := ts.Require()

	require.NoError(ts.utf.Open(ts.T()))
	require.NoError(ts.store.Connect())

	viper.Reset()
}

func (ts *readerTestSuite) TearDownTest() {
	ts.store.Disconnect()
	ts.utf.Close()
}

// first inventory definition test

func (ts *readerTestSuite) TestReadInventoryDefinition() {
	assert := ts.Assert()
	require := ts.Require()

	response, err := ReadInventoryDefinition(context.Background(), "./testdata/Basic")
	require.Nil(err)

	require.Equal(2, len(response.Racks))

	r, ok := response.Racks["rack1"]
	require.True(ok)
	assert.Equal(2, len(r.Blades))

	b, ok := r.Blades[1]
	require.True(ok)
	assert.Equal(int64(16), b.Cores)
	assert.Equal(int64(16834), b.MemoryInMb)
	assert.Equal(int64(240), b.DiskInGb)
	assert.Equal(int64(2048), b.NetworkBandwidthInMbps)
	assert.Equal("X64", b.Arch)

	s, ok := response.Racks["rack2"]
	require.True(ok)
	assert.Equal(2, len(s.Blades))

	c, ok := r.Blades[2]
	require.True(ok)
	assert.Equal(int64(8), c.Cores)
	assert.Equal(int64(16834), c.MemoryInMb)
	assert.Equal(int64(120), c.DiskInGb)
	assert.Equal(int64(2048), c.NetworkBandwidthInMbps)
}

func (ts *readerTestSuite) TestReadInventoryBogusPath() {
	assert := ts.Assert()
	require := ts.Require()

	response, err := ReadInventoryDefinition(context.Background(), "./missing/path")
	require.NotNil(err)
	assert.NotEqual("%v", response)
}

// TestInventoryUniqueRack test to check that zone always contain unique rack numbers
func (ts *readerTestSuite) TestInventoryUniqueRack() {
	assert := ts.Assert()
	require := ts.Require()

	_, err := ReadInventoryDefinition(context.Background(), "./testdata/BadYaml")
	require.NotNil(err)
	assert.Equal("Duplicate rack \"rack1\" detected", err.Error())
}

func (ts *readerTestSuite) TestInventoryUniqueBlade() {
	assert := ts.Assert()
	require := ts.Require()

	_, err := ReadInventoryDefinition(context.Background(), "./testdata/BadYamlBlade")
	require.NotNil(err)
	assert.Equal("Duplicate Blade 1 in Rack \"rack1\" detected", err.Error())
}

func (ts *readerTestSuite) TestInventoryValidateBlade() {
	assert := ts.Assert()
	require := ts.Require()

	_, err := ReadInventoryDefinition(context.Background(), "./testdata/BadYamlValidate")
	require.NotNil(err)
	assert.Equal("In rack \"rack1\": the field \"Blades[2].Cores\" must be greater than or equal to 1.  It is 0, which is invalid",
				 err.Error())
}

func (ts *readerTestSuite) TestReadInventoryDefinitionFromFile() {
	assert := ts.Assert()
	require := ts.Require()

	zonemap, err := ReadInventoryDefinitionFromFile(context.Background(), "./testdata/Basic")
	require.NoError(err)

	// There should only be a single zone.
	//
	require.Equal(1, len(zonemap.Zones))

	zone, ok := zonemap.Zones[DefaultZone]
	require.True(ok)

	assert.True(zone.Details.Enabled)
	assert.Equal(pb.State_in_service, zone.Details.State)
	assert.Equal("DC-PNW-0", zone.Details.Location)
	assert.Equal("Base zone", zone.Details.Notes)

	require.Equal(2, len(zone.Racks))

	for i := 1; i <= 2; i++ {

		name := fmt.Sprintf("rack%d", i)

		r, ok := zone.Racks[name]
		require.True(ok)

		assert.True(r.Details.Enabled)
		assert.Equal(pb.Condition_operational, r.Details.Condition)
		assert.Equal("DC-PNW-0-" + name, r.Details.Location)
		assert.Equal("RackName: " + name, r.Details.Notes)

		assert.Equal(1, len(r.Pdus))
		assert.Equal(1, len(r.Tors))
		assert.Equal(2, len(r.Blades))

		// There should be a single PDU at index 0
		//
		p0, ok := r.Pdus[0]
		require.True(ok)

		// The PDU should have a wired port for each of the two expected blades.
		//
		assert.Equal(2, len(p0.Ports))

		p0b1, ok := p0.Ports[1]
		require.True(ok)

		assert.True(p0b1.Wired)
		assert.Equal(pb.Hardware_blade, p0b1.Item.Type)
		assert.Equal(int64(1), p0b1.Item.Id)
		assert.Equal(int64(0), p0b1.Item.Port)

		p0b2, ok := p0.Ports[2]
		require.True(ok)

		assert.True(p0b2.Wired)
		assert.Equal(pb.Hardware_blade, p0b2.Item.Type)
		assert.Equal(int64(2), p0b2.Item.Id)
		assert.Equal(int64(0), p0b2.Item.Port)

		// There should be a single TOR at index 0
		//
		t0, ok := r.Tors[0]
		require.True(ok)

		// The TOR should have a wired port for each of the two expected blades.
		//
		assert.Equal(2, len(t0.Ports))

		t0b1, ok := t0.Ports[1]
		require.True(ok)

		assert.True(t0b1.Wired)
		assert.Equal(pb.Hardware_blade, t0b1.Item.Type)
		assert.Equal(int64(1), t0b1.Item.Id)
		assert.Equal(int64(0), t0b1.Item.Port)

		t0b2, ok := p0.Ports[2]
		require.True(ok)

		assert.True(t0b2.Wired)
		assert.Equal(pb.Hardware_blade, t0b2.Item.Type)
		assert.Equal(int64(2), t0b2.Item.Id)
		assert.Equal(int64(0), t0b2.Item.Port)

		// There should be exactly two blades at indices 1 and 2.
		//
		b1, ok := r.Blades[1]
		require.True(ok)

		assert.True(b1.Details.Enabled)
		assert.Equal(pb.Condition_operational, b1.Details.Condition)

		assert.Equal(int64(16),    b1.Capacity.Cores)
		assert.Equal(int64(16834), b1.Capacity.MemoryInMb)
		assert.Equal(int64(240),   b1.Capacity.DiskInGb)
		assert.Equal(int64(2048),  b1.Capacity.NetworkBandwidthInMbps)
		assert.Equal("X64",        b1.Capacity.Arch)

		b2, ok := r.Blades[2]
		require.True(ok)

		assert.True(b2.Details.Enabled)
		assert.Equal(pb.Condition_operational, b2.Details.Condition)

		assert.Equal(int64(8),     b2.Capacity.Cores)
		assert.Equal(int64(16834), b2.Capacity.MemoryInMb)
		assert.Equal(int64(120),   b2.Capacity.DiskInGb)
		assert.Equal(int64(2048),  b2.Capacity.NetworkBandwidthInMbps)
		assert.Equal("X64",        b2.Capacity.Arch)
	}
}

func (ts *readerTestSuite) TestReadInventoryDefinitionFromFileBogusPath() {
	assert := ts.Assert()
	require := ts.Require()

	response, err := ReadInventoryDefinitionFromFile(context.Background(), "./missing/path")
	require.Error(err)
	assert.NotEqual("%v", response)
}

// TestInventoryUniqueRack test to check that zone always contain unique rack numbers
//
func (ts *readerTestSuite) TestIReadInventoryDefinitionFromFileUniqueRack() {
	assert := ts.Assert()
	require := ts.Require()

	_, err := ReadInventoryDefinitionFromFile(context.Background(), "./testdata/BadYaml")
	require.Error(err)
	assert.Equal("Duplicate rack \"rack1\" detected", err.Error())
}

func (ts *readerTestSuite) TestReadInventoryDefinitionFromFileUniqueBlade() {
	assert := ts.Assert()
	require := ts.Require()

	_, err := ReadInventoryDefinitionFromFile(context.Background(), "./testdata/BadYamlBlade")
	require.Error(err)
	assert.Equal("Duplicate Blade 1 in Rack \"rack1\" detected", err.Error())
}

func (ts *readerTestSuite) TestReadInventoryDefinitionFromFileValidateBlade() {
	assert := ts.Assert()
	require := ts.Require()

	_, err := ReadInventoryDefinitionFromFile(context.Background(), "./testdata/BadYamlValidate")
	require.Error(err)
	assert.Equal("In rack \"rack1\": the field \"Blades[2].Cores\" must be greater than or equal to 1.  It is 0, which is invalid",
			err.Error())
}

func TestReaderTestSuite(t *testing.T) {
	suite.Run(t, new(readerTestSuite))
}
