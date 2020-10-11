package frontend

import (
	"github.com/stretchr/testify/require"
	"testing"
	"github.com/stretchr/testify/assert"
	)

func TestInitDBInventoryActual(t *testing.T){

 	err := InitDBInventoryActual(dbInventory)
	require.Nil(t, err)

	assert.Equal(t, len(dbInventory.Zone.Racks), len(dbInventoryActual.Zone.Racks))

	for name, rack := range dbInventory.Zone.Racks{
		r, ok := dbInventoryActual.Zone.Racks[name]
		require.True(t, ok)
		
		assert.Equal(t, len(rack.Blades), len(r.Blades))
		for bladeID := range rack.Blades{
			b, ok := r.Blades[bladeID]
			require.True(t, ok)
			assert.Equal(t, bladeWorking, b.State)
		}
		
		assert.Equal(t, rackWorking, r.State)
		assert.Equal(t, torWorking, r.Tor.State)
		assert.Equal(t, pduWorking, r.Pdu.State)

	}
			
}

 