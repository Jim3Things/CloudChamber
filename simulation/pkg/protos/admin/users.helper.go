package admin

import (
	"strings"
)

func (x *Rights) Describe() string {
	var result []string

	if x.CanInjectFaults {
		result = append(result, "can inject faults")
	}

	if x.CanManageAccounts {
		result = append(result, "can manage accounts")
	}

	if x.CanModifyInventory {
		result = append(result, "can modify inventory")
	}

	if x.CanModifyWorkloads {
		result = append(result, "can modify workloads")
	}

	if x.CanPerformRepairs {
		result = append(result, "can perform repairs")
	}

	if x.CanStepTime {
		result = append(result, "can modify the simulated time")
	}

	if len(result) == 0 {
		return ""
	}

	return strings.Join(result, ", ")
}

func (x *Rights) StrongerThan(r *Rights) bool {
	if x.CanManageAccounts {
		return true
	}

	return x.CanStepTime && r.CanStepTime == x.CanStepTime &&
		x.CanPerformRepairs && r.CanPerformRepairs == x.CanPerformRepairs &&
		x.CanModifyWorkloads && r.CanModifyWorkloads == x.CanModifyWorkloads &&
		x.CanModifyInventory && r.CanModifyInventory == x.CanModifyInventory &&
		x.CanInjectFaults && r.CanInjectFaults == x.CanInjectFaults
}

func (x *User) Update() bool {
	if x.CanManageAccounts {
		// This is using the old form of rights, so set the rights correctly and
		// clear this flag.
		x.Rights.CanManageAccounts = true
		x.Rights.CanInjectFaults = true
		x.Rights.CanModifyInventory = true
		x.Rights.CanModifyWorkloads = true
		x.Rights.CanPerformRepairs = true
		x.Rights.CanStepTime = true
		x.CanManageAccounts = false

		return true
	}

	return false
}
