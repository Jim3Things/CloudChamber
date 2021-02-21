package admin

import (
	"strings"
)

// Describe is a function that returns a string containing a comma separated
// list of enabled rights.
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

// StrongerThan verifies that the new set of rights, supplied as r, does not
// grant any rights that the current Rights do not themselves hold.  If the
// current rights include CanManageAccounts, then all new Rights are allowed,
// as the current set are able to escalate its rights at will.
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

// Update ensures that the user record is migrated to the latest schema.  It
// returns true if the record was modified; false, if it was already consistent
// with the version.
func (x *User) Update() bool {
	if x.CanManageAccounts {
		// This is using the old form of rights, so set the rights correctly and
		// clear this flag.
		x.Rights = &Rights{
			CanManageAccounts:  true,
			CanStepTime:        true,
			CanModifyWorkloads: true,
			CanModifyInventory: true,
			CanInjectFaults:    true,
			CanPerformRepairs:  true,
		}

		x.CanManageAccounts = false

		return true
	}

	return false
}

// FixMissingFields replaces nil pointers to sub-structs with instances set to
// their default values.  This fixes the result of JSON's omitempty attributes.
func (x *User) FixMissingFields() {
	if x.Rights == nil {
		x.Rights = &Rights{}
	}
}

// FixMissingFields replaces nil pointers to sub-structs with instances set to
// their default values.  This fixes the result of JSON's omitempty attributes.
func (x *UserUpdate) FixMissingFields() {
	if x.Rights == nil {
		x.Rights = &Rights{}
	}
}

// FixMissingFields replaces nil pointers to sub-structs with instances set to
// their default values.  This fixes the result of JSON's omitempty attributes.
func (x *UserDefinition) FixMissingFields() {
	if x.Rights == nil {
		x.Rights = &Rights{}
	}
}
