package ruler

import (
	"context"
	"fmt"
	"strings"

	"github.com/Jim3Things/CloudChamber/internal/clients/timestamp"
	"github.com/Jim3Things/CloudChamber/internal/tracing"
)

// Error definitions.  Placeholder for now.

type ErrInvalidType ValueType

func (e ErrInvalidType) Error() string {
	return fmt.Sprintf("unexpected value type %d encountered", int(e))
}

type ErrInvalidOp OpType

func (e ErrInvalidOp) Error() string {
	return fmt.Sprintf("unexpected operation %d encountered", int(e))
}

type ErrInvalidArgLen struct {
	op       string
	required string
	actual   int
}

func (e ErrInvalidArgLen) Error() string {
	return fmt.Sprintf("operation %s expects %s but received %d", e.op, e.required, e.actual)
}

type ErrMissingFieldName string

func (e ErrMissingFieldName) Error() string {
	return fmt.Sprintf(
		"key must have a table name, one or more path elements, and one field name.  "+
			"no field name was found in %q.", string(e))
}

type ErrExtraFieldNames string

func (e ErrExtraFieldNames) Error() string {
	return fmt.Sprintf(
		"key must have a table name, one or more path elements, and one field name.  "+
			"multiple possible field were names found in %q.",
		string(e))
}

type ErrMissingPath string

func (e ErrMissingPath) Error() string {
	return fmt.Sprintf(
		"key must have a table name, one or more path elements, and one field name.  "+
			"No path elements were found in %q.",
		string(e))
}

// Define the rules definition layout

// Proposal is the placeholder structure for the result from calling the
// output function
type Proposal struct {
	Path  string
	Value interface{}
}

// OutputFunc is the signature for a function that generates the result of a
// ruleset matching.
type OutputFunc func(ctx context.Context, args map[string]Term, ec *EvalContext) (*Proposal, error)

// Rule defines a matching rule to evaluate and, if matched, execute.
type Rule struct {
	// Name is a friendly string name used in the construction of the span name
	Name string

	// Where is the initial trigger test. If this evaluates to true, then the
	// various choices are evaluated until one matches.
	Where Term

	// Reason is the text for the log's reason attribute explaining why this
	// rule is evaluated.
	Reason string

	// Choices are the set of conditional subtests, which are listed in order
	// of decreasing priority.  This enables contextual operation, such as
	// repair action escalation.
	Choices []RuleChoice
}

// RuleChoice defines a single subtest for a Rule.
type RuleChoice struct {
	// Assuming contains the enabling test.  If this evaluates to true, then
	// this choice is taken, and this choice's output function is called.
	Assuming Term

	// Chosen is the supplemental reason text explaining why this choice was
	// picked.
	Chosen string

	// Rejected is the supplemental reason text explaining why this choice was
	// not picked.
	Rejected string

	// With contains the set of argument and context state to provide to the
	// output function
	With map[string]Term

	// Call is the function to call when this choice is taken.  The resulting
	// Proposal is then returned as the evaluation output for the Rule.
	Call OutputFunc
}

// [ { Where: All(
//			NotMatch(N("first/%blade%/state"), N("second/%blade%/state")),
//			Match(N("second/%pdu%/power"), V(true)),
//			Match(N("second/%pdu%/cables/%blade%/power"), V(true)),
//			Match(N("second/%tor%/state"), V(torWorking)),
//			Match(N("second/%tor%/cables/%blade%/connected"), V(true)),
//			Match(N("second/%blade%/booting"), V(false)),
//     Reason: "blade needs to be booted, it is bootable, and it has not been booted",
//	   Choices: []RuleChoice {
//			{ Assuming: All(
//				NotMatch(N("second/%blade%/repair"), V("boot"))
//	   		),
//			Chosen: "no previous attempt to boot",
//			Rejected: "already tried to boot",
//	   		With: []Arg {
//	  			{Name:"blade", From: N(%blade%)},
//	  			{Name:"boot", From:V(true)},
//	   		},
//     		Call: createBootBladeRepair },
//		},
// ]

// +++ Rule API functions
// This section contains functions that simplify creating a rule set

// N creates the Terms required to specify that the data associated with the
// supplied key is to be used.
func N(key string) Term {
	return NewNodeFetch(NewLeafString(key))
}

// V creates a Term holding the specified value.
func V(value interface{}) Term {
	return NewLeaf(value)
}

// IsFalse creates the Terms to fetch the data associated with the string, and
// test that its value is false.
func IsFalse(s string) Term {
	return NewNodeMatch(N(s), NewLeafBool(false))
}

// IsTrue creates the Terms to fetch the data associated with the string, and
// test that its value is true.
func IsTrue(s string) Term {
	return NewNodeMatch(N(s), NewLeafBool(true))
}

// Match creates the Terms to hold a Match test
func Match(l Term, r Term) Term {
	return NewNodeMatch(l, r)
}

// NotMatch creates the Terms to hold a NotMatch test
func NotMatch(l Term, r Term) Term {
	return NewNodeNotMatch(l, r)
}

// All creates the Terms to hold a All test
func All(terms ...Term) Term {
	return NewNodeAll(terms...)
}

// Any creates the Terms to hold a Any test
func Any(terms ...Term) Term {
	return NewNodeAny(terms...)
}

// --- Rule API functions

// Process is the main entry into the rules processing.  It processes each
// rule, looking for a match.  When it finds one, it adds it to the set of
// proposals.  When it gets an error processing, it stops and immediately
// returns.
//
// -- Should it return on the first proposal match?  If it gets an error
//    should it return the previously completed proposals?
func Process(ctx context.Context, rules []Rule, tables Tables, vars map[string]string) ([]*Proposal, error) {
	ec := &EvalContext{
		Replacements: varsToReplacements(vars),
		Tables:       tables,
	}

	r := strings.NewReplacer(ec.Replacements...)

	var proposals []*Proposal

	for _, rule := range rules {
		reason := ""

		v, err := processTerm(rule.Where, ec)
		if err != nil {
			return nil, err
		}

		but := "\nbut "
		so := "\nso "
		if v {
			reason = r.Replace(rule.Reason)
			anyChosen := false

			for _, choice := range rule.Choices {
				chosen, err := processTerm(choice.Assuming, ec)
				if err != nil {
					return nil, err
				}

				if chosen {
					anyChosen = true
					reason = fmt.Sprintf("%s%s%s", reason, so, r.Replace(choice.Chosen))
					ctx, span := tracing.StartSpan(
						ctx,
						tracing.WithName(fmt.Sprintf("Executing %q rule", rule.Name)),
						tracing.WithReason(reason),
						tracing.WithContextValue(timestamp.EnsureTickInContext))

					p, err := choice.Call(ctx, choice.With, ec)
					if err != nil {
						return nil, err
					}
					proposals = append(proposals, p)

					span.End()
					break
				} else {

					reason = fmt.Sprintf("%s%s%s", reason, but, r.Replace(choice.Rejected))
					but = ",\nand "
					so = ",\nso therefore "
				}
			}

			if !anyChosen {
				ctx, span := tracing.StartSpan(
					ctx,
					tracing.WithName(fmt.Sprintf("Executing %q rule", rule.Name)),
					tracing.WithReason(reason),
					tracing.WithContextValue(timestamp.EnsureTickInContext))
				tracing.Info(ctx, "No choice found to execute")
				span.End()
			}
		}
	}

	return proposals, nil
}

func varsToReplacements(vars map[string]string) []string {
	var r []string
	for k, v := range vars {
		r = append(r, k, v)
	}

	return r
}

func processTerm(where Term, ec *EvalContext) (bool, error) {
	res, err := where.Evaluate(ec)
	if err != nil {
		return false, err
	}

	return res.AsBool()
}
