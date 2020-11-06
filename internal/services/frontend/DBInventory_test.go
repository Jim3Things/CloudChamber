package frontend

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Jim3Things/CloudChamber/internal/clients/store"
	ct "github.com/Jim3Things/CloudChamber/pkg/protos/common"
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
		MaxCapacity:   &ct.BladeCapacity{},
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
	// 	db :=  &DBInventory{
	// 		mutex: sync.RWMutex{},
	// 		Zone: nil,
	// 		MaxBladeCount: 0,
	// 		MaxCapacity:   &ct.BladeCapacity{},
	// 		Store: store.NewStore(),
	// 	}

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
			Enabled:   true,
			Condition: pb.Definition_operational,
			Location:  "Pacific NW",
			Notes:     "Basic Zone for test",
		},
	)

	require.NoError(err)

	_, err = ts.db.CreateRack(
		ctx,
		ts.zoneName,
		ts.rackName,
		&pb.DefinitionRack{
			Enabled:   true,
			Condition: pb.Definition_operational,
			Location:  "In " + ts.zoneName,
			Notes:     "Basic rack for test",
			},
		)

	require.NoError(err)

	pdu := pb.DefinitionPdu{
		Enabled: true,
		Powered: true,
		Condition: pb.Definition_operational,
		Ports: make(map[int64]*pb.DefinitionPowerPort),
	}

	tor := pb.DefinitionTor{
		Enabled: true,
		Powered: true,
		Condition: pb.Definition_operational,
		Ports: make(map[int64]*pb.DefinitionNetworkPort),
	}

	blade := pb.DefinitionBlade{
		Enabled: true,
		Condition: pb.Definition_operational,
		Capacity: &ct.BladeCapacity{
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


// func (ts *DBInventoryTestSuite) ensureBasicRack() {
// 	require := ts.Require()

// 	ta.ensureBasicZone()
// }


func (ts *DBInventoryTestSuite) TestInitializeInventory() {
	assert := ts.Assert()

	db :=  &DBInventory{
		mutex: sync.RWMutex{},
		Zone: nil,
		MaxBladeCount: 0,
		MaxCapacity:   &ct.BladeCapacity{},
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

	// db :=  &DBInventory{
	// 	mutex: sync.RWMutex{},
	// 	Zone: nil,
	// 	MaxBladeCount: 0,
	// 	MaxCapacity:   &ct.BladeCapacity{},
	// 	Store: store.NewStore(),
	// }

	// err := db.Store.Connect()

	// require.NoError(err)

	zoneName := "zone1"

	zone := &pb.DefinitionZone{
		Enabled: true,
		Condition: pb.Definition_operational,
		Location: "Nowhere in particular",
		Notes: "empty notes",
	}

	revCreate, err := ts.db.CreateZone(ctx, zoneName, zone)
	require.NoError(err)
	assert.Less(int64(0), revCreate)

	z, revRead, err := ts.db.ReadZone(ctx, zoneName)
	assert.NoError(err)
	assert.Equal(revCreate, revRead)
	require.NotNil(z)

	assert.Equal(zone.Enabled,   z.Enabled)
	assert.Equal(zone.Condition, z.Condition)
	assert.Equal(zone.Location,  z.Location)
	assert.Equal(zone.Notes,     z.Notes)
	}
	
func (ts *DBInventoryTestSuite) TestCreateRack() {
	assert := ts.Assert()
	require := ts.Require()

	ctx := context.Background()

	rackName := "rack1"

	rack := &pb.DefinitionRack{
		Enabled:   true,
		Condition: pb.Definition_operational,
		Location:  "In " + ts.zoneName,
		Notes:     "Basic rack for test",
		}

	revCreate, err := ts.db.CreateRack(ctx, ts.zoneName, rackName, rack)
	require.NoError(err)

	r, revRead, err := ts.db.ReadRack(ctx, ts.zoneName, rackName)
	assert.NoError(err)
	assert.Equal(revCreate, revRead)
	require.NotNil(r)

	assert.Equal(rack.Enabled,   r.Enabled)
	assert.Equal(rack.Condition, r.Condition)
	assert.Equal(rack.Location,  r.Location)
	assert.Equal(rack.Notes,     r.Notes)
}

func (ts *DBInventoryTestSuite) TestCreatePdu() {
	assert := ts.Assert()
	require := ts.Require()

	ctx := context.Background()

	pduID := int64(1)

	pdu := &pb.DefinitionPdu{
		Enabled:   true,
		Powered:   true,
		Condition: pb.Definition_operational,
		Ports: make(map[int64]*pb.DefinitionPowerPort),
		}

	pdu.Ports[0] = &pb.DefinitionPowerPort{
		Connected: true,
		Powered: true,
	}

	pdu.Ports[1] = &pb.DefinitionPowerPort{
		Connected: false,
		Powered: false,
	}

	pdu.Ports[2] = &pb.DefinitionPowerPort{
		Connected: false,
		Powered: true,
	}

	pdu.Ports[3] = &pb.DefinitionPowerPort{
		Connected: true,
		Powered: false,
	}


	revCreate, err := ts.db.CreatePdu(ctx, ts.zoneName, ts.rackName, pduID, pdu)
	require.NoError(err)

	t, revRead, err := ts.db.ReadPdu(ctx, ts.zoneName, ts.rackName, pduID)
	assert.NoError(err)
	assert.Equal(revCreate, revRead)
	require.NotNil(t)

	assert.Equal(pdu.Enabled,   t.Enabled)
	assert.Equal(pdu.Powered,   t.Powered)
	assert.Equal(pdu.Condition, t.Condition)
	assert.Equal(len(pdu.Ports), len(t.Ports))

	for i, p := range pdu.Ports {
		tp, ok := t.Ports[i]
		assert.True(ok)
		assert.Equal(p.Connected, tp.Connected)
		assert.Equal(p.Powered, tp.Powered)
	}
}

func (ts *DBInventoryTestSuite) TestCreateTor() {
	assert := ts.Assert()
	require := ts.Require()

	ctx := context.Background()

	torID := int64(1)

	tor := &pb.DefinitionTor{
		Enabled:   true,
		Powered:   true,
		Condition: pb.Definition_operational,
		Ports: make(map[int64]*pb.DefinitionNetworkPort),
		}

	tor.Ports[0] = &pb.DefinitionNetworkPort{
		Connected: true,
		Enabled: true,
	}

	tor.Ports[1] = &pb.DefinitionNetworkPort{
		Connected: false,
		Enabled: false,
	}

	tor.Ports[2] = &pb.DefinitionNetworkPort{
		Connected: false,
		Enabled: true,
	}

	tor.Ports[3] = &pb.DefinitionNetworkPort{
		Connected: true,
		Enabled: false,
	}


	revCreate, err := ts.db.CreateTor(ctx, ts.zoneName, ts.rackName, torID, tor)
	require.NoError(err)

	t, revRead, err := ts.db.ReadTor(ctx, ts.zoneName, ts.rackName, torID)
	assert.NoError(err)
	assert.Equal(revCreate, revRead)
	require.NotNil(t)

	assert.Equal(tor.Enabled,   t.Enabled)
	assert.Equal(tor.Powered,   t.Powered)
	assert.Equal(tor.Condition, t.Condition)
	assert.Equal(len(tor.Ports), len(t.Ports))

	for i, p := range tor.Ports {
		tp, ok := t.Ports[i]
		assert.True(ok)
		assert.Equal(p.Enabled, tp.Enabled)
		assert.Equal(p.Connected, tp.Connected)
	}
}

func (ts *DBInventoryTestSuite) TestCreateBlade() {
	assert := ts.Assert()
	require := ts.Require()

	ctx := context.Background()

	bladeID := int64(1)

	blade := &pb.DefinitionBlade{
		Enabled: true,
		Condition: pb.Definition_operational,
		Capacity: &ct.BladeCapacity{
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

func (ts *DBInventoryTestSuite) TestLoadFromStore() {
	assert := ts.Assert()
	require := ts.Require()

	ctx := context.Background()

	db :=  &DBInventory{
		mutex: sync.RWMutex{},
		Zone: nil,
		MaxBladeCount: 0,
		MaxCapacity:   &ct.BladeCapacity{},
		Store: store.NewStore(),
	}

	err := db.Store.Connect()
	require.NoError(err)

	err = db.LoadFromStore(ctx)
	require.NoError(err)
	require.Equal(1, len(db.Zones))

	for _, z := range db.Zones {

		assert.True(z.Details.Enabled)
		assert.Equal(pb.Definition_operational, z.Details.Condition)
		assert.Equal(0,  len(z.Details.Location))
		assert.Equal(0,  len(z.Details.Notes))
		
		assert.Equal(8, len(z.Racks))

		for _, r := range z.Racks {
			assert.True(r.Details.Enabled)
			assert.Equal(pb.Definition_operational, r.Details.Condition)
			assert.Equal(0,  len(r.Details.Location))
			assert.Equal(0,  len(r.Details.Notes))
			
			assert.Equal(1, len(r.Pdus))

			for _, p := range r.Pdus {
				assert.True(p.Enabled)
				assert.True(p.Powered)

				assert.Equal(pb.Definition_operational, p.Condition)

				assert.Equal(9, len(p.Ports))
			}

			assert.Equal(1, len(r.Tors))

			for _, tor := range r.Tors {
				assert.True(tor.Enabled)
				assert.True(tor.Powered)

				assert.Equal(pb.Definition_operational, tor.Condition)

				assert.Equal(8, len(tor.Ports))
			}

			assert.Equal(8, len(z.Racks))

			for _, b := range r.Blades {
				assert.True(b.Enabled)
				assert.Equal(pb.Definition_operational, b.Condition)

				assert.Equal(8,     b.Capacity.Cores)
				assert.Equal(16834, b.Capacity.MemoryInMb)
				assert.Equal(120,   b.Capacity.DiskInGb)
				assert.Equal(2048,  b.Capacity.NetworkBandwidthInMbps)
				assert.Equal("X64", b.Capacity.Arch)
			}
		}
	}
}

// func (ts *DBInventoryTestSuite) TestUpdateFromFile() {
// 	assert := ts.Assert()
// 	require := ts.Require()

// 	ctx := context.Background()

// 	db :=  &DBInventory{
// 		mutex: sync.RWMutex{},
// 		Zone: nil,
// 		MaxBladeCount: 0,
// 		MaxCapacity:   &ct.BladeCapacity{},
// 		Store: store.NewStore(),
// 	}

// 	err := db.Store.Connect()

// 	require.NoError(err)

// 	err = db.LoadFromStore(ctx)

// 	require.NoError(err)
// 	require.Equal(1, len(db.zones))


// 	err = db.UpdateFromFile(ctx, ts.cfg)

// 	require.NoError(err)
// 	assert.Equal(1, len(db.zones))
// }


func TestDBInventoryTestSuite(t *testing.T) {
	suite.Run(t, new(DBInventoryTestSuite))
}
