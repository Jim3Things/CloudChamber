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

type ActualTor struct{
	State torStatus
}

type ActualPdu struct{
	State pduStatus
}

type ActualBlade struct{
	State bladeStatus
}

type ActualRack struct{
	Tor ActualTor
	Pdu ActualPdu
	Blades map[int64]*ActualBlade

	State rackStatus
}

type ActualZone struct {
	Racks map[string]*ActualRack
}

type DBInventoryActual struct {
	Mutex sync.Mutex

	Zone *ActualZone
}

var dbInventoryActual *DBInventoryActual

func InitDBInventoryActual(inven *DBInventory) error {
	if dbInventoryActual == nil{
		actual:= &DBInventoryActual{
			Mutex: sync.Mutex{},
			Zone: &ActualZone{
				Racks: make(map[string]*ActualRack),
			},
		}
		inven.Mutex.Lock()
		for name, rack := range inven.Zone.Racks{
			r := &ActualRack {
				Tor: ActualTor{
					State: torWorking,
				},
				Pdu: ActualPdu{
					State: pduWorking,
				},
				Blades: make(map[int64]*ActualBlade),
				State: rackWorking,
			}
			for bladeID := range rack.Blades{
				r.Blades[bladeID] = &ActualBlade {
					State: bladeWorking,
				}
			}
			actual.Zone.Racks[name] = r
		}
		inven.Mutex.Unlock()
		dbInventoryActual = actual
	}
	return nil
}
