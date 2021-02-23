package frontend

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Jim3Things/CloudChamber/simulation/internal/clients/inventory"
	pb "github.com/Jim3Things/CloudChamber/simulation/pkg/protos/inventory"
)

type DBInventoryTestSuite struct {
	testSuiteCore

	db *DBInventory

	zoneName string
	rackName string
	pduID    int64
	torID    int64
	bladeID  int64
}

func (ts *DBInventoryTestSuite) SetupSuite() {

	ts.testSuiteCore.SetupSuite()

	ts.zoneName = "zoneBasic"
	ts.rackName = "rackBasic"
	ts.pduID = int64(17)
	ts.torID = int64(31)
	ts.bladeID = int64(100)
}

func (ts *DBInventoryTestSuite) SetupTest() {
	require := ts.Require()

	_ = ts.utf.Open(ts.T())

	ts.db = NewDbInventory()

	err := ts.db.Initialize(context.Background(), ts.cfg)
	require.NoError(err)
}

func (ts *DBInventoryTestSuite) TearDownTest() {
	ts.db.Store.Disconnect()
	ts.utf.Close()
}

func (ts *DBInventoryTestSuite) TestInitializeInventory() {
	assert := ts.Assert()
	require := ts.Require()

	dbInventory := NewDbInventory()
	require.NotNil(dbInventory)
	assert.NotNil(dbInventory.Store)

	err := dbInventory.Initialize(context.Background(), ts.cfg)
	require.NoError(err)
	assert.NotNil(dbInventory.cfg)
}

func (ts *DBInventoryTestSuite) TestCreateZone() {
	assert := ts.Assert()
	require := ts.Require()

	ctx := context.Background()

	zoneName := "zone1"

	zone := &pb.Definition_Zone{
		Details: &pb.ZoneDetails{
			Enabled:  true,
			State:    pb.State_in_service,
			Location: "Nowhere in particular",
			Notes:    "empty notes",
		},
	}

	revCreate, err := ts.db.CreateZone(ctx, zoneName, zone)
	require.NoError(err)
	assert.Less(int64(0), revCreate)

	z, revRead, err := ts.db.ReadZone(ctx, zoneName)
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

	rackName := "rack1"

	rack := &pb.Definition_Rack{
		Details: &pb.RackDetails{
			Enabled:   true,
			Condition: pb.Condition_operational,
			Location:  "In " + ts.zoneName,
			Notes:     "Basic rack for test",
		},
	}

	revCreate, err := ts.db.CreateRack(ctx, ts.zoneName, rackName, rack)
	require.NoError(err)

	r, revRead, err := ts.db.ReadRack(ctx, ts.zoneName, rackName)
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

	revCreate, err := ts.db.CreatePdu(ctx, ts.zoneName, ts.rackName, pduID, pdu)
	require.NoError(err)

	t, revRead, err := ts.db.ReadPdu(ctx, ts.zoneName, ts.rackName, pduID)
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

	revCreate, err := ts.db.CreateTor(ctx, ts.zoneName, ts.rackName, torID, tor)
	require.NoError(err)

	t, revRead, err := ts.db.ReadTor(ctx, ts.zoneName, ts.rackName, torID)
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

	revCreate, err := ts.db.CreateBlade(ctx, ts.zoneName, ts.rackName, bladeID, blade)
	require.NoError(err)

	b, revRead, err := ts.db.ReadBlade(ctx, ts.zoneName, ts.rackName, bladeID)
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

func (ts *DBInventoryTestSuite) TestReadInventoryFromStore() {
	require := ts.Require()

	_, err := ts.db.readInventoryDefinitionFromStore(context.Background())
	require.NoError(err)
}

func (ts *DBInventoryTestSuite) TestReadInventoryDefinitionFromFileExBasic() {
	require := ts.Require()

	_, err := inventory.ReadInventoryDefinitionFromFileEx(context.Background(), "./testdata/basic")
	require.NoError(err)
}

func (ts *DBInventoryTestSuite) TestReadInventoryDefinitionFromFileExExtended() {
	require := ts.Require()

	_, err := inventory.ReadInventoryDefinitionFromFileEx(context.Background(), "./testdata/extended")
	require.NoError(err)
}

func (ts *DBInventoryTestSuite) TestReadInventoryDefinitionFromFileExStandard() {
	require := ts.Require()

	_, err := inventory.ReadInventoryDefinitionFromFileEx(context.Background(), "./testdata/standard")
	require.NoError(err)
}

func (ts *DBInventoryTestSuite) TestLoadInventoryIntoStore() {
	require := ts.Require()

	ctx := context.Background()

	root, err := ts.db.readInventoryDefinitionFromStore(ctx)
	require.NoError(err)

	err = ts.db.deleteInventoryDefinitionFromStore(ctx, root)
	require.NoError(err)

	root, err = inventory.ReadInventoryDefinitionFromFileEx(ctx, "./testdata/extended")
	require.NoError(err)
	require.NotNil(root)

	err = ts.db.writeInventoryDefinitionToStore(ctx, root)
	require.NoError(err)

	rootReload, err := ts.db.readInventoryDefinitionFromStore(ctx)
	require.NoError(err)
	require.NotNil(rootReload)

	err = ts.db.deleteInventoryDefinitionFromStore(ctx, rootReload)
	require.NoError(err)
}

func (ts *DBInventoryTestSuite) TestUpdateInventoryDefinitionBasic() {
	require := ts.Require()

	ctx := context.Background()

	err := ts.db.UpdateInventoryDefinition(ctx, "./testdata/basic")
	require.NoError(err)

	err = ts.db.UpdateInventoryDefinition(ctx, "./testdata/basic")
	require.NoError(err)
}

func (ts *DBInventoryTestSuite) TestUpdateInventoryDefinitionStandard() {
	require := ts.Require()

	ctx := context.Background()

	err := ts.db.UpdateInventoryDefinition(ctx, "./testdata/standard")
	require.NoError(err)

	err = ts.db.UpdateInventoryDefinition(ctx, "./testdata/standard")
	require.NoError(err)
}

func (ts *DBInventoryTestSuite) TestUpdateInventoryDefinitionExtended() {
	require := ts.Require()

	ctx := context.Background()

	err := ts.db.UpdateInventoryDefinition(ctx, "./testdata/extended")
	require.NoError(err)

	err = ts.db.UpdateInventoryDefinition(ctx, "./testdata/extended")
	require.NoError(err)
}

func (ts *DBInventoryTestSuite) TestDeleteInventoryDefinitionBasic() {
	require := ts.Require()

	ctx := context.Background()

	err := ts.db.UpdateInventoryDefinition(ctx, "./testdata/basic")
	require.NoError(err)

	err = ts.db.DeleteInventoryDefinition(ctx)
	require.NoError(err)
}

func (ts *DBInventoryTestSuite) TestDeleteInventoryDefinitionExtended() {
	require := ts.Require()

	ctx := context.Background()

	err := ts.db.UpdateInventoryDefinition(ctx, "./testdata/extended")
	require.NoError(err)

	err = ts.db.DeleteInventoryDefinition(ctx)
	require.NoError(err)
}

func TestDBInventoryTestSuite(t *testing.T) {
	suite.Run(t, new(DBInventoryTestSuite))
}
