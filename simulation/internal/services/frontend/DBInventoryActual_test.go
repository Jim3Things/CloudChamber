package frontend

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Jim3Things/CloudChamber/simulation/internal/clients/inventory"
)

type InventoryActualTestSuite struct {
	testSuiteCore

	inventoryDefinitionLoaded bool

	db *DBInventory
}

func (ts *InventoryActualTestSuite) ensureInventoryLoaded() error {
	ctx := context.Background()

	if !ts.inventoryDefinitionLoaded {
		if err := ts.db.UpdateInventoryDefinition(ctx, "./testdata/standard"); err != nil {
			return err
		}

		ts.inventoryDefinitionLoaded = true

		if err := ts.db.LoadInventoryActual(true); err != nil {
			return err
		}
	}

	return nil
}

func (ts *InventoryActualTestSuite) SetupSuite() {
	require := ts.Require()

	ctx := context.Background()

	ts.testSuiteCore.SetupSuite()

	ts.db = NewDbInventory()

	err := ts.db.Initialize(ctx, ts.cfg)
	require.NoError(err)
}

func (ts *InventoryActualTestSuite) SetupTest() {
	require := ts.Require()

	_ = ts.utf.Open(ts.T())

	err := ts.ensureInventoryLoaded()
	require.NoError(err)}

func (ts *InventoryActualTestSuite) TearDownTest() {
	ts.utf.Close()
}

func (ts *InventoryActualTestSuite) TestLoadInventoryActual() {
	require := ts.Require()
	assert := ts.Assert()

	ctx := context.Background()

	zone, err := inventory.NewZone(ctx, ts.db.Store, inventory.DefinitionTable, defaultRegion, defaultZone)
	require.NoError(err)
	require.NotNil(zone)

	_, rackNames, err := zone.ListChildren(ctx)

	assert.Equal(len(ts.db.Actual.Racks), len(rackNames))

	for _, rackName := range rackNames {
		r, ok := ts.db.Actual.Racks[rackName]
		require.True(ok)

		rack, err := zone.NewChild(ctx, rackName)
		require.NoError(err)
		require.NotNil(rack)

		_, bladeIDs, err := rack.ListBlades(ctx)
		require.NoError(err)
		require.NotNil(bladeIDs)

		assert.Equal(len(bladeIDs), len(r.Blades))

		for _, bladeID := range bladeIDs {
			b, ok := r.Blades[bladeID]
			require.True(ok)

			assert.Equal(bladeWorking, b.State)
		}

		assert.Equal(rackWorking, r.State)
		assert.Equal(torWorking, r.Tor.State)
		assert.Equal(pduWorking, r.Pdu.State)
	}
}

func TestInitDBInventoryActual(t *testing.T) {
	suite.Run(t, new(InventoryActualTestSuite))
}
