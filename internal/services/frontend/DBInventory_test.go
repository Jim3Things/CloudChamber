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

		assert.True(z.Enabled)
		assert.Equal(pb.Definition_operational, z.Condition)
		assert.Equal(0,  len(z.Location))
		assert.Equal(0,  len(z.Notes))
		
		assert.Equal(8, len(z.Racks))

		for _, r := range z.Racks {
			assert.True(r.Enabled)
			assert.Equal(pb.Definition_operational, r.Condition)
			assert.Equal(0,  len(r.Location))
			assert.Equal(0,  len(r.Notes))
			
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

// 	require.Nil(err)

// 	err = db.LoadFromStore(ctx)

// 	require.NoError(err)
// 	require.Equal(1, len(db.zones))


// 	err = db.UpdateFromFile(ctx, ts.cfg)

// 	require.NoError(err)
// 	assert.Equal(1, len(db.zones))
// }

// func (ts *DBInventoryTestSuite) TestCreateZone() {
// 	assert := ts.Assert()
// 	require := ts.Require()

// 	ts.ensureInventoryLoaded()

// 	zoneName := "zone1"

// 	ctx := context.Background()

// 	zone := &pb.DefinitionZone{
// 		Enabled: true,
// 		Condition: pb.Definition_operational,
// 		Location: "Nowhere in particular",
// 		Notes: "empty notes",
// 	}

// 	revCreate, err := ts.db.CreateZone(ctx, zoneName, zone)
// 	require.NoError(err)
// 	assert.Less(int64(0), revCreate)

// 	z, revRead, err := ts.db.ReadZone(ctx, zoneName)
// 	assert.NoError(err)
// 	assert.Equal(revCreate, revRead)
// 	require.NotNil(z)
// 	assert.Equal(zone.Enabled, z.Enabled)
// 	assert.Equal(zone.Condition, z.Condition)
// 	assert.Equal(zone.Location, z.Location)
// 	assert.Equal(zone.Notes, z.Notes)
// 	assert.Equal(0, len(z.Racks))
// }

func (ts *DBInventoryTestSuite) TestCreateRack() {

}

func (ts *DBInventoryTestSuite) TestCreateTor() {

}

func (ts *DBInventoryTestSuite) TestCreatePdu() {

}

func (ts *DBInventoryTestSuite) TestCreateBlade() {

}


func TestDBInventoryTestSuite(t *testing.T) {
	suite.Run(t, new(DBInventoryTestSuite))
}
