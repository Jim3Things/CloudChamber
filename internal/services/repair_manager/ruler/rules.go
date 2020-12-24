package ruler

import (
	"errors"
)

// Error definitions.  Placeholder for now.

var ErrInvalidType = errors.New("invalid type")
var ErrInvalidOp = errors.New("invalid operation")
var ErrInvalidArgLen = errors.New("invalid number of arguments")

// Define the rules definition layout

// Proposal is the placeholder structure for the result from calling the
// output function
type Proposal struct {
}

// Arg contains a context item to use when preparing the final output
type Arg struct {
	Name string
	From *Term
}

// OutputFunc is the signature for a function that generates the result of a
// ruleset matching.
type OutputFunc func(args []Arg) (*Proposal, error)

// Rule defines a matching rule to evaluate and, if matched, execute.
type Rule struct {
	// Where is the initial trigger test. If this evaluates to true, then the
	// various choices are evaluated until one matches.
	Where Term

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

	// With contains the set of argument and context state to provide to the
	// output function
	With []Arg

	// Call is the function to call when this choice is taken.  The resulting
	// Proposal is then returned as the evaluation output for the Rule.
	Call OutputFunc
}

// [ { Where: NotMatch(N("first/%blade%/state"), N("second/%blade%/state")),
//	   Choices: []RuleChoice {
//			{ Assuming: All(
//     			Match(N("second/%pdu%/power"), V(true)),
//				Match(N("second/%pdu%/cables/%blade%/power"), V(true)),
//				Match(N("second/%tor%/state"), V(torWorking)),
//				Match(N("second/%tor%/cables/%blade%/connected"), V(true)),
//				Match(N("second/%blade%/booting"), V(false)),
//	   		),
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
	switch v := value.(type) {
	case bool:
		return NewLeafBool(v)

	case int32:
		return NewLeafInt32(v)

	case int:
		return NewLeafInt32(int32(v))

	case int64:
		return NewLeafInt64(v)

	case string:
		return NewLeafString(v)
	}

	return nil
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
