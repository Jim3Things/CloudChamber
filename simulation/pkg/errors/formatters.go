package errors

import (
	"fmt"
)

// This module contains helper functions that ensure consistent formatting for
// common data types

func regionAddress(region string) string {
	return fmt.Sprintf("region %q", region)
}

func zoneAddress(region string, zone string) string {
	return fmt.Sprintf("zone %q in region %q", zone, region)
}

func rackAddress(region string, zone string, rack string) string {
	return fmt.Sprintf("rack %q in region %q, zone %q", rack, region, zone)
}

func bladeAddress(region string, zone string, rack string, bladeID int64) string {
	return fmt.Sprintf("blade %d in region %q, zone %q, rack %q", bladeID, region, zone, rack)
}

func bladeAddressName(region string, zone string, rack string, bladeID string) string {
	return fmt.Sprintf("blade %q in region %q, zone %q, rack %q", bladeID, region, zone, rack)
}

func pduAddress(region string, zone string, rack string, pdu int64) string {
	return fmt.Sprintf("pdu %d in region %q, zone %q, rack %q", pdu, region, zone, rack)
}

func pduAddressName(region string, zone string, rack string, pdu string) string {
	return fmt.Sprintf("pdu %q in region %q, zone %q, rack %q", pdu, region, zone, rack)
}

func torAddress(region string, zone string, rack string, tor int64) string {
	return fmt.Sprintf("tor %d in region %q, zone %q, rack %q", tor, region, zone, rack)
}

func torAddressName(region string, zone string, rack string, tor string) string {
	return fmt.Sprintf("tor %q in region %q, zone %q, rack %q", tor, region, zone, rack)
}
