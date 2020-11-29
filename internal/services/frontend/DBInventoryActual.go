package frontend

import (
	"sync"
)

type rackStatus int
const(
	rackWorking rackStatus = iota
	rackFaulted
)

type torStatus int
const(
	torWorking torStatus = iota
	torFaulted
)

type pduStatus int
const(
	pduWorking pduStatus = iota
	pduFaulted
)

type bladeStatus int
const(
	bladeWorking bladeStatus = iota
	bladeFaulted
)

type actualTor struct{
	State torStatus
}

type actualPdu struct{
	State pduStatus
}

type actualBlade struct{
	State bladeStatus
}

type actualRack struct{
	Tor actualTor
	Pdu actualPdu
	Blades map[int64]*actualBlade

	State rackStatus
}

type actualZone struct {
	Racks map[string]*actualRack
}

// DBInventoryActual holds the actual state of the defined inventory.
type DBInventoryActual struct {
	Mutex sync.Mutex

	Zone *actualZone
}

var dbInventoryActual *DBInventoryActual

// InitDBInventoryActual creates the starting actual state for the defined inventory.
func InitDBInventoryActual(inven *DBInventory) error {
	
	if dbInventoryActual == nil {
		actual := &DBInventoryActual {
			Mutex: sync.Mutex{},
			Zone: &actualZone {
				Racks: make(map[string]*actualRack),
			},
		}
		
		inven.mutex.RLock()

		for name, rack := range inven.Zone.Racks {
			r := &actualRack {
				Tor: actualTor {
					State: torWorking,
				},

				Pdu: actualPdu {
					State: pduWorking,
				},

				Blades: make(map[int64]*actualBlade),
				State: rackWorking,
			}

			for bladeID := range rack.Blades{
				r.Blades[bladeID] = &actualBlade {
					State: bladeWorking,
				}
			}

			actual.Zone.Racks[name] = r
		}

		inven.mutex.RUnlock()

		dbInventoryActual = actual
	}

	return nil
}
