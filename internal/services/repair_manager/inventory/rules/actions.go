package rules

import (
	"context"
	"fmt"

	"github.com/Jim3Things/CloudChamber/internal/common"
	r "github.com/Jim3Things/CloudChamber/internal/services/repair_manager/ruler"
	"github.com/Jim3Things/CloudChamber/internal/tracing"
)

func PowerChangeBlade(ctx context.Context, args map[string]r.Term, ec *r.EvalContext) (*r.Proposal, error) {
	rackName, err := argToString("rack", args, ec)
	if err != nil {
		return nil, err
	}

	bladeID, err := argToString("blade", args, ec)
	if err != nil {
		return nil, err
	}

	power, err := argToBool("power", args, ec)
	if err != nil {
		return nil, err
	}

	// Temporary logic
	tracing.Info(ctx, "Powering %s blade %s in rack %q", common.AOrB(power, "on", "off"), bladeID, rackName)
	return &r.Proposal{
		Path:  fmt.Sprintf("racks/%s/blades/%s.powering", rackName, bladeID),
		Value: power,
	}, nil
}

func NetworkChangeBlade(ctx context.Context, args map[string]r.Term, ec *r.EvalContext) (*r.Proposal, error) {
	rackName, err := argToString("rack", args, ec)
	if err != nil {
		return nil, err
	}

	bladeID, err := argToString("blade", args, ec)
	if err != nil {
		return nil, err
	}

	connect, err := argToBool("connect", args, ec)
	if err != nil {
		return nil, err
	}

	// Temporary logic
	tracing.Info(
		ctx,
		"%s network for %s blade %s in rack %q",
		common.AOrB(connect, "Connecting", "Disconnecting"),
		bladeID,
		rackName)
	return &r.Proposal{
		Path:  fmt.Sprintf("racks/%s/blades/%s.connecting", rackName, bladeID),
		Value: connect,
	}, nil
}
func HumanIntervention(ctx context.Context, args map[string]r.Term, ec *r.EvalContext) (*r.Proposal, error) {
	message, err := argToString("message", args, ec)
	if err != nil {
		return nil, err
	}

	tracing.Info(ctx, "Requesting human intervention: %q", message)
	return &r.Proposal{
		Path:  "message",
		Value: message,
	}, nil
}

func argToString(name string, args map[string]r.Term, ec *r.EvalContext) (string, error) {
	t, ok := args[name]
	if !ok {
		return "", r.ErrMissingFieldName(name)
	}

	leaf, err := t.Evaluate(ec)
	if err != nil {
		return "", err
	}

	return leaf.AsName(ec.Replacements)
}

func argToBool(name string, args map[string]r.Term, ec *r.EvalContext) (bool, error) {
	t, ok := args[name]
	if !ok {
		return false, r.ErrMissingFieldName(name)
	}

	leaf, err := t.Evaluate(ec)
	if err != nil {
		return false, err
	}

	return leaf.AsBool()
}
