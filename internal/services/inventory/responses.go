package inventory

// This file contains helper functions that simplify the creation of response
// messages to repair operations.

import (
	ct "github.com/Jim3Things/CloudChamber/pkg/protos/common"
	"github.com/Jim3Things/CloudChamber/pkg/protos/services"
)

// droppedResponse constructs a dropped response message with the correct time
// and target.
func droppedResponse(
	target *services.InventoryAddress,
	occursAt int64) *services.InventoryRepairResp {
	return &services.InventoryRepairResp{
		Source: target,
		At:     &ct.Timestamp{Ticks: occursAt},
		Rsp:    &services.InventoryRepairResp_Dropped{},
	}
}

// failedResponse constructs a failure response message with the correct time,
// target, and reason.
func failedResponse(
	target *services.InventoryAddress,
	occursAt int64,
	msg string) *services.InventoryRepairResp {
	return &services.InventoryRepairResp{
		Source: target,
		At:     &ct.Timestamp{Ticks: occursAt},
		Rsp:    &services.InventoryRepairResp_Failed{
			Failed: msg,
		},
	}
}

// successResponse constructs a success response message with the correct time
// and target.
func successResponse(
	target *services.InventoryAddress,
	occursAt int64) *services.InventoryRepairResp {
	return &services.InventoryRepairResp{
		Source: target,
		At:     &ct.Timestamp{Ticks: occursAt},
		Rsp:    &services.InventoryRepairResp_Success{},
	}
}
