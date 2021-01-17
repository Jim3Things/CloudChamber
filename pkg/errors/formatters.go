package errors

import (
	"fmt"
)

// This module contains helper functions that ensure consistent formatting for
// common data types

func bladeAddress(zone string, rack string, bladeID int64) string {
	return fmt.Sprintf("blade %d in zone %q, rack %q", bladeID, zone, rack)
}

func pduAddress(zone string, rack string, pdu int64) string {
	return fmt.Sprintf("pdu %d in zone %q, rack %q", pdu, zone, rack)
}

func torAddress(zone string, rack string, tor int64) string {
	return fmt.Sprintf("tor %d in zone %q, rack %q", tor, zone, rack)
}
