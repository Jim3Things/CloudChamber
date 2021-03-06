package frontend

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Jim3Things/CloudChamber/simulation/internal/clients/inventory"
	"github.com/Jim3Things/CloudChamber/simulation/internal/clients/namespace"
)

type InventoryActualTestSuite struct {
	testSuiteCore

	db *DBInventory
}

func (ts *InventoryActualTestSuite) SetupSuite() {
	ts.testSuiteCore.SetupSuite()

	// The standard "frontend" initialisation will create a dbInventory structure
	// which will lead to the initialization of the inventory within the store.
	// This means we can just use the global store as long as we remember that
	// any records written here will persist for this test session and so the
	// names use should not conflict with those being used in the standard
	// inventory definition file.
	//
	ts.db = dbInventory
}

func (ts *InventoryActualTestSuite) SetupTest() {
	_ = ts.utf.Open(ts.T())
}

func (ts *InventoryActualTestSuite) TearDownTest() {
	ts.utf.Close()
}

func (ts *InventoryActualTestSuite) TestLoadInventoryActual() {
	require := ts.Require()

	ctx := context.Background()

	zone, err := ts.db.inventory.NewZone(namespace.DefinitionTable, inventory.DefaultRegion, inventory.DefaultZone)
	require.NoError(err)
	require.NotNil(zone)

	_, rackNames, err := zone.ListChildren(ctx)
	require.NoError(err)
	require.NotNil(rackNames)

	require.Equal(len(ts.db.Actual.Racks), len(rackNames))

	for _, rackName := range rackNames {
		r, ok := ts.db.Actual.Racks[rackName]
		require.True(ok)

		rack, err := zone.NewChild(rackName)
		require.NoError(err)
		require.NotNil(rack)

		_, bladeIDs, err := rack.ListBlades(ctx)
		require.NoError(err)
		require.NotNil(bladeIDs)

		require.Equal(len(bladeIDs), len(r.Blades))

		for _, bladeID := range bladeIDs {
			b, ok := r.Blades[bladeID]
			require.True(ok)

			require.Equal(bladeWorking, b.State)
		}

		require.Equal(rackWorking, r.State)
		require.Equal(torWorking, r.Tor.State)
		require.Equal(pduWorking, r.Pdu.State)
	}
}

func TestInitDBInventoryActual(t *testing.T) {
	suite.Run(t, new(InventoryActualTestSuite))
}
