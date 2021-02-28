package frontend

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Jim3Things/CloudChamber/simulation/internal/clients/inventory"
	"github.com/Jim3Things/CloudChamber/simulation/pkg/errors"
	pb "github.com/Jim3Things/CloudChamber/simulation/pkg/protos/inventory"
)

type DBInventoryTestSuite struct {
	testSuiteCore

	db *DBInventory

	regionName string
	zoneName   string
	rackName   string
	pduID      int64
	torID      int64
	bladeID    int64
}

func (ts *DBInventoryTestSuite) SetupSuite() {

	ts.testSuiteCore.SetupSuite()

	ts.regionName = "regionTest"
	ts.zoneName = "zoneTest"
	ts.rackName = "rackTest"
	ts.pduID = int64(0)
	ts.torID = int64(0)
	ts.bladeID = int64(100)

	// The standard "frontend" initialisation will create a dbInventory structure
	// which will lead to the initialization of the inventory within the store.
	// This means we can just use the global store as long as we remember that
	// any records written here will persist for this test session and so the
	// names use should not conflict with those being used in the standard
	// inventory definition file.
	//
	ts.db = dbInventory
}

func (ts *DBInventoryTestSuite) SetupTest() {
	_ = ts.utf.Open(ts.T())
	ts.Require().NoError(ts.cleanRegion(ts.regionName))

}

func (ts *DBInventoryTestSuite) TearDownTest() {
	ts.utf.Close()
}

func (ts *DBInventoryTestSuite) cleanRegion(regionName string) error {
	ctx := context.Background()

	err := ts.db.ScanZonesInRegion(regionName, func(zoneName string) error {
		err := ts.cleanZone(regionName, zoneName)
		if err != nil {
			return err
		}

		_, err = ts.db.DeleteZone(ctx, regionName, zoneName)
		return err
	})

	_, err = ts.db.DeleteRegion(ctx, regionName)

	return err
}

func (ts *DBInventoryTestSuite) cleanZone(regionName string, zoneName string) error {
	ctx := context.Background()

	err := ts.db.ScanRacksInZone(regionName, zoneName, func(rackName string) error {
		if err := ts.cleanRack(regionName, zoneName, rackName); err != nil {
			return err
		}

		_, err := ts.db.DeleteRack(ctx, regionName, zoneName, rackName)
		return err
	})

	_, err = ts.db.DeleteZone(ctx, regionName, zoneName)
	return err
}

func (ts *DBInventoryTestSuite) cleanRack(regionName string, zoneName string, rackName string) error {
	ctx := context.Background()

	if _, err := ts.db.DeletePdu(ctx, regionName, zoneName, rackName, ts.pduID); err != nil {
		return err
	}

	if _, err := ts.db.DeleteTor(ctx, regionName, zoneName, rackName, ts.torID); err != nil {
		return err
	}

	err := ts.db.ScanBladesInRack(regionName, zoneName, rackName, func(index int64)error {
		_, err := ts.db.DeleteBlade(ctx, regionName, zoneName, rackName, index)
		return err
	})

	_, err = ts.db.DeleteRack(ctx, regionName, zoneName, rackName)
	return err
}

func (ts *DBInventoryTestSuite) TestCreateRegion() {
	assert := ts.Assert()
	require := ts.Require()

	ctx := context.Background()

	region := &pb.Definition_Region{
		Details: &pb.RegionDetails{
			State:    pb.State_in_service,
			Location: "Nowhere in particular",
			Notes:    "empty notes for region",
		},
	}

	revCreate, err := ts.db.CreateRegion(ctx, ts.regionName, region)
	require.NoError(err)
	assert.Less(int64(0), revCreate)

	r, revRead, err := ts.db.ReadRegion(ctx, ts.regionName)
	assert.NoError(err)
	assert.Equal(revCreate, revRead)
	require.NotNil(r)

	assert.Equal(region.Details.State, r.Details.State)
	assert.Equal(region.Details.Location, r.Details.Location)
	assert.Equal(region.Details.Notes, r.Details.Notes)
}

func (ts *DBInventoryTestSuite) TestCreateZone() {
	assert := ts.Assert()
	require := ts.Require()

	ctx := context.Background()

	zone := &pb.Definition_Zone{
		Details: &pb.ZoneDetails{
			Enabled:  true,
			State:    pb.State_in_service,
			Location: "Nowhere in particular",
			Notes:    "empty notes",
		},
	}

	revCreate, err := ts.db.CreateZone(ctx, ts.regionName, ts.zoneName, zone)
	require.NoError(err)
	assert.Less(int64(0), revCreate)

	z, revRead, err := ts.db.ReadZone(ctx, ts.regionName, ts.zoneName)
	assert.NoError(err)
	assert.Equal(revCreate, revRead)
	require.NotNil(z)

	assert.Equal(zone.Details.Enabled, z.Details.Enabled)
	assert.Equal(zone.Details.State, z.Details.State)
	assert.Equal(zone.Details.Location, z.Details.Location)
	assert.Equal(zone.Details.Notes, z.Details.Notes)
}

func (ts *DBInventoryTestSuite) TestCreateRack() {
	assert := ts.Assert()
	require := ts.Require()

	ctx := context.Background()

	rack := &pb.Definition_Rack{
		Details: &pb.RackDetails{
			Enabled:   true,
			Condition: pb.Condition_operational,
			Location:  "In " + ts.zoneName,
			Notes:     "Basic rack for test",
		},
	}

	revCreate, err := ts.db.CreateRack(ctx, ts.regionName, ts.zoneName, ts.rackName, rack)
	require.NoError(err)

	r, revRead, err := ts.db.ReadRack(ctx, ts.regionName, ts.zoneName, ts.rackName)
	assert.NoError(err)
	assert.Equal(revCreate, revRead)
	require.NotNil(r)

	assert.Equal(rack.Details.Enabled, r.Details.Enabled)
	assert.Equal(rack.Details.Condition, r.Details.Condition)
	assert.Equal(rack.Details.Location, r.Details.Location)
	assert.Equal(rack.Details.Notes, r.Details.Notes)
}

func (ts *DBInventoryTestSuite) TestCreatePdu() {
	assert := ts.Assert()
	require := ts.Require()

	ctx := context.Background()

	pduID := int64(1)

	pdu := &pb.Definition_Pdu{
		Details: &pb.PduDetails{
			Enabled:   true,
			Condition: pb.Condition_operational,
		},
		Ports: make(map[int64]*pb.PowerPort),
	}

	pdu.Ports[0] = &pb.PowerPort{
		Wired: false,
	}

	pdu.Ports[1] = &pb.PowerPort{
		Wired: true,
		Item: &pb.Hardware{
			Type: pb.Hardware_unknown,
		},
	}

	pdu.Ports[2] = &pb.PowerPort{
		Wired: true,
		Item: &pb.Hardware{
			Type: pb.Hardware_tor,
			Id:   0,
			Port: 0,
		},
	}

	pdu.Ports[3] = &pb.PowerPort{
		Wired: true,
		Item: &pb.Hardware{
			Type: pb.Hardware_tor,
			Id:   1,
			Port: 0,
		},
	}

	pdu.Ports[4] = &pb.PowerPort{
		Wired: true,
		Item: &pb.Hardware{
			Type: pb.Hardware_blade,
			Id:   0,
			Port: 0,
		},
	}

	pdu.Ports[5] = &pb.PowerPort{
		Wired: true,
		Item: &pb.Hardware{
			Type: pb.Hardware_blade,
			Id:   0,
			Port: 1,
		},
	}

	pdu.Ports[6] = &pb.PowerPort{
		Wired: true,
		Item: &pb.Hardware{
			Type: pb.Hardware_blade,
			Id:   1,
			Port: 0,
		},
	}

	pdu.Ports[7] = &pb.PowerPort{
		Wired: true,
		Item: &pb.Hardware{
			Type: pb.Hardware_blade,
			Id:   1,
			Port: 1,
		},
	}

	revCreate, err := ts.db.CreatePdu(ctx, ts.regionName, ts.zoneName, ts.rackName, pduID, pdu)
	require.NoError(err)

	t, revRead, err := ts.db.ReadPdu(ctx, ts.regionName, ts.zoneName, ts.rackName, pduID)
	assert.NoError(err)
	assert.Equal(revCreate, revRead)
	require.NotNil(t)

	assert.Equal(pdu.Details.Enabled, t.Details.Enabled)
	assert.Equal(pdu.Details.Condition, t.Details.Condition)
	assert.Equal(len(pdu.Ports), len(t.Ports))

	for i, p := range pdu.Ports {
		tp, ok := t.Ports[i]
		assert.True(ok)
		assert.Equal(p.Wired, tp.Wired)

		if p.Item == nil {
			assert.Nil(tp.Item)
		} else {
			require.NotNil(tp.Item)

			assert.Equal(p.Item.Type, tp.Item.Type)
			assert.Equal(p.Item.Id, tp.Item.Id)
			assert.Equal(p.Item.Port, tp.Item.Port)
		}
	}
}

func (ts *DBInventoryTestSuite) TestCreateTor() {
	assert := ts.Assert()
	require := ts.Require()

	ctx := context.Background()

	torID := int64(1)

	tor := &pb.Definition_Tor{
		Details: &pb.TorDetails{
			Enabled:   true,
			Condition: pb.Condition_operational,
		},
		Ports: make(map[int64]*pb.NetworkPort),
	}

	tor.Ports[0] = &pb.NetworkPort{
		Wired: false,
	}

	tor.Ports[1] = &pb.NetworkPort{
		Wired: true,
		Item: &pb.Hardware{
			Type: pb.Hardware_unknown,
		},
	}

	tor.Ports[2] = &pb.NetworkPort{
		Wired: true,
		Item: &pb.Hardware{
			Type: pb.Hardware_pdu,
			Id:   0,
			Port: 0,
		},
	}

	tor.Ports[3] = &pb.NetworkPort{
		Wired: true,
		Item: &pb.Hardware{
			Type: pb.Hardware_pdu,
			Id:   1,
			Port: 0,
		},
	}

	tor.Ports[4] = &pb.NetworkPort{
		Wired: true,
		Item: &pb.Hardware{
			Type: pb.Hardware_blade,
			Id:   0,
			Port: 0,
		},
	}

	tor.Ports[5] = &pb.NetworkPort{
		Wired: true,
		Item: &pb.Hardware{
			Type: pb.Hardware_blade,
			Id:   0,
			Port: 1,
		},
	}

	tor.Ports[6] = &pb.NetworkPort{
		Wired: true,
		Item: &pb.Hardware{
			Type: pb.Hardware_blade,
			Id:   1,
			Port: 0,
		},
	}

	tor.Ports[7] = &pb.NetworkPort{
		Wired: true,
		Item: &pb.Hardware{
			Type: pb.Hardware_blade,
			Id:   1,
			Port: 1,
		},
	}

	revCreate, err := ts.db.CreateTor(ctx, ts.regionName, ts.zoneName, ts.rackName, torID, tor)
	require.NoError(err)

	t, revRead, err := ts.db.ReadTor(ctx, ts.regionName, ts.zoneName, ts.rackName, torID)
	assert.NoError(err)
	assert.Equal(revCreate, revRead)
	require.NotNil(t)

	assert.Equal(tor.Details.Enabled, t.Details.Enabled)
	assert.Equal(tor.Details.Condition, t.Details.Condition)
	assert.Equal(len(tor.Ports), len(t.Ports))

	for i, p := range tor.Ports {
		tp, ok := t.Ports[i]
		assert.True(ok)
		assert.Equal(p.Wired, tp.Wired)

		if p.Item == nil {
			assert.Nil(tp.Item)
		} else {
			require.NotNil(tp.Item)

			assert.Equal(p.Item.Type, tp.Item.Type)
			assert.Equal(p.Item.Id, tp.Item.Id)
			assert.Equal(p.Item.Port, tp.Item.Port)
		}
	}
}

func (ts *DBInventoryTestSuite) TestCreateBlade() {
	assert := ts.Assert()
	require := ts.Require()

	ctx := context.Background()

	bladeID := int64(1)

	blade := &pb.Definition_Blade{
		Details: &pb.BladeDetails{
			Enabled:   true,
			Condition: pb.Condition_operational,
		},
		Capacity: &pb.BladeCapacity{
			Cores:                  8,
			MemoryInMb:             8192,
			DiskInGb:               8192,
			NetworkBandwidthInMbps: 1024,
			Arch:                   "vax",
			Accelerators:           nil,
		},
		BootOnPowerOn: true,
		BootInfo: &pb.BladeBootInfo{
			Source:     pb.BladeBootInfo_local,
			Image:      "test-image.vhdx",
			Version:    "20201225-0000",
			Parameters: "-param1=val1 -param2=val2",
		},
	}

	revCreate, err := ts.db.CreateBlade(ctx, ts.regionName, ts.zoneName, ts.rackName, bladeID, blade)
	require.NoError(err)

	b, revRead, err := ts.db.ReadBlade(ctx, ts.regionName, ts.zoneName, ts.rackName, bladeID)
	assert.NoError(err)
	assert.Equal(revCreate, revRead)
	require.NotNil(b)

	assert.Equal(blade.Details.Enabled, b.Details.Enabled)
	assert.Equal(blade.Details.Condition, b.Details.Condition)

	assert.Equal(blade.Capacity.Cores, b.Capacity.Cores)
	assert.Equal(blade.Capacity.MemoryInMb, b.Capacity.MemoryInMb)
	assert.Equal(blade.Capacity.DiskInGb, b.Capacity.DiskInGb)
	assert.Equal(blade.Capacity.NetworkBandwidthInMbps, b.Capacity.NetworkBandwidthInMbps)
	assert.Equal(blade.Capacity.Arch, b.Capacity.Arch)
	assert.Equal(blade.Capacity.Accelerators, b.Capacity.Accelerators)

	assert.Equal(blade.BootOnPowerOn, b.BootOnPowerOn)

	assert.Equal(blade.BootInfo.Source, b.BootInfo.Source)
	assert.Equal(blade.BootInfo.Image, b.BootInfo.Image)
	assert.Equal(blade.BootInfo.Version, b.BootInfo.Version)
	assert.Equal(blade.BootInfo.Parameters, b.BootInfo.Parameters)
}

func TestDBInventoryTestSuite(t *testing.T) {
	suite.Run(t, new(DBInventoryTestSuite))
}

func (ts *DBInventoryTestSuite) TestScanRegions() {
	assert := ts.Assert()
	require := ts.Require()

	expected := []string{
		"standard",
	}

	actual := make([]string, 0)

	err := ts.db.ScanRegions(func(name string) error {
		actual = append(actual, name)

		return nil
	})

	require.NoError(err)
	assert.ElementsMatch(expected, actual)
}

func (ts *DBInventoryTestSuite) TestGetRegion() {
	assert := ts.Assert()
	require := ts.Require()

	expected := &pb.RegionDetails{
		State: pb.State_in_service,
		Location: "Pacific NW",
		Notes: "Standard Test Region",
	}

	region, err := ts.db.GetRegion(inventory.DefaultRegion)
	require.NoError(err)
	require.NotNil(region)

	assert.Equal(expected.State, region.Details.State)
	assert.Equal(expected.Location, region.Details.Location)
	assert.Equal(expected.Notes, region.Details.Notes)
}


func (ts *DBInventoryTestSuite) TestScanZonesInRegion() {
	assert := ts.Assert()
	require := ts.Require()

	expected := []string{
		inventory.DefaultRegion,
	}

	actual := make([]string, 0)

	err := ts.db.ScanZonesInRegion(inventory.DefaultRegion, func(name string) error {
		actual = append(actual, name)

		return nil
	})

	require.NoError(err)
	assert.ElementsMatch(expected, actual)
}

func (ts *DBInventoryTestSuite) TestGetZone() {
	assert := ts.Assert()
	require := ts.Require()

	expected := &pb.ZoneDetails{
		Enabled: true,
		State: pb.State_in_service,
		Location: "Pacific NW, standard Zone",
		Notes: "Standard Test Zone definition",
	}

	zone, err := ts.db.GetZone(inventory.DefaultRegion, inventory.DefaultZone)
	require.NoError(err)
	require.NotNil(zone)

	actual := zone.GetDetails()

	assert.Equal(expected.Enabled, actual.Enabled)
	assert.Equal(expected.State, actual.State)
	assert.Equal(expected.Location, actual.Location)
	assert.Equal(expected.Notes, actual.Notes)
}

func (ts *DBInventoryTestSuite) TestScanRacksInZone() {
	assert := ts.Assert()
	require := ts.Require()

	expected := []string{
		"rack1",
		"rack2",
		"rack3",
		"rack4",
		"rack5",
		"rack6",
		"rack7",
		"rack8",
	}

	actual := make([]string, 0)

	err := ts.db.ScanRacksInZone(inventory.DefaultRegion, inventory.DefaultZone, func(name string) error {
		actual = append(actual, name)

		return nil
	})

	require.NoError(err)
	assert.ElementsMatch(expected, actual)
}

func (ts *DBInventoryTestSuite) TestScanBladesInRack() {
	assert := ts.Assert()
	require := ts.Require()

	expected := []int64{
		1,
		2,
		3,
		4,
		5,
		6,
		7,
		8,
	}

	actual := make([]int64, 0)

	err := ts.db.ScanBladesInRack(inventory.DefaultRegion, inventory.DefaultZone, "rack1", func(index int64) error {
		actual = append(actual, index)

		return nil
	})

	require.NoError(err)
	assert.ElementsMatch(expected, actual)
}



func (ts *DBInventoryTestSuite) equalHwItem(expected *pb.Hardware, actual *pb.Hardware) bool {
	assert := ts.Assert()

	if !assert.Equal(expected.Type, actual.Type) ||
	!assert.Equal(expected.Id, actual.Id) ||
	!assert.Equal(expected.Port, actual.Port) {
		return false
	}

	return true
}

func (ts *DBInventoryTestSuite) equalPowerPort(expected *pb.PowerPort, actual *pb.PowerPort) bool {
	assert := ts.Assert()

	if !assert.Equal(expected.Wired, actual.Wired) || !assert.True(ts.equalHwItem(expected.Item, actual.Item)) {
		return false
	}

	return true
}

func (ts *DBInventoryTestSuite) equalNetworkPort(expected *pb.NetworkPort, actual *pb.NetworkPort) bool {
	return expected.Wired == actual.Wired && ts.equalHwItem(expected.Item, actual.Item)
}

func (ts *DBInventoryTestSuite) equalRegionDetails(expected *pb.RegionDetails, actual *pb.RegionDetails) bool {
	assert := ts.Assert()

	if !assert.Equal(expected.State, actual.State) ||
	!assert.Equal(expected.Location, actual.Location) ||
	!assert.Equal(expected.Notes, actual.Notes) {
		return false
	}

	return true
}

func (ts *DBInventoryTestSuite) equalZoneDetails(expected *pb.ZoneDetails, actual *pb.ZoneDetails) bool {
	assert := ts.Assert()

	if !assert.Equal(expected.Enabled, actual.Enabled) ||
	!assert.Equal(expected.State, actual.State) ||
	!assert.Equal(expected.Location, actual.Location) ||
	!assert.Equal(expected.Notes, actual.Notes) {
		return false
	}

	return true
}

func (ts *DBInventoryTestSuite) equalRackDetails(expected *pb.RackDetails, actual *pb.RackDetails) bool {
	assert := ts.Assert()

	if !assert.Equal(expected.Enabled, actual.Enabled) ||
	!assert.Equal(expected.Condition, actual.Condition) ||
	!assert.Equal(expected.Location, actual.Location) ||
	!assert.Equal(expected.Notes, actual.Notes) {
		return false
	}

	return true
}

func (ts *DBInventoryTestSuite) equalPduDetails(expected *pb.PduDetails, actual *pb.PduDetails) bool {
	assert := ts.Assert()

	if !assert.Equal(expected.Condition, actual.Condition) || !assert.Equal(expected.Enabled, actual.Enabled) {
		return false
	}

	return true
}

func (ts *DBInventoryTestSuite) equalTorDetails(expected *pb.TorDetails, actual *pb.TorDetails) bool {
	assert := ts.Assert()

	if !assert.Equal(expected.Condition, actual.Condition) || !assert.Equal(expected.Enabled, actual.Enabled) {
		return false
	}

	return true
}

func (ts *DBInventoryTestSuite) equalBladeDetails(expected *pb.BladeDetails, actual *pb.BladeDetails) bool {
	assert := ts.Assert()

	if !assert.Equal(expected.Condition, actual.Condition) || !assert.Equal(expected.Enabled, actual.Enabled) {
		return false
	}

	return true
}

func (ts *DBInventoryTestSuite) equalPdu(expected *pb.Definition_Pdu, actual *pb.Definition_Pdu) bool {
	assert := ts.Assert()

	if !assert.True(ts.equalPduDetails(expected.Details, actual.Details)) || !assert.Equal(len(expected.Ports), len(actual.Ports)) {
		return false
	}

	for i, v := range expected.Ports {
		if !assert.True(ts.equalPowerPort(v, actual.Ports[i])) {
			return false
		}
	}

	return true
}

func (ts *DBInventoryTestSuite) equalTor(expected *pb.Definition_Tor, actual *pb.Definition_Tor) bool {
	assert := ts.Assert()

	if !assert.True(ts.equalTorDetails(expected.Details, actual.Details)) || !assert.Equal(len(expected.Ports), len(actual.Ports)) {
		return false
	}

	for i, v := range expected.Ports {
		if !assert.True(ts.equalNetworkPort(v, actual.Ports[i])) {
			return false
		}
	}

	return true
}

func (ts *DBInventoryTestSuite) equalBladeBootInfo(expected *pb.BladeBootInfo, actual *pb.BladeBootInfo) bool {
	assert := ts.Assert()

	if !assert.Equal(expected.Source, actual.Source) ||
	!assert.Equal(expected.Image, actual.Image) ||
	!assert.Equal(expected.Version, actual.Version) ||
	!assert.Equal(expected.Parameters, actual.Parameters) {
		return false
	}

	return true
}

func (ts *DBInventoryTestSuite) equalBladeCapacity(expected *pb.BladeCapacity, actual *pb.BladeCapacity) bool {
	assert := ts.Assert()

	if !assert.Equal(expected.Cores, actual.Cores) ||
	!assert.Equal(expected.MemoryInMb, actual.MemoryInMb) ||
	!assert.Equal(expected.DiskInGb, actual.DiskInGb) ||
	!assert.Equal(expected.Arch, actual.Arch) ||
	!assert.Equal(expected.NetworkBandwidthInMbps, actual.NetworkBandwidthInMbps) {
		return false
	}
	return true
}

func (ts *DBInventoryTestSuite) equalBlade(expected *pb.Definition_Blade, actual *pb.Definition_Blade) bool {
	assert := ts.Assert()

	if !assert.True(ts.equalBladeDetails(expected.Details, actual.Details)) ||
	!assert.True(ts.equalBladeCapacity(expected.Capacity, actual.Capacity)) ||
	!assert.True(ts.equalBladeBootInfo(expected.BootInfo, actual.BootInfo)) ||
	!assert.Equal(expected.BootOnPowerOn, actual.BootOnPowerOn) {
		return false
	}

	return true
}

func (ts *DBInventoryTestSuite) equalRack(expected *pb.Definition_Rack, actual *pb.Definition_Rack) bool {
	assert := ts.Assert()

	if !assert.True(ts.equalRackDetails(expected.Details, actual.Details)) ||
	!assert.Equal(len(expected.Pdus), len(actual.Pdus)) ||
	!assert.Equal(len(expected.Tors), len(actual.Tors)) ||
	!assert.Equal(len(expected.Blades), len(actual.Blades)) {
		return false
	}

	for i, v := range expected.Pdus {
		if !assert.True(ts.equalPdu(v, actual.Pdus[i])) {
			return false
		}
	}

	for i, v := range expected.Tors {
		if !assert.True(ts.equalTor(v, actual.Tors[i])) {
			return false
		}
	}

	for i, v := range expected.Blades {
		if !assert.True(ts.equalBlade(v, actual.Blades[i])) {
			return false
		}
	}

	return true
}

func (ts *DBInventoryTestSuite) TestGetRackInZone() {
	assert := ts.Assert()
	require := ts.Require()

	expectedDetails := &pb.RackDetails{
		Enabled: true,
		Condition: pb.Condition_operational,
		Location: "Pacific NW, rack 1",
		Notes: "rack definition, 1 pdu, 1 tor, 8 blades",
	}

	expectedPdus := map[int64]*pb.Definition_Pdu{
		int64(0) : {
			Details: &pb.PduDetails{Enabled: true, Condition: pb.Condition_operational},
			Ports: map[int64]*pb.PowerPort{
				int64(0): {Wired: true, Item: &pb.Hardware{Type: pb.Hardware_tor,   Id: int64(0), Port: int64(0)}},
				int64(1): {Wired: true, Item: &pb.Hardware{Type: pb.Hardware_blade, Id: int64(1), Port: int64(0)}},
				int64(2): {Wired: true, Item: &pb.Hardware{Type: pb.Hardware_blade, Id: int64(2), Port: int64(0)}},
				int64(3): {Wired: true, Item: &pb.Hardware{Type: pb.Hardware_blade, Id: int64(3), Port: int64(0)}},
				int64(4): {Wired: true, Item: &pb.Hardware{Type: pb.Hardware_blade, Id: int64(4), Port: int64(0)}},
				int64(5): {Wired: true, Item: &pb.Hardware{Type: pb.Hardware_blade, Id: int64(5), Port: int64(0)}},
				int64(6): {Wired: true, Item: &pb.Hardware{Type: pb.Hardware_blade, Id: int64(6), Port: int64(0)}},
				int64(7): {Wired: true, Item: &pb.Hardware{Type: pb.Hardware_blade, Id: int64(7), Port: int64(0)}},
				int64(8): {Wired: true, Item: &pb.Hardware{Type: pb.Hardware_blade, Id: int64(8), Port: int64(0)}},
			},
		},
	}

	expectedTors := map[int64]*pb.Definition_Tor{
		int64(0) : {
			Details: &pb.TorDetails{Enabled: true, Condition: pb.Condition_operational},
			Ports: map[int64]*pb.NetworkPort{
				int64(0): {Wired: true, Item: &pb.Hardware{Type: pb.Hardware_pdu,   Id: int64(0), Port: int64(0)}},
				int64(1): {Wired: true, Item: &pb.Hardware{Type: pb.Hardware_blade, Id: int64(1), Port: int64(0)}},
				int64(2): {Wired: true, Item: &pb.Hardware{Type: pb.Hardware_blade, Id: int64(2), Port: int64(0)}},
				int64(3): {Wired: true, Item: &pb.Hardware{Type: pb.Hardware_blade, Id: int64(3), Port: int64(0)}},
				int64(4): {Wired: true, Item: &pb.Hardware{Type: pb.Hardware_blade, Id: int64(4), Port: int64(0)}},
				int64(5): {Wired: true, Item: &pb.Hardware{Type: pb.Hardware_blade, Id: int64(5), Port: int64(0)}},
				int64(6): {Wired: true, Item: &pb.Hardware{Type: pb.Hardware_blade, Id: int64(6), Port: int64(0)}},
				int64(7): {Wired: true, Item: &pb.Hardware{Type: pb.Hardware_blade, Id: int64(7), Port: int64(0)}},
				int64(8): {Wired: true, Item: &pb.Hardware{Type: pb.Hardware_blade, Id: int64(8), Port: int64(0)}},
			},
		},
	}

	expectedBlades := map[int64]*pb.Definition_Blade{
		int64(1): {
			Details:       &pb.BladeDetails{Enabled: true, Condition: pb.Condition_operational},
			Capacity:      &pb.BladeCapacity{Cores: 16, MemoryInMb: 16384, DiskInGb: 240, NetworkBandwidthInMbps: 2048, Arch: "X64"},
			BootInfo:      &pb.BladeBootInfo{Source: pb.BladeBootInfo_network, Image: "standard.vhdx", Version: "latest", Parameters: "-version=1 -node=R1Z1R1B1"},
			BootOnPowerOn: true,
		},
		int64(2): {
			Details:       &pb.BladeDetails{Enabled: true, Condition: pb.Condition_operational},
			Capacity:      &pb.BladeCapacity{Cores: 32, MemoryInMb: 16384, DiskInGb: 120, NetworkBandwidthInMbps: 2048, Arch: "X64"},
			BootInfo:      &pb.BladeBootInfo{Source: pb.BladeBootInfo_network, Image: "standard.vhdx", Version: "latest", Parameters: "-version=1 -node=R1Z1R1B2"},
			BootOnPowerOn: true,
		},
		int64(3): {
			Details:       &pb.BladeDetails{Enabled: true, Condition: pb.Condition_operational},
			Capacity:      &pb.BladeCapacity{Cores: 32, MemoryInMb: 16384, DiskInGb: 120, NetworkBandwidthInMbps: 2048, Arch: "X64"},
			BootInfo:      &pb.BladeBootInfo{Source: pb.BladeBootInfo_network, Image: "standard.vhdx", Version: "latest", Parameters: "-version=1 -node=R1Z1R1B3"},
			BootOnPowerOn: true,
		},
		int64(4): {
			Details:       &pb.BladeDetails{Enabled: true, Condition: pb.Condition_operational},
			Capacity:      &pb.BladeCapacity{Cores: 32, MemoryInMb: 16384, DiskInGb: 120, NetworkBandwidthInMbps: 2048, Arch: "X64"},
			BootInfo:      &pb.BladeBootInfo{Source: pb.BladeBootInfo_network, Image: "standard.vhdx", Version: "latest", Parameters: "-version=1 -node=R1Z1R1B4"},
			BootOnPowerOn: true,
		},
		int64(5): {
			Details:       &pb.BladeDetails{Enabled: true, Condition: pb.Condition_operational},
			Capacity:      &pb.BladeCapacity{Cores: 32, MemoryInMb: 16384, DiskInGb: 120, NetworkBandwidthInMbps: 2048, Arch: "X64"},
			BootInfo:      &pb.BladeBootInfo{Source: pb.BladeBootInfo_network, Image: "standard.vhdx", Version: "latest", Parameters: "-version=1 -node=R1Z1R1B5"},
			BootOnPowerOn: true,
		},
		int64(6): {
			Details:       &pb.BladeDetails{Enabled: true, Condition: pb.Condition_operational},
			Capacity:      &pb.BladeCapacity{Cores: 32, MemoryInMb: 16384, DiskInGb: 120, NetworkBandwidthInMbps: 2048, Arch: "X64"},
			BootInfo:      &pb.BladeBootInfo{Source: pb.BladeBootInfo_network, Image: "standard.vhdx", Version: "latest", Parameters: "-version=1 -node=R1Z1R1B6"},
			BootOnPowerOn: true,
		},
		int64(7): {
			Details:       &pb.BladeDetails{Enabled: true, Condition: pb.Condition_operational},
			Capacity:      &pb.BladeCapacity{Cores: 32, MemoryInMb: 16384, DiskInGb: 120, NetworkBandwidthInMbps: 2048, Arch: "X64"},
			BootInfo:      &pb.BladeBootInfo{Source: pb.BladeBootInfo_network, Image: "standard.vhdx", Version: "latest", Parameters: "-version=1 -node=R1Z1R1B7"},
			BootOnPowerOn: true,
		},
		int64(8): {
			Details:       &pb.BladeDetails{Enabled: true, Condition: pb.Condition_operational},
			Capacity:      &pb.BladeCapacity{Cores: 32, MemoryInMb: 16384, DiskInGb: 120, NetworkBandwidthInMbps: 2048, Arch: "X64"},
			BootInfo:      &pb.BladeBootInfo{Source: pb.BladeBootInfo_network, Image: "standard.vhdx", Version: "latest", Parameters: "-version=1 -node=R1Z1R1B8"},
			BootOnPowerOn: true,
		},
	}

	expectedRack := &pb.Definition_Rack{
		Details: expectedDetails,
		Pdus:    expectedPdus,
		Tors:    expectedTors,
		Blades:  expectedBlades,
	}

	rack, err := ts.db.GetRackInZone(inventory.DefaultRegion, inventory.DefaultZone, "rack1")
	require.NoError(err)
	require.NotNil(rack)

	require.NotNil(rack.Details)
	require.NotNil(rack.Pdus)
	require.NotNil(rack.Tors)
	require.NotNil(rack.Blades)

	assert.True(ts.equalRack(expectedRack, rack))
}

func (ts *DBInventoryTestSuite) TestGetRack() {
	assert := ts.Assert()
	require := ts.Require()

	rack, err := ts.db.GetRack(inventory.DefaultRegion, inventory.DefaultZone, "rack2")
	require.NoError(err)
	require.NotNil(rack)

	expectedCapacityBlade1 := &pb.BladeCapacity{Cores: 16, MemoryInMb: 16384, DiskInGb: 240, NetworkBandwidthInMbps: 2048, Arch: "X64"}
	expectedCapacity       := &pb.BladeCapacity{Cores: 32, MemoryInMb: 16384, DiskInGb: 120, NetworkBandwidthInMbps: 2048, Arch: "X64"}

	assert.Equal(8, len(rack.Blades))

	for i, v := range rack.Blades {
		if i == 1 {
			assert.True(ts.equalBladeCapacity(expectedCapacityBlade1, v))

		} else {
			assert.True(ts.equalBladeCapacity(expectedCapacity, v))
		}
	}
}

func (ts *DBInventoryTestSuite) TestUpdateRegion() {
	assert := ts.Assert()
	require := ts.Require()

	ctx := context.Background()

	region1 := &pb.Definition_Region{
		Details: &pb.RegionDetails{
			State:    pb.State_in_service,
			Location: "Nowhere in particular",
			Notes:    "empty notes",
		},
	}

	region2 := &pb.Definition_Region{
		Details: &pb.RegionDetails{
			State:    pb.State_decommissioning,
			Location: "Nowhere in particular - still",
			Notes:    "About to be removed",
		},
	}

	revCreate, err := ts.db.CreateRegion(ctx, ts.regionName, region1)
	require.NoError(err)
	assert.Less(int64(0), revCreate)

	r1, revRead1, err := ts.db.ReadRegion(ctx, ts.regionName)
	assert.NoError(err)
	assert.Equal(revCreate, revRead1)
	require.NotNil(r1)

	assert.True(ts.equalRegionDetails(region1.Details, r1.Details))

	revUpdate, err := ts.db.UpdateRegion(ctx, ts.regionName, region2)
	require.NoError(err)
	assert.Less(revCreate, revUpdate)

	r2, revRead2, err := ts.db.ReadRegion(ctx, ts.regionName)
	assert.NoError(err)
	assert.Equal(revUpdate, revRead2)
	require.NotNil(r1)

	assert.True(ts.equalRegionDetails(region2.Details, r2.Details))
}

func (ts *DBInventoryTestSuite) TestUpdateZone() {
	assert := ts.Assert()
	require := ts.Require()

	ctx := context.Background()

	zone1 := &pb.Definition_Zone{
		Details: &pb.ZoneDetails{
			Enabled:  true,
			State:    pb.State_in_service,
			Location: "Nowhere in particular",
			Notes:    "empty notes",
		},
	}

	zone2 := &pb.Definition_Zone{
		Details: &pb.ZoneDetails{
			Enabled:  false,
			State:    pb.State_decommissioning,
			Location: "Nowhere in particular - still",
			Notes:    "About to be removed",
		},
	}

	revCreate, err := ts.db.CreateZone(ctx, ts.regionName, ts.zoneName, zone1)
	require.NoError(err)
	assert.Less(int64(0), revCreate)

	z1, revRead1, err := ts.db.ReadZone(ctx, ts.regionName, ts.zoneName)
	assert.NoError(err)
	assert.Equal(revCreate, revRead1)
	require.NotNil(z1)

	assert.True(ts.equalZoneDetails(zone1.Details, z1.Details))

	revUpdate, err := ts.db.UpdateZone(ctx, ts.regionName, ts.zoneName, zone2)
	require.NoError(err)
	assert.Less(revCreate, revUpdate)

	z2, revRead2, err := ts.db.ReadZone(ctx, ts.regionName, ts.zoneName)
	assert.NoError(err)
	assert.Equal(revUpdate, revRead2)
	require.NotNil(z1)

	assert.True(ts.equalZoneDetails(zone2.Details, z2.Details))
}

func (ts *DBInventoryTestSuite) TestUpdateRack() {
	assert := ts.Assert()
	require := ts.Require()

	ctx := context.Background()

	rackName := "rackUpdatePdu"

	rack1 := &pb.Definition_Rack{
		Details: &pb.RackDetails{
			Enabled:   true,
			Condition: pb.Condition_operational,
			Location:  "Nowhere in particular",
			Notes:     "empty notes",
		},
	}

	rack2 := &pb.Definition_Rack{
		Details: &pb.RackDetails{
			Enabled:   false,
			Condition: pb.Condition_out_for_repair,
			Location:  "Nowhere in particular - still",
			Notes:     "being repaired",
		},
	}

	revCreate, err := ts.db.CreateRack(ctx, ts.regionName, ts.zoneName, rackName, rack1)
	require.NoError(err)
	assert.Less(int64(0), revCreate)

	r1, revRead1, err := ts.db.ReadRack(ctx, ts.regionName, ts.zoneName, rackName)
	assert.NoError(err)
	assert.Equal(revCreate, revRead1)
	require.NotNil(r1)

	assert.True(ts.equalRackDetails(rack1.Details, r1.Details))

	revUpdate, err := ts.db.UpdateRack(ctx, ts.regionName, ts.zoneName, rackName, rack2)
	require.NoError(err)
	assert.Less(revCreate, revUpdate)

	r2, revRead2, err := ts.db.ReadRack(ctx, ts.regionName, ts.zoneName, rackName)
	assert.NoError(err)
	assert.Equal(revUpdate, revRead2)
	require.NotNil(r1)

	assert.True(ts.equalRackDetails(rack2.Details, r2.Details))
}

func (ts *DBInventoryTestSuite) TestUpdatePdu() {
	assert := ts.Assert()
	require := ts.Require()

	ctx := context.Background()

	rackName := "rackUpdatePdu"

	pdu1 := &pb.Definition_Pdu{
		Details: &pb.PduDetails{Enabled: true, Condition: pb.Condition_operational},
		Ports: map[int64]*pb.PowerPort{
			int64(0): {Wired: true, Item: &pb.Hardware{Type: pb.Hardware_tor, Id: int64(0), Port: int64(0)}},
		},
	}

	pdu2 := &pb.Definition_Pdu{
		Details: &pb.PduDetails{Enabled: false,	Condition: pb.Condition_out_for_repair},
		Ports: map[int64]*pb.PowerPort{
			int64(0): {Wired: true, Item: &pb.Hardware{Type: pb.Hardware_tor,   Id: int64(0), Port: int64(0)}},
		},
	}

	revCreate, err := ts.db.CreatePdu(ctx, ts.regionName, ts.zoneName, rackName, ts.pduID, pdu1)
	require.NoError(err)
	assert.Less(int64(0), revCreate)

	p1, revRead1, err := ts.db.ReadPdu(ctx, ts.regionName, ts.zoneName, rackName,ts.pduID)
	assert.NoError(err)
	assert.Equal(revCreate, revRead1)
	require.NotNil(p1)

	assert.True(ts.equalPdu(pdu1, p1))

	revUpdate, err := ts.db.UpdatePdu(ctx, ts.regionName, ts.zoneName, rackName, ts.pduID, pdu2)
	require.NoError(err)
	assert.Less(revCreate, revUpdate)

	p2, revRead2, err := ts.db.ReadPdu(ctx, ts.regionName, ts.zoneName, rackName, ts.pduID)
	assert.NoError(err)
	assert.Equal(revUpdate, revRead2)
	require.NotNil(p1)

	assert.True(ts.equalPdu(pdu2, p2))
}

func (ts *DBInventoryTestSuite) TestUpdateTor() {
	assert := ts.Assert()
	require := ts.Require()

	ctx := context.Background()

	rackName := "rackUpdateTor"

	tor1 := &pb.Definition_Tor{
		Details: &pb.TorDetails{Enabled: true, Condition: pb.Condition_operational},
		Ports: map[int64]*pb.NetworkPort{
			int64(0): {Wired: true, Item: &pb.Hardware{Type: pb.Hardware_pdu, Id: int64(0), Port: int64(0)}},
		},
	}

	tor2 := &pb.Definition_Tor{
		Details: &pb.TorDetails{Enabled: true, Condition: pb.Condition_retiring},
		Ports: map[int64]*pb.NetworkPort{
			int64(0): {Wired: true, Item: &pb.Hardware{Type: pb.Hardware_pdu, Id: int64(0), Port: int64(0)}},
		},
	}

	revCreate, err := ts.db.CreateTor(ctx, ts.regionName, ts.zoneName, rackName, ts.torID, tor1)
	require.NoError(err)
	assert.Less(int64(0), revCreate)

	t1, revRead1, err := ts.db.ReadTor(ctx, ts.regionName, ts.zoneName, rackName, ts.torID)
	assert.NoError(err)
	assert.Equal(revCreate, revRead1)
	require.NotNil(t1)

	assert.True(ts.equalTor(tor1, t1))

	revUpdate, err := ts.db.UpdateTor(ctx, ts.regionName, ts.zoneName, rackName, ts.torID, tor2)
	require.NoError(err)
	assert.Less(revCreate, revUpdate)

	t2, revRead2, err := ts.db.ReadTor(ctx, ts.regionName, ts.zoneName, rackName, ts.torID)
	assert.NoError(err)
	assert.Equal(revUpdate, revRead2)
	require.NotNil(t1)

	assert.True(ts.equalTor(tor2, t2))
}

func (ts *DBInventoryTestSuite) TestUpdateBlade() {
	assert := ts.Assert()
	require := ts.Require()

	ctx := context.Background()

	rackName := "rackUpdateBlade"

	blade1 := &pb.Definition_Blade{
		Details:       &pb.BladeDetails{Enabled: true, Condition: pb.Condition_operational},
		Capacity:      &pb.BladeCapacity{Cores: 16, MemoryInMb: 16384, DiskInGb: 240, NetworkBandwidthInMbps: 2048, Arch: "X64"},
		BootInfo:      &pb.BladeBootInfo{Source: pb.BladeBootInfo_network, Image: "standard.vhdx", Version: "latest", Parameters: "-version=1 -node=R1Z1R1B1"},
		BootOnPowerOn: true,
}

	blade2 := &pb.Definition_Blade{
		Details:       &pb.BladeDetails{Enabled: true, Condition: pb.Condition_retiring},
		Capacity:      &pb.BladeCapacity{Cores: 16, MemoryInMb: 16384, DiskInGb: 240, NetworkBandwidthInMbps: 2048, Arch: "X64"},
		BootInfo:      &pb.BladeBootInfo{Source: pb.BladeBootInfo_network, Image: "standard.vhdx", Version: "latest", Parameters: "-version=1 -node=R1Z1R1B1"},
		BootOnPowerOn: true,
	}

	revCreate, err := ts.db.CreateBlade(ctx, ts.regionName, ts.zoneName, rackName, ts.bladeID, blade1)
	require.NoError(err)
	assert.Less(int64(0), revCreate)

	b1, revRead1, err := ts.db.ReadBlade(ctx, ts.regionName, ts.zoneName, rackName, ts.bladeID)
	assert.NoError(err)
	assert.Equal(revCreate, revRead1)
	require.NotNil(b1)

	assert.True(ts.equalBlade(blade1, b1))

	revUpdate, err := ts.db.UpdateBlade(ctx, ts.regionName, ts.zoneName, rackName, ts.bladeID, blade2)
	require.NoError(err)
	assert.Less(revCreate, revUpdate)

	b2, revRead2, err := ts.db.ReadBlade(ctx, ts.regionName, ts.zoneName, rackName, ts.bladeID)
	assert.NoError(err)
	assert.Equal(revUpdate, revRead2)
	require.NotNil(b1)

	assert.True(ts.equalBlade(blade2, b2))
}

func (ts *DBInventoryTestSuite) TestDeleteRegion() {
	assert := ts.Assert()
	require := ts.Require()

	ctx := context.Background()

	regionName := "regionDeleteRegion"

	region1 := &pb.Definition_Region{
		Details: &pb.RegionDetails{
			State:    pb.State_in_service,
			Location: "Nowhere in particular",
			Notes:    "empty notes",
		},
	}

	revCreate, err := ts.db.CreateRegion(ctx, regionName, region1)
	require.NoError(err)
	assert.Less(int64(0), revCreate)

	r1, revRead1, err := ts.db.ReadRegion(ctx, regionName)
	assert.NoError(err)
	assert.Equal(revCreate, revRead1)
	require.NotNil(r1)

	assert.True(ts.equalRegionDetails(region1.Details, r1.Details))

	revDelete, err := ts.db.DeleteRegion(ctx, regionName)
	require.NoError(err)
	assert.Less(revCreate, revDelete)

	r2, revRead2, err := ts.db.ReadRegion(ctx, regionName)
	require.Equal(err, errors.ErrRegionNotFound{Region: regionName})
	assert.Equal(int64(InvalidRev), revRead2)
	assert.Nil(r2)
}

func (ts *DBInventoryTestSuite) TestDeleteZone() {
	assert := ts.Assert()
	require := ts.Require()

	ctx := context.Background()

	zoneName := "zoneDeleteZone"

	zone1 := &pb.Definition_Zone{
		Details: &pb.ZoneDetails{
			Enabled:  true,
			State:    pb.State_in_service,
			Location: "Nowhere in particular",
			Notes:    "empty notes",
		},
	}

	revCreate, err := ts.db.CreateZone(ctx, ts.regionName, zoneName, zone1)
	require.NoError(err)
	assert.Less(int64(0), revCreate)

	z1, revRead1, err := ts.db.ReadZone(ctx, ts.regionName, zoneName)
	assert.NoError(err)
	assert.Equal(revCreate, revRead1)
	require.NotNil(z1)

	assert.True(ts.equalZoneDetails(zone1.Details, z1.Details))

	revDelete, err := ts.db.DeleteZone(ctx, ts.regionName, zoneName)
	require.NoError(err)
	assert.Less(revCreate, revDelete)

	z2, revRead2, err := ts.db.ReadZone(ctx, ts.regionName, zoneName)
	require.Equal(err, errors.ErrZoneNotFound{Region: ts.regionName, Zone: zoneName})
	assert.Equal(int64(InvalidRev), revRead2)
	assert.Nil(z2)
}

func (ts *DBInventoryTestSuite) TestDeleteRack() {
	assert := ts.Assert()
	require := ts.Require()

	ctx := context.Background()

	rackName := "rackDeleteRack"

	rack := &pb.Definition_Rack{
		Details: &pb.RackDetails{
			Enabled:   true,
			Condition: pb.Condition_operational,
			Location:  "Nowhere in particular",
			Notes:     "empty notes",
		},
	}

	revCreate, err := ts.db.CreateRack(ctx, ts.regionName, ts.zoneName, rackName, rack)
	require.NoError(err)
	assert.Less(int64(0), revCreate)

	r1, revRead1, err := ts.db.ReadRack(ctx, ts.regionName, ts.zoneName, rackName)
	assert.NoError(err)
	assert.Equal(revCreate, revRead1)
	require.NotNil(r1)

	assert.True(ts.equalRackDetails(rack.Details, r1.Details))

	revDelete, err := ts.db.DeleteRack(ctx, ts.regionName, ts.zoneName, rackName)
	require.NoError(err)
	assert.Less(revCreate, revDelete)

	r2, revRead2, err := ts.db.ReadRack(ctx, ts.regionName, ts.zoneName, rackName)
	require.Equal(err, errors.ErrRackNotFound{Region: ts.regionName, Zone: ts.zoneName, Rack: rackName})
	assert.Equal(int64(InvalidRev), revRead2)
	assert.Nil(r2)
}

func (ts *DBInventoryTestSuite) TestDeletePdu() {
	assert := ts.Assert()
	require := ts.Require()

	ctx := context.Background()

	rackName := "rackDeletePdu"

	pdu := &pb.Definition_Pdu{
		Details: &pb.PduDetails{Enabled: true, Condition: pb.Condition_operational},
		Ports: map[int64]*pb.PowerPort{
			int64(0): {Wired: true, Item: &pb.Hardware{Type: pb.Hardware_tor, Id: int64(0), Port: int64(0)}},
		},
	}

	revCreate, err := ts.db.CreatePdu(ctx, ts.regionName, ts.zoneName, rackName, ts.pduID, pdu)
	require.NoError(err)
	assert.Less(int64(0), revCreate)

	p1, revRead1, err := ts.db.ReadPdu(ctx, ts.regionName, ts.zoneName, rackName, ts.pduID)
	assert.NoError(err)
	assert.Equal(revCreate, revRead1)
	require.NotNil(p1)

	assert.True(ts.equalPduDetails(pdu.Details, p1.Details))

	revDelete, err := ts.db.DeletePdu(ctx, ts.regionName, ts.zoneName, rackName, ts.pduID)
	require.NoError(err)
	assert.Less(revCreate, revDelete)

	p2, revRead2, err := ts.db.ReadPdu(ctx, ts.regionName, ts.zoneName, rackName, ts.pduID)
	require.Equal(err, errors.ErrPduNotFound{Region: ts.regionName, Zone: ts.zoneName, Rack: rackName, Pdu: ts.pduID})
	assert.Equal(int64(InvalidRev), revRead2)
	assert.Nil(p2)
}

func (ts *DBInventoryTestSuite) TestDeleteTor() {
	assert := ts.Assert()
	require := ts.Require()

	ctx := context.Background()

	rackName := "rackDeleteTor"

	tor := &pb.Definition_Tor{
		Details: &pb.TorDetails{Enabled: true, Condition: pb.Condition_operational},
		Ports: map[int64]*pb.NetworkPort{
			int64(0): {Wired: true, Item: &pb.Hardware{Type: pb.Hardware_pdu, Id: int64(0), Port: int64(0)}},
		},
	}

	revCreate, err := ts.db.CreateTor(ctx, ts.regionName, ts.zoneName, rackName, ts.torID, tor)
	require.NoError(err)
	assert.Less(int64(0), revCreate)

	t1, revRead1, err := ts.db.ReadTor(ctx, ts.regionName, ts.zoneName, rackName, ts.torID)
	assert.NoError(err)
	assert.Equal(revCreate, revRead1)
	require.NotNil(t1)

	assert.True(ts.equalTorDetails(tor.Details, t1.Details))

	revDelete, err := ts.db.DeleteTor(ctx, ts.regionName, ts.zoneName, rackName, ts.torID)
	require.NoError(err)
	assert.Less(revCreate, revDelete)

	t2, revRead2, err := ts.db.ReadTor(ctx, ts.regionName, ts.zoneName, rackName, ts.torID)
	require.Equal(err, errors.ErrTorNotFound{Region: ts.regionName, Zone: ts.zoneName, Rack: rackName, Tor: ts.torID})
	assert.Equal(int64(InvalidRev), revRead2)
	assert.Nil(t2)
}

func (ts *DBInventoryTestSuite) TestDeleteBlade() {
	assert := ts.Assert()
	require := ts.Require()

	ctx := context.Background()

	rackName := "rackUpdateBlade"

	blade1 := &pb.Definition_Blade{
		Details:       &pb.BladeDetails{Enabled: true, Condition: pb.Condition_operational},
		Capacity:      &pb.BladeCapacity{Cores: 16, MemoryInMb: 16384, DiskInGb: 240, NetworkBandwidthInMbps: 2048, Arch: "X64"},
		BootInfo:      &pb.BladeBootInfo{Source: pb.BladeBootInfo_network, Image: "standard.vhdx", Version: "latest", Parameters: "-version=1 -node=R1Z1R1B1"},
		BootOnPowerOn: true,
	}

	revCreate, err := ts.db.CreateBlade(ctx, ts.regionName, ts.zoneName, rackName, ts.bladeID, blade1)
	require.NoError(err)
	assert.Less(int64(0), revCreate)

	b1, revRead1, err := ts.db.ReadBlade(ctx, ts.regionName, ts.zoneName, rackName, ts.bladeID)
	assert.NoError(err)
	assert.Equal(revCreate, revRead1)
	require.NotNil(b1)

	assert.True(ts.equalBlade(blade1, b1))

	revDelete, err := ts.db.DeleteBlade(ctx, ts.regionName, ts.zoneName, rackName, ts.bladeID)
	require.NoError(err)
	assert.Less(revCreate, revDelete)

	b2, revRead2, err := ts.db.ReadBlade(ctx, ts.regionName, ts.zoneName, rackName, ts.bladeID)
	require.Equal(err, errors.ErrBladeNotFound{Region: ts.regionName, Zone: ts.zoneName, Rack: rackName, Blade: ts.bladeID})
	assert.Equal(int64(InvalidRev), revRead2)
	assert.Nil(b2)
}


