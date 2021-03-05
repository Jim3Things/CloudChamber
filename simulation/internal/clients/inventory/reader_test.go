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

	response, err := ReadInventoryDefinitionFromFileEx(context.Background(), "./testdata/Simple")
	require.NoError(err)
	require.NotNil(response)

	require.Equal(1, len(response.Regions))
	region1, ok := response.Regions["region1"]
	require.True(ok)
	assert.Equal(1, len(region1.Zones))

	zone1, ok := region1.Zones["zone1"]
	require.True(ok)
	require.Equal(2, len(zone1.Racks))

	r, ok := zone1.Racks["rack1"]
	require.True(ok)
	assert.Equal(2, len(r.Blades))

	b, ok := r.Blades[1]
	require.True(ok)
	require.NotNil(b.Capacity)
	assert.Equal(int64(16), b.Capacity.Cores)
	assert.Equal(int64(16384), b.Capacity.MemoryInMb)
	assert.Equal(int64(240), b.Capacity.DiskInGb)
	assert.Equal(int64(2048), b.Capacity.NetworkBandwidthInMbps)
	assert.Equal("X64", b.Capacity.Arch)

	s, ok := zone1.Racks["rack2"]
	require.True(ok)
	assert.Equal(2, len(s.Blades))

	c, ok := r.Blades[2]
	require.True(ok)
	require.NotNil(b.Capacity)
	assert.Equal(int64(8), c.Capacity.Cores)
	assert.Equal(int64(16384), c.Capacity.MemoryInMb)
	assert.Equal(int64(120), c.Capacity.DiskInGb)
	assert.Equal(int64(2048), c.Capacity.NetworkBandwidthInMbps)
	assert.Equal("X64", b.Capacity.Arch)
}

// func (ts *readerTestSuite) TestReadInventoryBogusPath() {
// 	require := ts.Require()

// 	response, err := ReadInventoryDefinitionFromFileEx(context.Background(), "./missing/path")
// 	require.EqualError(err, "no inventory definition found at ./missing/path/inventory.yaml (yaml)")
// 	require.Nil(response)
// }

// // TestInventoryUniqueRack test to check that zone always contain unique rack numbers
// func (ts *readerTestSuite) TestInventoryUniqueRack() {
// 	require := ts.Require()

// 	response, err := ReadInventoryDefinitionFromFileEx(context.Background(), "./testdata/BadYaml")
// 	require.EqualError(err, "Duplicate rack \"rack1\" detected")
// 	require.Nil(response)
// }

// func (ts *readerTestSuite) TestInventoryUniqueBlade() {
// 	require := ts.Require()

// 	response, err := ReadInventoryDefinitionFromFileEx(context.Background(), "./testdata/BadYamlBlade")
// 	require.EqualError(err, "Duplicate Blade 1 in Rack \"rack1\" detected")
// 	require.Nil(response)
// }

// func (ts *readerTestSuite) TestInventoryValidateBlade() {
// 	require := ts.Require()

// 	response, err := ReadInventoryDefinitionFromFileEx(context.Background(), "./testdata/BadYamlValidate")
// 	require.EqualError(err, "In rack \"rack1\": the field \"Blades[2].Cores\" must be greater than or equal to 1.  It is 0, which is invalid")
// 	require.Nil(response)
// }

func (ts *readerTestSuite) TestReadInventoryDefinitionFromFile() {
	assert := ts.Assert()
	require := ts.Require()

	response, err := ReadInventoryDefinitionFromFileEx(context.Background(), "./testdata/Simple")
	require.NoError(err)
	require.NotNil(response)

	require.Equal(1, len(response.Regions))
	region, ok := response.Regions["region1"]
	require.True(ok)

	// There should only be a single zone.
	//
	require.Equal(1, len(region.Zones))

	zone, ok := region.Zones["zone1"]
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

		// The PDU should have a wired port for each of the two expected blades and one for the tor.
		//
		assert.Equal(3, len(p0.Ports))

		p0b0, ok := p0.Ports[1]
		require.True(ok)

		assert.True(p0b0.Wired)
		assert.Equal(pb.Hardware_tor, p0b0.Item.Type)
		assert.Equal(int64(0), p0b0.Item.Id)
		assert.Equal(int64(1), p0b0.Item.Port)

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

		// The TOR should have a wired port for each of the two expected blades and one for the pdu.
		//
		assert.Equal(3, len(t0.Ports))

		t0b0, ok := t0.Ports[0]
		require.True(ok)

		assert.True(t0b0.Wired)
		assert.Equal(pb.Hardware_pdu, t0b0.Item.Type)
		assert.Equal(int64(0), t0b0.Item.Id)
		assert.Equal(int64(1), t0b0.Item.Port)

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
		assert.Equal(int64(16384), b1.Capacity.MemoryInMb)
		assert.Equal(int64(240),   b1.Capacity.DiskInGb)
		assert.Equal(int64(2048),  b1.Capacity.NetworkBandwidthInMbps)
		assert.Equal("X64",        b1.Capacity.Arch)

		b2, ok := r.Blades[2]
		require.True(ok)

		assert.True(b2.Details.Enabled)
		assert.Equal(pb.Condition_operational, b2.Details.Condition)

		assert.Equal(int64(8),     b2.Capacity.Cores)
		assert.Equal(int64(16384), b2.Capacity.MemoryInMb)
		assert.Equal(int64(120),   b2.Capacity.DiskInGb)
		assert.Equal(int64(2048),  b2.Capacity.NetworkBandwidthInMbps)
		assert.Equal("X64",        b2.Capacity.Arch)
	}
}

func (ts *readerTestSuite) TestReadInventoryDefinitionFromFileBogusPath() {
	require := ts.Require()

	response, err := ReadInventoryDefinitionFromFileEx(context.Background(), "./missing/path")
	require.EqualError(err, "no inventory definition found at ./missing/path/inventory.yaml (yaml)")
	require.Nil(response)
}

// TestInventoryUniqueRack test to check that zone always contain unique rack numbers
//
func (ts *readerTestSuite) TestIReadInventoryDefinitionFromFileUniqueRack() {
	require := ts.Require()

	response, err := ReadInventoryDefinitionFromFileEx(context.Background(), "./testdata/BadYaml")
	require.EqualError(err, "Duplicate rack \"rack1\" detected")
	require.Nil(response)
}

func (ts *readerTestSuite) TestReadInventoryDefinitionFromFileUniqueBlade() {
	require := ts.Require()

	response, err := ReadInventoryDefinitionFromFileEx(context.Background(), "./testdata/BadYamlBlade")
	require.EqualError(err, "Duplicate Blade 1 in Rack \"rack1\" detected")
	require.Nil(response)
}

func (ts *readerTestSuite) TestReadInventoryDefinitionFromFileValidateBlade() {
	require := ts.Require()

	response, err := ReadInventoryDefinitionFromFileEx(context.Background(), "./testdata/BadYamlValidate")
	require.EqualError(err, "In rack \"rack1\": the field \"Blades[2].Cores\" must be greater than or equal to 1.  It is 0, which is invalid")
	require.Nil(response)
}

func (ts *readerTestSuite) TestReadInventoryDefinitionBasic() {
	assert := ts.Assert()
	require := ts.Require()

	root, err := ReadInventoryDefinitionFromFileEx(context.Background(), "./testdata/Basic")
	require.NoError(err)
	require.NotNil(root)

	// There should only be a single region.
	//
	regionExpectedCount := 1
	regionExpectedName  := "Region1"

	assert.Equal(regionExpectedCount, len(root.Regions))

	for regionName, region := range root.Regions {

		if !assert.Equal(regionName, regionExpectedName, "Found unexpected region: %s", regionName) {
			continue
		}

		assert.Equal(pb.State_in_service, region.Details.State)
		assert.Equal("Pacific NW", region.Details.Location)
		assert.Equal("Test Region", region.Details.Notes)

		// There should only be a single zone.
		//
		zoneExpectedCount := 1
		zoneExpectedName := "zone1"

		assert.Equal(zoneExpectedCount, len(region.Zones))

		for zoneName, zone := range region.Zones {

			if !assert.Equal(zoneName, zoneExpectedName, "Found unexpected zone: %s", zoneName) {
				continue
			}

			assert.True(zone.Details.Enabled)
			assert.Equal(pb.State_in_service, zone.Details.State)
			assert.Equal("Pacific NW, row 1", zone.Details.Location)
			assert.Equal("Simple zone definition", zone.Details.Notes)

			// There should only be a single rack.
			//
			rackExpectedCount := 1
			rackExpectedName := "rack1"

			assert.Equal(rackExpectedCount, len(zone.Racks))

			for rackName, rack := range zone.Racks {

				if !assert.Equal(rackName, rackExpectedName, "Found unexpected rack: %s", rackName) {
					continue
				}
		
				assert.True(rack.Details.Enabled)
				assert.Equal(pb.Condition_operational, rack.Details.Condition)
				assert.Equal("Pacific NW, row 1, rack 1", rack.Details.Location)
				assert.Equal("Simple rack definition" , rack.Details.Notes)

				// There should be a single PDU at index 0
				//
				assert.Equal(1, len(rack.Pdus))

				p0, ok := rack.Pdus[0]
				require.True(ok)

				// The PDU should have a wired port for the tor and two ports
				// for one of the blades, and a single port for the other blade.
				//
				assert.Equal(4, len(p0.Ports))

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
				assert.Equal(int64(1), p0b2.Item.Id)
				assert.Equal(int64(1), p0b2.Item.Port)

				p0b3, ok := p0.Ports[3]
				require.True(ok)

				assert.True(p0b3.Wired)
				assert.Equal(pb.Hardware_tor, p0b3.Item.Type)
				assert.Equal(int64(0), p0b3.Item.Id)
				assert.Equal(int64(1), p0b3.Item.Port)

				p0b4, ok := p0.Ports[4]
				require.True(ok)

				assert.True(p0b4.Wired)
				assert.Equal(pb.Hardware_blade, p0b4.Item.Type)
				assert.Equal(int64(2), p0b4.Item.Id)
				assert.Equal(int64(4), p0b4.Item.Port)

				// There should be a single TOR at index 0
				//
				assert.Equal(1, len(rack.Tors))

				t0, ok := rack.Tors[0]
				require.True(ok)

				// The TOR should have a wired port for the pdu and two ports
				// for one of the blades, and a single port for the other blade.
				//
				assert.Equal(4, len(t0.Ports))

				t0b1, ok := t0.Ports[1]
				require.True(ok)

				assert.True(t0b1.Wired)
				assert.Equal(pb.Hardware_blade, t0b1.Item.Type)
				assert.Equal(int64(1), t0b1.Item.Id)
				assert.Equal(int64(0), t0b1.Item.Port)

				t0b2, ok := t0.Ports[2]
				require.True(ok)

				assert.True(t0b2.Wired)
				assert.Equal(pb.Hardware_blade, t0b2.Item.Type)
				assert.Equal(int64(1), t0b2.Item.Id)
				assert.Equal(int64(1), t0b2.Item.Port)

				t0b3, ok := t0.Ports[3]
				require.True(ok)

				assert.True(t0b3.Wired)
				assert.Equal(pb.Hardware_pdu, t0b3.Item.Type)
				assert.Equal(int64(0), t0b3.Item.Id)
				assert.Equal(int64(1), t0b3.Item.Port)

				t0b4, ok := t0.Ports[4]
				require.True(ok)

				assert.True(t0b4.Wired)
				assert.Equal(pb.Hardware_blade, t0b4.Item.Type)
				assert.Equal(int64(2), t0b4.Item.Id)
				assert.Equal(int64(3), t0b4.Item.Port)

				// There should be exactly two blades at indices 1 and 2
				//
				assert.Equal(2, len(rack.Blades))

				b1, ok := rack.Blades[1]
				require.True(ok)

				assert.True(b1.Details.Enabled)
				assert.Equal(pb.Condition_operational, b1.Details.Condition)

				assert.Equal(int64(16),    b1.Capacity.Cores)
				assert.Equal(int64(16384), b1.Capacity.MemoryInMb)
				assert.Equal(int64(240),   b1.Capacity.DiskInGb)
				assert.Equal(int64(2048),  b1.Capacity.NetworkBandwidthInMbps)
				assert.Equal("X64",        b1.Capacity.Arch)

				b2, ok := rack.Blades[2]
				require.True(ok)

				assert.True(b2.Details.Enabled)
				assert.Equal(pb.Condition_operational, b2.Details.Condition)

				assert.Equal(int64(24),    b2.Capacity.Cores)
				assert.Equal(int64(32768), b2.Capacity.MemoryInMb)
				assert.Equal(int64(480),   b2.Capacity.DiskInGb)
				assert.Equal(int64(4096),  b2.Capacity.NetworkBandwidthInMbps)
				assert.Equal("X64",        b2.Capacity.Arch)
			}
		}
	}
}

func TestReaderTestSuite(t *testing.T) {
	suite.Run(t, new(readerTestSuite))
}
