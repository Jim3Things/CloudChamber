package rules

import (
	"github.com/Jim3Things/CloudChamber/simulation/internal/services/repair_manager/ruler"
)

var powerChangeArgs = map[string]ruler.Term{
	"rack":  ruler.V("%rack%"),
	"blade": ruler.V("%blade%"),
	"power": ruler.N("%target%/racks/%rack%/blades/%blade%.power"),
}

var powerOnBlade = ruler.Rule{
	Name: "Power on a blade",
	Where: ruler.All(
		ruler.IsTrue("%target%/racks/%rack%/blades/%blade%.power"),
		ruler.NotMatch(
			ruler.N("%target%/racks/%rack%/blades/%blade%.power"),
			ruler.N("%observed%/racks/%rack%/blades/%blade%.power")),
	),
	Reason: "blade %blade% in rack \"%rack%\" needs to be powered on",
	Choices: []ruler.RuleChoice{
		{
			Assuming: ruler.All(
				ruler.IsFalse("%observed%/racks/%rack%/blades/%blade%.powering"),
				ruler.IsFalse("%observed%/racks/%rack%/pdu/blade_%blade%.faulted"),
			),
			Chosen:   "tell PDU in \"%rack%\" to power the blade on",
			Rejected: "it may be powering on already or the power connection is faulty",
			With:     powerChangeArgs,
			Call:     PowerChangeBlade,
		},
		{
			Assuming: ruler.IsTrue("%observed%/racks/%rack%/pdu/blade_%blade%.faulted"),
			Chosen:   "since the power connection is faulty, request human intervention",
			Rejected: "the power connection is working",
			With: map[string]ruler.Term{
				"message": ruler.V("fix the power connection for blade %blade% in rack \"%rack%\""),
			},
			Call: HumanIntervention,
		},
	},
}

var powerOffBlade = ruler.Rule{
	Name: "Power off a blade",
	Where: ruler.All(
		ruler.IsFalse("%target%/racks/%rack%/blades/%blade%.power"),
		ruler.NotMatch(
			ruler.N("%target%/racks/%rack%/blades/%blade%.power"),
			ruler.N("%observed%/racks/%rack%/blades/%blade%.power")),
	),
	Reason: "blade %blade% in rack \"%rack%\" needs to be powered off",
	Choices: []ruler.RuleChoice{
		{
			Assuming: ruler.IsFalse("%observed%/racks/%rack%/pdu/blade_%blade%.faulted"),
			Chosen:   "tell PDU in \"%rack%\" to power the blade off",
			Rejected: "the power connection is faulty",
			With:     powerChangeArgs,
			Call:     PowerChangeBlade,
		},
		{
			Assuming: ruler.IsTrue("%observed%/racks/%rack%/pdu/blade_%blade%.faulted"),
			Chosen:   "request human intervention",
			Rejected: "the power connection is working",
			With: map[string]ruler.Term{
				"message": ruler.V("fix the power connection for blade %blade% in rack \"%rack%\""),
			},
			Call: HumanIntervention,
		},
	},
}
