package frontend

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Jim3Things/CloudChamber/internal/clients/store"
	pb "github.com/Jim3Things/CloudChamber/pkg/protos/inventory"
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
//	require := ts.Require()

	ts.testSuiteCore.SetupSuite()

//	configPath := flag.String("config", "./testdata", "path to the configuration file")
//	flag.Parse()

//	cfg, err := config.ReadGlobalConfig(*configPath)
//	require.NoError(err)
//	require.NotNil(cfg)

//	ts.cfg = cfg

	ts.zoneName = "zoneBasic"
	ts.rackName = "rackBasic"
	ts.pduID    = int64(17)
	ts.torID    = int64(31)
	ts.bladeID  = int64(100)
}

func (ts *DBInventoryTestSuite) SetupTest() {
	require := ts.Require()

	_ = ts.utf.Open(ts.T())

	ts.db =  &DBInventory{
		mutex: sync.RWMutex{},
		Zone: nil,
		MaxBladeCount: 0,
		MaxCapacity:   &pb.BladeCapacity{},
		Store: store.NewStore(),
	}

	err := ts.db.Store.Connect()
	require.NoError(err)
}

func (ts *DBInventoryTestSuite) TearDownTest() {
	ts.utf.Close()
}



func (ts *DBInventoryTestSuite) ensureInventoryLoaded() {
	// require := ts.Require()

	// ctx := context.Background()

	// if ts.db == nil {

		// err := db.Initialize(ctx, ts.cfg)
		// require.NoError(err)
		// require.NotNil(dbInventory)

		// if ts.db == nil {
		// 	ts.db = db
		// }
	// }
}

func (ts *DBInventoryTestSuite) ensureBasicZone() {
	require := ts.Require()

	ctx := context.Background()

	_, err := ts.db.CreateZone(
		ctx,
		ts.zoneName,
		&pb.DefinitionZone{
			Details: &pb.ZoneDetails{
				Enabled:   true,
				State:     pb.State_in_service,
				Location:  "Pacific NW",
				Notes:     "Basic Zone for test",
			},
		},
	)

	require.NoError(err)

	_, err = ts.db.CreateRack(
		ctx,
		ts.zoneName,
		ts.rackName,
		&pb.DefinitionRack{
			Details: &pb.RackDetails{
				Enabled:   true,
				Condition:     pb.Condition_operational,
				Location:  "In " + ts.zoneName,
				Notes:     "Basic rack for test",
				},
			},
		)

	require.NoError(err)

	pdu := pb.DefinitionPdu{
		Details: &pb.PduDetails{
			Enabled: true,
			Condition: pb.Condition_operational,
		},
		Ports: make(map[int64]*pb.PowerPort),
	}

	tor := pb.DefinitionTor{
		Details: &pb.TorDetails{
			Enabled: true,
			Condition: pb.Condition_operational,
		},
		Ports: make(map[int64]*pb.NetworkPort),
	}

	blade := pb.DefinitionBlade{
		Enabled: true,
		Condition: pb.Condition_operational,
		Capacity: &pb.BladeCapacity{
			Cores: 8,
			MemoryInMb: 8192,
			DiskInGb: 8192,
		},
	}

	_, err = ts.db.CreatePdu(ctx, ts.zoneName, ts.rackName, ts.pduID, &pdu)
	require.NoError(err)

	_, err = ts.db.CreateTor(ctx, ts.zoneName, ts.rackName, ts.torID, &tor,)
	require.NoError(err)

	_, err = ts.db.CreateBlade(ctx, ts.zoneName, ts.rackName, ts.bladeID, &blade)
	require.NoError(err)
}

func (ts *DBInventoryTestSuite) TestInitializeInventory() {
	assert := ts.Assert()

	db :=  &DBInventory{
		mutex: sync.RWMutex{},
		Zone: nil,
		MaxBladeCount: 0,
		MaxCapacity:   &pb.BladeCapacity{},
		Store: store.NewStore(),
	}

	assert.NotNil(db.Store)

	// err := db.Initialize(context.Background(), ts.cfg)
	// assert.NoError(err)
	// assert.NotNil(dbInventory)
}

func (ts *DBInventoryTestSuite) TestCreateZone() {
	assert := ts.Assert()
	require := ts.Require()

	ctx := context.Background()

	zoneName := "zone1"

	zone := &pb.DefinitionZone{
		Details: &pb.ZoneDetails{
			Enabled: true,
			State: pb.State_in_service,
			Location: "Nowhere in particular",
			Notes: "empty notes",
		},
	}

	revCreate, err := ts.db.CreateZone(ctx, zoneName, zone)
	require.NoError(err)
	assert.Less(int64(0), revCreate)

	z, revRead, err := ts.db.ReadZone(ctx, zoneName)
	assert.NoError(err)
	assert.Equal(revCreate, revRead)
	require.NotNil(z)

	assert.Equal(zone.Details.Enabled,  z.Details.Enabled)
	assert.Equal(zone.Details.State,    z.Details.State)
	assert.Equal(zone.Details.Location, z.Details.Location)
	assert.Equal(zone.Details.Notes,    z.Details.Notes)
	}
	
func (ts *DBInventoryTestSuite) TestCreateRack() {
	assert := ts.Assert()
	require := ts.Require()

	ctx := context.Background()

	rackName := "rack1"

	rack := &pb.DefinitionRack{
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

	assert.Equal(rack.Details.Enabled,   r.Details.Enabled)
	assert.Equal(rack.Details.Condition, r.Details.Condition)
	assert.Equal(rack.Details.Location,  r.Details.Location)
	assert.Equal(rack.Details.Notes,     r.Details.Notes)
}

func (ts *DBInventoryTestSuite) TestCreatePdu() {
	assert := ts.Assert()
	require := ts.Require()

	ctx := context.Background()

	pduID := int64(1)

	pdu := &pb.DefinitionPdu{
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
		Item:  &pb.Hardware{
			Type: pb.Hardware_unknown,
		},
	}

	pdu.Ports[2] = &pb.PowerPort{
		Wired: true,
		Item:  &pb.Hardware{
			Type: pb.Hardware_tor,
			Id: 0,
			Port: 0,
		},
	}

	pdu.Ports[3] = &pb.PowerPort{
		Wired: true,
		Item:  &pb.Hardware{
			Type: pb.Hardware_tor,
			Id: 1,
			Port: 0,
		},
	}

	pdu.Ports[4] = &pb.PowerPort{
		Wired: true,
		Item:  &pb.Hardware{
			Type: pb.Hardware_blade,
			Id: 0,
			Port: 0,
		},
	}

	pdu.Ports[5] = &pb.PowerPort{
		Wired: true,
		Item:  &pb.Hardware{
			Type: pb.Hardware_blade,
			Id: 0,
			Port: 1,
		},
	}

	pdu.Ports[6] = &pb.PowerPort{
		Wired: true,
		Item:  &pb.Hardware{
			Type: pb.Hardware_blade,
			Id: 1,
			Port: 0,
		},
	}

	pdu.Ports[7] = &pb.PowerPort{
		Wired: true,
		Item:  &pb.Hardware{
			Type: pb.Hardware_blade,
			Id: 1,
			Port: 1,
		},
	}


	revCreate, err := ts.db.CreatePdu(ctx, ts.zoneName, ts.rackName, pduID, pdu)
	require.NoError(err)

	t, revRead, err := ts.db.ReadPdu(ctx, ts.zoneName, ts.rackName, pduID)
	assert.NoError(err)
	assert.Equal(revCreate, revRead)
	require.NotNil(t)

	assert.Equal(pdu.Details.Enabled,   t.Details.Enabled)
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
			assert.Equal(p.Item.Id,   tp.Item.Id)
			assert.Equal(p.Item.Port, tp.Item.Port)
		}
	}
}

func (ts *DBInventoryTestSuite) TestCreateTor() {
	assert := ts.Assert()
	require := ts.Require()

	ctx := context.Background()

	torID := int64(1)

	tor := &pb.DefinitionTor{
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
		Item:  &pb.Hardware{
			Type: pb.Hardware_unknown,			
		},
	}

	tor.Ports[2] = &pb.NetworkPort{
		Wired: true,
		Item:  &pb.Hardware{
			Type: pb.Hardware_pdu,
			Id: 0,
			Port: 0,
		},
	}

	tor.Ports[3] = &pb.NetworkPort{
		Wired: true,
		Item:  &pb.Hardware{
			Type: pb.Hardware_pdu,
			Id: 1,
			Port: 0,
		},
	}

	tor.Ports[4] = &pb.NetworkPort{
		Wired: true,
		Item:  &pb.Hardware{
			Type: pb.Hardware_blade,
			Id: 0,
			Port: 0,
		},
	}

	tor.Ports[5] = &pb.NetworkPort{
		Wired: true,
		Item:  &pb.Hardware{
			Type: pb.Hardware_blade,
			Id: 0,
			Port: 1,
		},
	}

	tor.Ports[6] = &pb.NetworkPort{
		Wired: true,
		Item:  &pb.Hardware{
			Type: pb.Hardware_blade,
			Id: 1,
			Port: 0,
		},
	}

	tor.Ports[7] = &pb.NetworkPort{
		Wired: true,
		Item:  &pb.Hardware{
			Type: pb.Hardware_blade,
			Id: 1,
			Port: 1,
		},
	}

	revCreate, err := ts.db.CreateTor(ctx, ts.zoneName, ts.rackName, torID, tor)
	require.NoError(err)

	t, revRead, err := ts.db.ReadTor(ctx, ts.zoneName, ts.rackName, torID)
	assert.NoError(err)
	assert.Equal(revCreate, revRead)
	require.NotNil(t)

	assert.Equal(tor.Details.Enabled,   t.Details.Enabled)
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
			assert.Equal(p.Item.Id,   tp.Item.Id)
			assert.Equal(p.Item.Port, tp.Item.Port)
		}
	}
}

func (ts *DBInventoryTestSuite) TestCreateBlade() {
	assert := ts.Assert()
	require := ts.Require()

	ctx := context.Background()

	bladeID := int64(1)

	blade := &pb.DefinitionBlade{
		Enabled: true,
		Condition: pb.Condition_operational,
		Capacity: &pb.BladeCapacity{
			Cores: 8,
			MemoryInMb: 8192,
			DiskInGb: 8192,
			NetworkBandwidthInMbps: 1024,
			Arch: "vax",
			Accelerators: nil,
		},
	}

	revCreate, err := ts.db.CreateBlade(ctx, ts.zoneName, ts.rackName, bladeID, blade)
	require.NoError(err)

	b, revRead, err := ts.db.ReadBlade(ctx, ts.zoneName, ts.rackName, bladeID)
	assert.NoError(err)
	assert.Equal(revCreate, revRead)
	require.NotNil(b)

	assert.Equal(blade.Enabled,   b.Enabled)
	assert.Equal(blade.Condition, b.Condition)
	
	assert.Equal(blade.Capacity.Cores,                  b.Capacity.Cores)
	assert.Equal(blade.Capacity.MemoryInMb,             b.Capacity.MemoryInMb)
	assert.Equal(blade.Capacity.DiskInGb,               b.Capacity.DiskInGb)
	assert.Equal(blade.Capacity.NetworkBandwidthInMbps, b.Capacity.NetworkBandwidthInMbps)
	assert.Equal(blade.Capacity.Arch,                   b.Capacity.Arch)
	assert.Equal(blade.Capacity.Accelerators,           b.Capacity.Accelerators)
}

func TestDBInventoryTestSuite(t *testing.T) {
	suite.Run(t, new(DBInventoryTestSuite))
}
