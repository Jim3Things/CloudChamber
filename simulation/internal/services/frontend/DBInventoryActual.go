package frontend

import (
	"context"

	"github.com/Jim3Things/CloudChamber/simulation/internal/clients/inventory"
)

type rackStatus int

const (
	rackWorking rackStatus = iota
	rackFaulted
)

type torStatus int

const (
	torWorking torStatus = iota
	torFaulted
)

type pduStatus int

const (
	pduWorking pduStatus = iota
	pduFaulted
)

type bladeStatus int

const (
	bladeWorking bladeStatus = iota
	bladeFaulted
)

type actualTor struct {
	State torStatus
}

type actualPdu struct {
	State pduStatus
}

type actualBlade struct {
	State bladeStatus
}

type actualRack struct {
	Tor    actualTor
	Pdu    actualPdu
	Blades map[int64]*actualBlade

	State rackStatus
}

type actualZone struct {
	Racks map[string]*actualRack
}

// LoadInventoryActual creates the starting actual state for the defined inventory.
//
// This uses the existing inventory definition entries from the store to create an
// in-memory "actual" state. Eventually, we expect this to move to the store once
// all the required store features are in place, primarily Watch().
//
func (m *DBInventory) LoadInventoryActual(force bool) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	ctx := context.Background()

	if !m.actualLoaded || force  {

		m.mutex.Unlock()

		actual := &actualZone{Racks: make(map[string]*actualRack)}

		zone, err := inventory.NewZone(ctx, m.Store, inventory.DefinitionTable, defaultRegion, defaultZone)
		if err != nil {
			return err
		}

		_, rackNames, err := zone.ListChildren(ctx)

		for _, rackName := range rackNames {
			r := &actualRack{
				Tor: actualTor{
					State: torWorking,
				},

				Pdu: actualPdu{
					State: pduWorking,
				},

				Blades: make(map[int64]*actualBlade),
				State:  rackWorking,
			}

			rack, err := zone.NewChild(ctx, rackName)
			if err != nil {
				return err
			}
	
			_, bladeIDs, err := rack.ListBlades(ctx)

			for _, bladeID := range bladeIDs {
				r.Blades[bladeID] = &actualBlade{
					State: bladeWorking,
				}
			}

			actual.Racks[rackName] = r
		}

		m.mutex.Lock()

		if !m.actualLoaded {
			m.Actual = actual
			m.actualLoaded = true
		}
	}

	return nil
}
