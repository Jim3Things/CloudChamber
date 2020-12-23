package ruler

// EvalContext defines the context to supply to an Evaluate call on a Term.
type EvalContext struct {
	// Replacements hold sequences of variable/replacement strings to use when
	// expanding a name string.
	Replacements []string

	// -- data access values are TBD
}

// Term is the general definition for an entry in the ruleset - either an
// intermediate Node that must be executed to get a value, or a Leaf that
// already holds a final value.
type Term interface {
	// Evaluate performs whatever operations are necessary for the Term to
	// produce the final value.
	Evaluate(ec *EvalContext) (*Leaf, error)

	// Format produces a structured string of the (sub-)tree rooted in this
	// Term instance.
	Format(indent string) string
}
