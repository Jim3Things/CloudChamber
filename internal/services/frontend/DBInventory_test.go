package frontend

import (
	"context"
	"flag"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Jim3Things/CloudChamber/internal/clients/store"
	"github.com/Jim3Things/CloudChamber/internal/config"
	ct "github.com/Jim3Things/CloudChamber/pkg/protos/common"
	pb "github.com/Jim3Things/CloudChamber/pkg/protos/inventory"
)


var (
	cfg *config.GlobalConfig
)


func TestLoadFromStore(t *testing.T) {

	_ = utf.Open(t)
	defer utf.Close()

	configPath := flag.String("config", "./testdata", "path to the configuration file")
	flag.Parse()

	assert.Nil(t, dbInventory)

	cfg, err := config.ReadGlobalConfig(*configPath)
	require.Nil(t, err)
	require.NotNil(t, cfg)

	ctx := context.Background()

	db :=  &DBInventory{
		mutex: sync.RWMutex{},
		Zone: nil,
		MaxBladeCount: 0,
		MaxCapacity:   &ct.BladeCapacity{},
		Store: store.NewStore(),
	}

	err = db.Store.Connect()

	require.Nil(t, err)

	err = db.LoadFromStore(ctx)

	require.Nil(t, err)
	require.Equal(t, 1, len(db.zones))

	
	for _, z := range db.zones {

		assert.True(t, z.Enabled)
		assert.Equal(t, pb.Definition_operational, z.Condition)
		assert.Equal(t, 0,  len(z.Location))
		assert.Equal(t, 0,  len(z.Notes))
		
		assert.Equal(t, 8, len(z.Racks))

		for _, r := range z.Racks {
			assert.True(t, r.Enabled)
			assert.Equal(t, pb.Definition_operational, r.Condition)
			assert.Equal(t, 0,  len(r.Location))
			assert.Equal(t, 0,  len(r.Notes))
			
			assert.Equal(t, 1, len(r.Pdus))

			for _, p := range r.Pdus {
				assert.True(t, p.Enabled)
				assert.True(t, p.Powered)

				assert.Equal(t, pb.Definition_operational, p.Condition)

				assert.Equal(t, 9, len(p.Ports))
			}

			assert.Equal(t, 1, len(r.Tors))

			for _, tor := range r.Tors {
				assert.True(t, tor.Enabled)
				assert.True(t, tor.Powered)

				assert.Equal(t, pb.Definition_operational, tor.Condition)

				assert.Equal(t, 8, len(tor.Ports))
			}

			assert.Equal(t, 8, len(z.Racks))

			for _, b := range r.Blades {
				assert.True(t, b.Enabled)
				assert.Equal(t, pb.Definition_operational, b.Condition)

				assert.Equal(t, 8,     b.Capacity.Cores)
				assert.Equal(t, 16834, b.Capacity.MemoryInMb)
				assert.Equal(t, 120,   b.Capacity.DiskInGb)
				assert.Equal(t, 2048,  b.Capacity.NetworkBandwidthInMbps)
				assert.Equal(t, "X64", b.Capacity.Arch)
			}
		}
	}

}

func TestUpdateFromFile(t *testing.T) {
	_ = utf.Open(t)
	defer utf.Close()

}


func TestInitializeInventory(t *testing.T) {
	_ = utf.Open(t)
	defer utf.Close()

	configPath := flag.String("config", "./testdata", "path to the configuration file")
	flag.Parse()

	assert.Nil(t, dbInventory)

	cfg, err := config.ReadGlobalConfig(*configPath)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	db :=  &DBInventory{
		mutex: sync.RWMutex{},
		Zone: nil,
		MaxBladeCount: 0,
		MaxCapacity:   &ct.BladeCapacity{},
		Store: store.NewStore(),
	}

	err = db.Initialize(context.Background(), cfg)
	assert.NoError(t, err)
	assert.NotNil(t, dbInventory)
}

func TestCreateZone(t *testing.T) {
	_ = utf.Open(t)
	defer utf.Close()

}

func TestCreateRack(t *testing.T) {
	_ = utf.Open(t)
	defer utf.Close()

}

func TestCreateTor(t *testing.T) {
	_ = utf.Open(t)
	defer utf.Close()

}

func TestCreatePdu(t *testing.T) {
	_ = utf.Open(t)
	defer utf.Close()

}

func TestCreateBlade(t *testing.T) {
	_ = utf.Open(t)
	defer utf.Close()

}
