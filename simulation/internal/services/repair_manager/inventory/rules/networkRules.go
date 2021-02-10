package rules

import (
	r "github.com/Jim3Things/CloudChamber/simulation/internal/services/repair_manager/ruler"
)

var networkChangeArgs = map[string]r.Term{
	"rack":    r.V("%rack%"),
	"blade":   r.V("%blade%"),
	"connect": r.N("%target%/racks/%rack%/blades/%blade%.connected"),
}

var connectBlade = r.Rule{
	Name: "Connect the network to a blade",
	Where: r.All(
		r.IsTrue("%target%/racks/%rack%/blades/%blade%.connected"),
		r.NotMatch(
			r.N("%target%/racks/%rack%/blades/%blade%.connected"),
			r.N("%observed%/racks/%rack%/blades/%blade%.connected")),
	),
	Reason: "blade %blade% in rack \"%rack%\" needs to be connected to the network",
	Choices: []r.RuleChoice{
		{
			Assuming: r.IsFalse("%observed%/racks/%rack%/tor/blade_%blade%.faulted"),
			Chosen:   "tell TOR in \"%rack%\" to connect the blade",
			Rejected: "the network cable connection is faulty",
			With:     networkChangeArgs,
			Call:     NetworkChangeBlade,
		},
		{
			Assuming: r.IsTrue("%observed%/racks/%rack%/tor/blade_%blade%.faulted"),
			Chosen:   "request human intervention",
			Rejected: "the network connection is working",
			With: map[string]r.Term{
				"message": r.V("fix the network cable connection for blade %blade% in rack \"%rack%\""),
			},
			Call: HumanIntervention,
		},
	},
}

var disconnectBlade = r.Rule{
	Name: "Disconnect the network to a blade",
	Where: r.All(
		r.IsFalse("%target%/racks/%rack%/blades/%blade%.connected"),
		r.NotMatch(
			r.N("%target%/racks/%rack%/blades/%blade%.connected"),
			r.N("%observed%/racks/%rack%/blades/%blade%.connected")),
	),
	Reason: "blade %blade% in rack \"%rack%\" needs to be disconnected from the network",
	Choices: []r.RuleChoice{
		{
			Assuming: r.IsFalse("%observed%/racks/%rack%/tor/blade_%blade%.faulted"),
			Chosen:   "tell TOR in \"%rack%\" to disconnect the blade",
			Rejected: "the network cable connection is faulty",
			With:     networkChangeArgs,
			Call:     NetworkChangeBlade,
		},
		{
			Assuming: r.IsTrue("%observed%/racks/%rack%/tor/blade_%blade%.faulted"),
			Chosen:   "request human intervention",
			Rejected: "the network cable connection is working",
			With: map[string]r.Term{
				"message": r.V("fix the network cable connection for blade %blade% in rack \"%rack%\""),
			},
			Call: HumanIntervention,
		},
	},
}
