package frontend

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type InventoryActualTestSuite struct {
	testSuiteCore
}

func (ts *InventoryActualTestSuite) TestInitActual() {
	require := ts.Require()
	assert := ts.Assert()

	assert.Equal(len(dbInventory.Zone.Racks), len(dbInventoryActual.Zone.Racks))

	for name, rack := range dbInventory.Zone.Racks {
		r, ok := dbInventoryActual.Zone.Racks[name]
		require.True(ok)

		assert.Equal(len(rack.Blades), len(r.Blades))

		for bladeID := range rack.Blades {
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
