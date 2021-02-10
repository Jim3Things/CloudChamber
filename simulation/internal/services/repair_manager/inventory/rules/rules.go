package rules

// This module contains the inventory repair rule definitions.

import (
	r "github.com/Jim3Things/CloudChamber/simulation/internal/services/repair_manager/ruler"
)

// Rules contains the full set of rules to process for a repair request.
var Rules = []r.Rule{
	powerOnBlade,
	powerOffBlade,

	connectBlade,
	disconnectBlade,
}
