package services

import (
	"fmt"
)

func (x *InventoryAddress) Describe() string {
	preamble := ""

	switch elem := x.Element.(type) {
	case *InventoryAddress_Tor:
		preamble = "the TOR"

	case *InventoryAddress_Pdu:
		preamble = "the PDU"

	case *InventoryAddress_BladeId:
		preamble = fmt.Sprintf("blade %d", elem.BladeId)
	}

	return fmt.Sprintf("%s in rack %q", preamble, x.Rack)
}
