package frontend

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Jim3Things/CloudChamber/simulation/internal/clients/inventory"
	"github.com/Jim3Things/CloudChamber/simulation/internal/clients/timestamp"
	"github.com/Jim3Things/CloudChamber/simulation/internal/tracing"
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
	ctx, span := tracing.StartSpan(
		context.Background(),
		tracing.WithContextValue(timestamp.OutsideTime))
	defer span.End()

	err := ts.db.ScanZonesInRegion(ctx, regionName, func(zoneName string) error {
		err := ts.cleanZone(regionName, zoneName)
		if err != nil {
			return err
		}

		_, err = ts.db.DeleteZone(ctx, regionName, zoneName)
		return err
	})
	if err != nil {
		return err
	}

	_, err = ts.db.DeleteRegion(ctx, regionName)

	return err
}

func (ts *DBInventoryTestSuite) cleanZone(regionName string, zoneName string) error {
	ctx, span := tracing.StartSpan(
		context.Background(),
		tracing.WithContextValue(timestamp.OutsideTime))
	defer span.End()

	err := ts.db.ScanRacksInZone(ctx, regionName, zoneName, func(rackName string) error {
		if err := ts.cleanRack(regionName, zoneName, rackName); err != nil {
			return err
		}

		_, err := ts.db.DeleteRack(ctx, regionName, zoneName, rackName)
		return err
	})
	if err != nil {
		return err
	}

	_, err = ts.db.DeleteZone(ctx, regionName, zoneName)
	return err
}

func (ts *DBInventoryTestSuite) cleanRack(regionName string, zoneName string, rackName string) error {
	ctx, span := tracing.StartSpan(
		context.Background(),
		tracing.WithContextValue(timestamp.OutsideTime))
	defer span.End()

	if _, err := ts.db.DeletePdu(ctx, regionName, zoneName, rackName, ts.pduID); err != nil {
		return err
	}

	if _, err := ts.db.DeleteTor(ctx, regionName, zoneName, rackName, ts.torID); err != nil {
		return err
	}

	err := ts.db.ScanBladesInRack(ctx, regionName, zoneName, rackName, func(index int64) error {
		_, err := ts.db.DeleteBlade(ctx, regionName, zoneName, rackName, index)
		return err
	})
	if err != nil {
		return err
	}

	_, err = ts.db.DeleteRack(ctx, regionName, zoneName, rackName)
	return err
}

func (ts *DBInventoryTestSuite) TestCreateRegion() {
	require := ts.Require()

	ctx, span := tracing.StartSpan(
		context.Background(),
		tracing.WithContextValue(timestamp.OutsideTime))
	defer span.End()

	region := &pb.Definition_Region{
		Details: &pb.RegionDetails{
			State:    pb.State_in_service,
			Location: "Nowhere in particular",
			Notes:    "empty notes for region",
		},
	}

	revCreate, err := ts.db.CreateRegion(ctx, ts.regionName, region)
	require.NoError(err)
	require.Less(int64(0), revCreate)

	r, revRead, err := ts.db.ReadRegion(ctx, ts.regionName)
	require.NoError(err)
	require.Equal(revCreate, revRead)
	require.NotNil(r)

	require.Equal(region.Details, r.Details)
}

func (ts *DBInventoryTestSuite) TestCreateZone() {
	require := ts.Require()

	ctx, span := tracing.StartSpan(
		context.Background(),
		tracing.WithContextValue(timestamp.OutsideTime))
	defer span.End()

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
	require.Less(int64(0), revCreate)

	z, revRead, err := ts.db.ReadZone(ctx, ts.regionName, ts.zoneName)
	require.NoError(err)
	require.Equal(revCreate, revRead)
	require.NotNil(z)

	require.Equal(zone.Details, z.Details)
}

func (ts *DBInventoryTestSuite) TestCreateRack() {
	require := ts.Require()

	ctx, span := tracing.StartSpan(
		context.Background(),
		tracing.WithContextValue(timestamp.OutsideTime))
	defer span.End()

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
	require.NoError(err)
	require.Equal(revCreate, revRead)
	require.NotNil(r)

	require.Equal(rack.Details.Enabled, r.Details.Enabled)
	require.Equal(rack.Details.Condition, r.Details.Condition)
	require.Equal(rack.Details.Location, r.Details.Location)
	require.Equal(rack.Details.Notes, r.Details.Notes)
}

func (ts *DBInventoryTestSuite) TestCreatePdu() {
	require := ts.Require()

	ctx, span := tracing.StartSpan(
		context.Background(),
		tracing.WithContextValue(timestamp.OutsideTime))
	defer span.End()

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
	require.NoError(err)
	require.Equal(revCreate, revRead)
	require.NotNil(t)

	require.Equal(pdu.Details.Enabled, t.Details.Enabled)
	require.Equal(pdu.Details.Condition, t.Details.Condition)
	require.Equal(len(pdu.Ports), len(t.Ports))

	for i, p := range pdu.Ports {
		tp, ok := t.Ports[i]
		require.True(ok)
		require.Equal(p.Wired, tp.Wired)

		if p.Item == nil {
			require.Nil(tp.Item)
		} else {
			require.NotNil(tp.Item)

			require.Equal(p.Item.Type, tp.Item.Type)
			require.Equal(p.Item.Id, tp.Item.Id)
			require.Equal(p.Item.Port, tp.Item.Port)
		}
	}
}

func (ts *DBInventoryTestSuite) TestCreateTor() {
	require := ts.Require()

	ctx, span := tracing.StartSpan(
		context.Background(),
		tracing.WithContextValue(timestamp.OutsideTime))
	defer span.End()

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
	require.NoError(err)
	require.Equal(revCreate, revRead)
	require.NotNil(t)

	require.Equal(tor.Details.Enabled, t.Details.Enabled)
	require.Equal(tor.Details.Condition, t.Details.Condition)
	require.Equal(len(tor.Ports), len(t.Ports))

	for i, p := range tor.Ports {
		tp, ok := t.Ports[i]
		require.True(ok)
		require.Equal(p.Wired, tp.Wired)

		if p.Item == nil {
			require.Nil(tp.Item)
		} else {
			require.NotNil(tp.Item)

			require.Equal(p.Item.Type, tp.Item.Type)
			require.Equal(p.Item.Id, tp.Item.Id)
			require.Equal(p.Item.Port, tp.Item.Port)
		}
	}
}

func (ts *DBInventoryTestSuite) TestCreateBlade() {
	require := ts.Require()

	ctx, span := tracing.StartSpan(
		context.Background(),
		tracing.WithContextValue(timestamp.OutsideTime))
	defer span.End()

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
	require.NoError(err)
	require.Equal(revCreate, revRead)
	require.NotNil(b)

	require.Equal(blade.Details.Enabled, b.Details.Enabled)
	require.Equal(blade.Details.Condition, b.Details.Condition)

	require.Equal(blade.Capacity.Cores, b.Capacity.Cores)
	require.Equal(blade.Capacity.MemoryInMb, b.Capacity.MemoryInMb)
	require.Equal(blade.Capacity.DiskInGb, b.Capacity.DiskInGb)
	require.Equal(blade.Capacity.NetworkBandwidthInMbps, b.Capacity.NetworkBandwidthInMbps)
	require.Equal(blade.Capacity.Arch, b.Capacity.Arch)
	require.Equal(blade.Capacity.Accelerators, b.Capacity.Accelerators)

	require.Equal(blade.BootOnPowerOn, b.BootOnPowerOn)

	require.Equal(blade.BootInfo.Source, b.BootInfo.Source)
	require.Equal(blade.BootInfo.Image, b.BootInfo.Image)
	require.Equal(blade.BootInfo.Version, b.BootInfo.Version)
	require.Equal(blade.BootInfo.Parameters, b.BootInfo.Parameters)
}

func TestDBInventoryTestSuite(t *testing.T) {
	suite.Run(t, new(DBInventoryTestSuite))
}

func (ts *DBInventoryTestSuite) TestScanRegions() {
	require := ts.Require()

	expected := []string{
		"standard",
	}

	ctx, span := tracing.StartSpan(
		context.Background(),
		tracing.WithContextValue(timestamp.OutsideTime))
	defer span.End()

	actual := make([]string, 0)

	err := ts.db.ScanRegions(ctx, func(name string) error {
		actual = append(actual, name)

		return nil
	})

	require.NoError(err)
	require.ElementsMatch(expected, actual)
}

func (ts *DBInventoryTestSuite) TestGetRegion() {
	require := ts.Require()

	expected := &pb.RegionDetails{
		State:    pb.State_in_service,
		Location: "Pacific NW",
		Notes:    "Standard Test Region",
	}

	ctx, span := tracing.StartSpan(
		context.Background(),
		tracing.WithContextValue(timestamp.OutsideTime))
	defer span.End()

	region, err := ts.db.GetRegion(ctx, inventory.DefaultRegion)
	require.NoError(err)
	require.NotNil(region)

	require.Equal(expected.State, region.Details.State)
	require.Equal(expected.Location, region.Details.Location)
	require.Equal(expected.Notes, region.Details.Notes)
}

func (ts *DBInventoryTestSuite) TestScanZonesInRegion() {
	require := ts.Require()

	expected := []string{
		inventory.DefaultRegion,
	}

	actual := make([]string, 0)

	ctx, span := tracing.StartSpan(
		context.Background(),
		tracing.WithContextValue(timestamp.OutsideTime))
	defer span.End()

	err := ts.db.ScanZonesInRegion(ctx, inventory.DefaultRegion, func(name string) error {
		actual = append(actual, name)

		return nil
	})

	require.NoError(err)
	require.ElementsMatch(expected, actual)
}

func (ts *DBInventoryTestSuite) TestGetZone() {
	require := ts.Require()

	expected := &pb.ZoneDetails{
		Enabled:  true,
		State:    pb.State_in_service,
		Location: "Pacific NW, standard Zone",
		Notes:    "Standard Test Zone definition",
	}

	ctx, span := tracing.StartSpan(
		context.Background(),
		tracing.WithContextValue(timestamp.OutsideTime))
	defer span.End()

	zone, err := ts.db.GetZone(ctx, inventory.DefaultRegion, inventory.DefaultZone)
	require.NoError(err)
	require.NotNil(zone)

	actual := zone.GetDetails()

	require.Equal(expected.Enabled, actual.Enabled)
	require.Equal(expected.State, actual.State)
	require.Equal(expected.Location, actual.Location)
	require.Equal(expected.Notes, actual.Notes)
}

func (ts *DBInventoryTestSuite) TestScanRacksInZone() {
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
	ctx := context.Background()

	err := ts.db.ScanRacksInZone(ctx, inventory.DefaultRegion, inventory.DefaultZone, func(name string) error {
		actual = append(actual, name)

		return nil
	})

	require.NoError(err)
	require.ElementsMatch(expected, actual)
}

func (ts *DBInventoryTestSuite) TestScanBladesInRack() {
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

	ctx, span := tracing.StartSpan(
		context.Background(),
		tracing.WithContextValue(timestamp.OutsideTime))
	defer span.End()

	err := ts.db.ScanBladesInRack(ctx, inventory.DefaultRegion, inventory.DefaultZone, "rack1", func(index int64) error {
		actual = append(actual, index)

		return nil
	})

	require.NoError(err)
	require.ElementsMatch(expected, actual)
}

func (ts *DBInventoryTestSuite) TestGetRackInZone() {
	require := ts.Require()

	expectedDetails := &pb.RackDetails{
		Enabled:   true,
		Condition: pb.Condition_operational,
		Location:  "Pacific NW, rack 1",
		Notes:     "rack definition, 1 pdu, 1 tor, 8 blades",
	}

	expectedPdus := map[int64]*pb.Definition_Pdu{
		int64(0): {
			Details: &pb.PduDetails{Enabled: true, Condition: pb.Condition_operational},
			Ports: map[int64]*pb.PowerPort{
				int64(0): {Wired: true, Item: &pb.Hardware{Type: pb.Hardware_tor, Id: int64(0), Port: int64(0)}},
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
		int64(0): {
			Details: &pb.TorDetails{Enabled: true, Condition: pb.Condition_operational},
			Ports: map[int64]*pb.NetworkPort{
				int64(0): {Wired: true, Item: &pb.Hardware{Type: pb.Hardware_pdu, Id: int64(0), Port: int64(0)}},
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

	ctx, span := tracing.StartSpan(
		context.Background(),
		tracing.WithContextValue(timestamp.OutsideTime))
	defer span.End()

	rack, err := ts.db.GetRackInZone(ctx, inventory.DefaultRegion, inventory.DefaultZone, "rack1")
	require.NoError(err)
	require.NotNil(rack)

	require.NotNil(rack.Details)
	require.NotNil(rack.Pdus)
	require.NotNil(rack.Tors)
	require.NotNil(rack.Blades)

	require.Equal(expectedRack, rack)
}

func (ts *DBInventoryTestSuite) TestUpdateRegion() {
	require := ts.Require()

	ctx, span := tracing.StartSpan(
		context.Background(),
		tracing.WithContextValue(timestamp.OutsideTime))
	defer span.End()

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
	require.Less(int64(0), revCreate)

	r1, revRead1, err := ts.db.ReadRegion(ctx, ts.regionName)
	require.NoError(err)
	require.Equal(revCreate, revRead1)
	require.NotNil(r1)

	require.Equal(region1.Details, r1.Details)

	revUpdate, err := ts.db.UpdateRegion(ctx, ts.regionName, region2)
	require.NoError(err)
	require.Less(revCreate, revUpdate)

	r2, revRead2, err := ts.db.ReadRegion(ctx, ts.regionName)
	require.NoError(err)
	require.Equal(revUpdate, revRead2)
	require.NotNil(r1)

	require.Equal(region2.Details, r2.Details)
}

func (ts *DBInventoryTestSuite) TestUpdateZone() {
	require := ts.Require()

	ctx, span := tracing.StartSpan(
		context.Background(),
		tracing.WithContextValue(timestamp.OutsideTime))
	defer span.End()

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
	require.Less(int64(0), revCreate)

	z1, revRead1, err := ts.db.ReadZone(ctx, ts.regionName, ts.zoneName)
	require.NoError(err)
	require.Equal(revCreate, revRead1)
	require.NotNil(z1)

	require.Equal(zone1.Details, z1.Details)

	revUpdate, err := ts.db.UpdateZone(ctx, ts.regionName, ts.zoneName, zone2)
	require.NoError(err)
	require.Less(revCreate, revUpdate)

	z2, revRead2, err := ts.db.ReadZone(ctx, ts.regionName, ts.zoneName)
	require.NoError(err)
	require.Equal(revUpdate, revRead2)
	require.NotNil(z1)

	require.Equal(zone2.Details, z2.Details)
}

func (ts *DBInventoryTestSuite) TestUpdateRack() {
	require := ts.Require()

	ctx, span := tracing.StartSpan(
		context.Background(),
		tracing.WithContextValue(timestamp.OutsideTime))
	defer span.End()

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
	require.Less(int64(0), revCreate)

	r1, revRead1, err := ts.db.ReadRack(ctx, ts.regionName, ts.zoneName, rackName)
	require.NoError(err)
	require.Equal(revCreate, revRead1)
	require.NotNil(r1)

	require.Equal(rack1.Details, r1.Details)

	revUpdate, err := ts.db.UpdateRack(ctx, ts.regionName, ts.zoneName, rackName, rack2)
	require.NoError(err)
	require.Less(revCreate, revUpdate)

	r2, revRead2, err := ts.db.ReadRack(ctx, ts.regionName, ts.zoneName, rackName)
	require.NoError(err)
	require.Equal(revUpdate, revRead2)
	require.NotNil(r1)

	require.Equal(rack2.Details, r2.Details)
}

func (ts *DBInventoryTestSuite) TestUpdatePdu() {
	require := ts.Require()

	ctx, span := tracing.StartSpan(
		context.Background(),
		tracing.WithContextValue(timestamp.OutsideTime))
	defer span.End()

	rackName := "rackUpdatePdu"

	pdu1 := &pb.Definition_Pdu{
		Details: &pb.PduDetails{Enabled: true, Condition: pb.Condition_operational},
		Ports: map[int64]*pb.PowerPort{
			int64(0): {Wired: true, Item: &pb.Hardware{Type: pb.Hardware_tor, Id: int64(0), Port: int64(0)}},
		},
	}

	pdu2 := &pb.Definition_Pdu{
		Details: &pb.PduDetails{Enabled: false, Condition: pb.Condition_out_for_repair},
		Ports: map[int64]*pb.PowerPort{
			int64(0): {Wired: true, Item: &pb.Hardware{Type: pb.Hardware_tor, Id: int64(0), Port: int64(0)}},
		},
	}

	revCreate, err := ts.db.CreatePdu(ctx, ts.regionName, ts.zoneName, rackName, ts.pduID, pdu1)
	require.NoError(err)
	require.Less(int64(0), revCreate)

	p1, revRead1, err := ts.db.ReadPdu(ctx, ts.regionName, ts.zoneName, rackName, ts.pduID)
	require.NoError(err)
	require.Equal(revCreate, revRead1)
	require.NotNil(p1)

	require.Equal(pdu1, p1)

	revUpdate, err := ts.db.UpdatePdu(ctx, ts.regionName, ts.zoneName, rackName, ts.pduID, pdu2)
	require.NoError(err)
	require.Less(revCreate, revUpdate)

	p2, revRead2, err := ts.db.ReadPdu(ctx, ts.regionName, ts.zoneName, rackName, ts.pduID)
	require.NoError(err)
	require.Equal(revUpdate, revRead2)
	require.NotNil(p1)

	require.Equal(pdu2, p2)
}

func (ts *DBInventoryTestSuite) TestUpdateTor() {
	require := ts.Require()

	ctx, span := tracing.StartSpan(
		context.Background(),
		tracing.WithContextValue(timestamp.OutsideTime))
	defer span.End()

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
	require.Less(int64(0), revCreate)

	t1, revRead1, err := ts.db.ReadTor(ctx, ts.regionName, ts.zoneName, rackName, ts.torID)
	require.NoError(err)
	require.Equal(revCreate, revRead1)
	require.NotNil(t1)

	require.Equal(tor1, t1)

	revUpdate, err := ts.db.UpdateTor(ctx, ts.regionName, ts.zoneName, rackName, ts.torID, tor2)
	require.NoError(err)
	require.Less(revCreate, revUpdate)

	t2, revRead2, err := ts.db.ReadTor(ctx, ts.regionName, ts.zoneName, rackName, ts.torID)
	require.NoError(err)
	require.Equal(revUpdate, revRead2)
	require.NotNil(t1)

	require.Equal(tor2, t2)
}

func (ts *DBInventoryTestSuite) TestUpdateBlade() {
	require := ts.Require()

	ctx, span := tracing.StartSpan(
		context.Background(),
		tracing.WithContextValue(timestamp.OutsideTime))
	defer span.End()

	rackName := "rackUpdateBlade"

	blade1 := &pb.Definition_Blade{
		Details:       &pb.BladeDetails{Enabled: true, Condition: pb.Condition_operational},
		Capacity:      &pb.BladeCapacity{Cores: 16, MemoryInMb: 16384, DiskInGb: 240, NetworkBandwidthInMbps: 2048, Arch: "X64"},
		BootInfo:      &pb.BladeBootInfo{Source: pb.BladeBootInfo_network, Image: "standard.vhdx", Version: "latest", Parameters: "-version=1 -node=R1Z1R1B1"},
		BootOnPowerOn: true,
	}

	blade2 := &pb.Definition_Blade{
		Details:       &pb.BladeDetails{Enabled: true, Condition: pb.Condition_retiring},
		Capacity:      &pb.BladeCapacity{Cores: 24, MemoryInMb: 32768, DiskInGb: 240, NetworkBandwidthInMbps: 2048, Arch: "X64"},
		BootInfo:      &pb.BladeBootInfo{Source: pb.BladeBootInfo_network, Image: "standard.vhdx", Version: "latest", Parameters: "-version=1 -node=R1Z1R1B1"},
		BootOnPowerOn: false,
	}

	revCreate, err := ts.db.CreateBlade(ctx, ts.regionName, ts.zoneName, rackName, ts.bladeID, blade1)
	require.NoError(err)
	require.Less(int64(0), revCreate)

	b1, revRead1, err := ts.db.ReadBlade(ctx, ts.regionName, ts.zoneName, rackName, ts.bladeID)
	require.NoError(err)
	require.Equal(revCreate, revRead1)
	require.NotNil(b1)

	require.Equal(blade1, b1)

	revUpdate, err := ts.db.UpdateBlade(ctx, ts.regionName, ts.zoneName, rackName, ts.bladeID, blade2)
	require.NoError(err)
	require.Less(revCreate, revUpdate)

	b2, revRead2, err := ts.db.ReadBlade(ctx, ts.regionName, ts.zoneName, rackName, ts.bladeID)
	require.NoError(err)
	require.Equal(revUpdate, revRead2)
	require.NotNil(b1)

	require.Equal(blade2, b2)
}

func (ts *DBInventoryTestSuite) TestDeleteRegion() {
	require := ts.Require()

	ctx, span := tracing.StartSpan(
		context.Background(),
		tracing.WithContextValue(timestamp.OutsideTime))
	defer span.End()

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
	require.Less(int64(0), revCreate)

	r1, revRead1, err := ts.db.ReadRegion(ctx, regionName)
	require.NoError(err)
	require.Equal(revCreate, revRead1)
	require.NotNil(r1)

	require.Equal(region1.Details, r1.Details)

	revDelete, err := ts.db.DeleteRegion(ctx, regionName)
	require.NoError(err)
	require.Less(revCreate, revDelete)

	r2, revRead2, err := ts.db.ReadRegion(ctx, regionName)
	require.Equal(err, errors.ErrRegionNotFound{Region: regionName})
	require.Equal(int64(InvalidRev), revRead2)
	require.Nil(r2)
}

func (ts *DBInventoryTestSuite) TestDeleteZone() {
	require := ts.Require()

	ctx, span := tracing.StartSpan(
		context.Background(),
		tracing.WithContextValue(timestamp.OutsideTime))
	defer span.End()

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
	require.Less(int64(0), revCreate)

	z1, revRead1, err := ts.db.ReadZone(ctx, ts.regionName, zoneName)
	require.NoError(err)
	require.Equal(revCreate, revRead1)
	require.NotNil(z1)

	require.Equal(zone1.Details, z1.Details)

	revDelete, err := ts.db.DeleteZone(ctx, ts.regionName, zoneName)
	require.NoError(err)
	require.Less(revCreate, revDelete)

	z2, revRead2, err := ts.db.ReadZone(ctx, ts.regionName, zoneName)
	require.Equal(err, errors.ErrZoneNotFound{Region: ts.regionName, Zone: zoneName})
	require.Equal(int64(InvalidRev), revRead2)
	require.Nil(z2)
}

func (ts *DBInventoryTestSuite) TestDeleteRack() {
	require := ts.Require()

	ctx, span := tracing.StartSpan(
		context.Background(),
		tracing.WithContextValue(timestamp.OutsideTime))
	defer span.End()

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
	require.Less(int64(0), revCreate)

	r1, revRead1, err := ts.db.ReadRack(ctx, ts.regionName, ts.zoneName, rackName)
	require.NoError(err)
	require.Equal(revCreate, revRead1)
	require.NotNil(r1)

	require.Equal(rack.Details, r1.Details)

	revDelete, err := ts.db.DeleteRack(ctx, ts.regionName, ts.zoneName, rackName)
	require.NoError(err)
	require.Less(revCreate, revDelete)

	r2, revRead2, err := ts.db.ReadRack(ctx, ts.regionName, ts.zoneName, rackName)
	require.Equal(err, errors.ErrRackNotFound{Region: ts.regionName, Zone: ts.zoneName, Rack: rackName})
	require.Equal(int64(InvalidRev), revRead2)
	require.Nil(r2)
}

func (ts *DBInventoryTestSuite) TestDeletePdu() {
	require := ts.Require()

	ctx, span := tracing.StartSpan(
		context.Background(),
		tracing.WithContextValue(timestamp.OutsideTime))
	defer span.End()

	rackName := "rackDeletePdu"

	pdu := &pb.Definition_Pdu{
		Details: &pb.PduDetails{Enabled: true, Condition: pb.Condition_operational},
		Ports: map[int64]*pb.PowerPort{
			int64(0): {Wired: true, Item: &pb.Hardware{Type: pb.Hardware_tor, Id: int64(0), Port: int64(0)}},
		},
	}

	revCreate, err := ts.db.CreatePdu(ctx, ts.regionName, ts.zoneName, rackName, ts.pduID, pdu)
	require.NoError(err)
	require.Less(int64(0), revCreate)

	p1, revRead1, err := ts.db.ReadPdu(ctx, ts.regionName, ts.zoneName, rackName, ts.pduID)
	require.NoError(err)
	require.Equal(revCreate, revRead1)
	require.NotNil(p1)

	require.Equal(pdu.Details, p1.Details)

	revDelete, err := ts.db.DeletePdu(ctx, ts.regionName, ts.zoneName, rackName, ts.pduID)
	require.NoError(err)
	require.Less(revCreate, revDelete)

	p2, revRead2, err := ts.db.ReadPdu(ctx, ts.regionName, ts.zoneName, rackName, ts.pduID)
	require.Equal(err, errors.ErrPduNotFound{Region: ts.regionName, Zone: ts.zoneName, Rack: rackName, Pdu: ts.pduID})
	require.Equal(int64(InvalidRev), revRead2)
	require.Nil(p2)
}

func (ts *DBInventoryTestSuite) TestDeleteTor() {
	require := ts.Require()

	ctx, span := tracing.StartSpan(
		context.Background(),
		tracing.WithContextValue(timestamp.OutsideTime))
	defer span.End()

	rackName := "rackDeleteTor"

	tor := &pb.Definition_Tor{
		Details: &pb.TorDetails{Enabled: true, Condition: pb.Condition_operational},
		Ports: map[int64]*pb.NetworkPort{
			int64(0): {Wired: true, Item: &pb.Hardware{Type: pb.Hardware_pdu, Id: int64(0), Port: int64(0)}},
		},
	}

	revCreate, err := ts.db.CreateTor(ctx, ts.regionName, ts.zoneName, rackName, ts.torID, tor)
	require.NoError(err)
	require.Less(int64(0), revCreate)

	t1, revRead1, err := ts.db.ReadTor(ctx, ts.regionName, ts.zoneName, rackName, ts.torID)
	require.NoError(err)
	require.Equal(revCreate, revRead1)
	require.NotNil(t1)

	require.Equal(tor.Details, t1.Details)

	revDelete, err := ts.db.DeleteTor(ctx, ts.regionName, ts.zoneName, rackName, ts.torID)
	require.NoError(err)
	require.Less(revCreate, revDelete)

	t2, revRead2, err := ts.db.ReadTor(ctx, ts.regionName, ts.zoneName, rackName, ts.torID)
	require.Equal(err, errors.ErrTorNotFound{Region: ts.regionName, Zone: ts.zoneName, Rack: rackName, Tor: ts.torID})
	require.Equal(int64(InvalidRev), revRead2)
	require.Nil(t2)
}

func (ts *DBInventoryTestSuite) TestDeleteBlade() {
	require := ts.Require()

	ctx, span := tracing.StartSpan(
		context.Background(),
		tracing.WithContextValue(timestamp.OutsideTime))
	defer span.End()

	rackName := "rackUpdateBlade"

	blade1 := &pb.Definition_Blade{
		Details:       &pb.BladeDetails{Enabled: true, Condition: pb.Condition_operational},
		Capacity:      &pb.BladeCapacity{Cores: 16, MemoryInMb: 16384, DiskInGb: 240, NetworkBandwidthInMbps: 2048, Arch: "X64"},
		BootInfo:      &pb.BladeBootInfo{Source: pb.BladeBootInfo_network, Image: "standard.vhdx", Version: "latest", Parameters: "-version=1 -node=R1Z1R1B1"},
		BootOnPowerOn: true,
	}

	revCreate, err := ts.db.CreateBlade(ctx, ts.regionName, ts.zoneName, rackName, ts.bladeID, blade1)
	require.NoError(err)
	require.Less(int64(0), revCreate)

	b1, revRead1, err := ts.db.ReadBlade(ctx, ts.regionName, ts.zoneName, rackName, ts.bladeID)
	require.NoError(err)
	require.Equal(revCreate, revRead1)
	require.NotNil(b1)

	require.Equal(blade1, b1)

	revDelete, err := ts.db.DeleteBlade(ctx, ts.regionName, ts.zoneName, rackName, ts.bladeID)
	require.NoError(err)
	require.Less(revCreate, revDelete)

	b2, revRead2, err := ts.db.ReadBlade(ctx, ts.regionName, ts.zoneName, rackName, ts.bladeID)
	require.Equal(err, errors.ErrBladeNotFound{Region: ts.regionName, Zone: ts.zoneName, Rack: rackName, Blade: ts.bladeID})
	require.Equal(int64(InvalidRev), revRead2)
	require.Nil(b2)
}
