package ruler

import (
	"fmt"
)

type OpType int

const (
	OpInvalid OpType = iota
	OpFetch
	OpMatch
	OpNotMatch
	OpAll
	OpAny
)

const (
	bump = "    "
)

// Node holds a computation directive, which requires at least one child term
// to complete.  That child term may be another Node, or a Leaf.
type Node struct {
	Op   OpType
	Args []Term
}

// Two key Node functions are table driven, based on the operation type.  The
// table structure and contents are defined here.

// evalFunc is the signature for a function that implements the Evaluate logic
// for a particular operation type.
type evalFunc func(args []Term, ec *EvalContext) (*Leaf, error)

// opEntry defines the structure of an entry in the per-operation table.  This
// consists of the text string for the operation, used in Format, and a pointer
// to the specific Evaluate worker function.
type opEntry struct {
	name string
	eval evalFunc
}

var opToEntry = map[OpType]*opEntry{
	OpFetch: {
		name: "Fetch",
		eval: doOpFetch,
	},
	OpMatch: {
		name: "Match",
		eval: doOpMatch,
	},
	OpNotMatch: {
		name: "Not Match",
		eval: doOpNotMatch,
	},
	OpAll: {
		name: "All Match",
		eval: doOpAll,
	},
	OpAny: {
		name: "At Least One Matches",
		eval: doOpAny,
	},
}

// Evaluate executes the calculation defined by this Node.  It returns the
// final result as a Leaf item, or an error, if the calculation fails.
func (n *Node) Evaluate(ec *EvalContext) (*Leaf, error) {
	item, ok := opToEntry[n.Op]
	if !ok {
		return nil, ErrInvalidOp(n.Op)
	}

	return item.eval(n.Args, ec)
}

// Format returns a structured string representation of the Node and its
// children.
func (n *Node) Format(indent string) string {
	res := fmt.Sprintf("%s%s", indent, n.opString())
	for _, arg := range n.Args {
		res = fmt.Sprintf("%s\n%s", res, arg.Format(indent+bump))
	}

	return res
}

// +++ Constructor functions

// NewNodeFetch creates a new Node instance with that fetches data from a
// supplied table using the specified key (after variable expansion).
func NewNodeFetch(name Term) *Node {
	return &Node{
		Op:   OpFetch,
		Args: []Term{name},
	}
}

// NewNodeMatch creates a new Node instance that tests whether the two child
// Term instances are equivalent.  It returns that in a boolean Leaf instance,
// or an error if the test failed.
func NewNodeMatch(left Term, right Term) *Node {
	return &Node{
		Op:   OpMatch,
		Args: []Term{left, right},
	}
}

// NewNodeNotMatch creates a new Node instance that tests whether the two
// child Term instances differ.  It returns that in a boolean Leaf instance,
// or an error if the test failed.
func NewNodeNotMatch(left Term, right Term) *Node {
	return &Node{
		Op:   OpNotMatch,
		Args: []Term{left, right},
	}
}

// NewNodeAll creates a Node instance that tests whether all the child
// instances are logically true.  It returns that in a boolean Leaf instance,
// // or an error if the test failed.
func NewNodeAll(terms ...Term) *Node {
	return &Node{
		Op:   OpAll,
		Args: terms,
	}
}

// NewNodeAny creates a Node instance that tests whether any of the child
// instances are logically true.  It returns that in a boolean Leaf instance,
// or an error if the test failed.
func NewNodeAny(terms ...Term) *Node {
	return &Node{
		Op:   OpAny,
		Args: terms,
	}
}

// --- Constructor functions

// +++ Evaluate helper functions

func doOpFetch(args []Term, ec *EvalContext) (*Leaf, error) {
	if len(args) != 1 {
		return nil, ErrInvalidArgLen{
			op:       "Fetch",
			required: "exactly 1 argument",
			actual:   len(args),
		}
	}

	leaf, err := args[0].Evaluate(ec)
	if err != nil {
		return nil, err
	}

	// need to fill in what to do with the key.
	path, err := leaf.AsName(ec.Replacements)
	if err != nil {
		return nil, err
	}

	key, err := MakeKey(path)
	if err != nil {
		return nil, err
	}

	table, err := ec.Tables.GetTable(key)
	if err != nil {
		return nil, err
	}

	v, err := table.GetValue(key)
	if err != nil {
		return nil, err
	}

	l := NewLeaf(v)
	if l == nil {
		return nil, ErrInvalidType(ValueInvalid)
	}

	return l, nil
}

func doOpMatch(args []Term, ec *EvalContext) (*Leaf, error) {
	if len(args) != 2 {
		return nil, ErrInvalidArgLen{
			op:       "Match",
			required: "exactly 2 arguments",
			actual:   len(args),
		}
	}

	lTerm := args[0]
	rTerm := args[1]

	lLeaf, err := lTerm.Evaluate(ec)
	if err != nil {
		return nil, err
	}

	rLeaf, err := rTerm.Evaluate(ec)
	if err != nil {
		return nil, err
	}

	return equals(lLeaf, rLeaf)
}

func doOpNotMatch(args []Term, ec *EvalContext) (*Leaf, error) {
	if len(args) != 2 {
		return nil, ErrInvalidArgLen{
			op:       "NotMatch",
			required: "exactly 2 arguments",
			actual:   len(args),
		}
	}

	lTerm := args[0]
	rTerm := args[1]

	lLeaf, err := lTerm.Evaluate(ec)
	if err != nil {
		return nil, err
	}

	rLeaf, err := rTerm.Evaluate(ec)
	if err != nil {
		return nil, err
	}

	return notEquals(lLeaf, rLeaf)
}

func doOpAll(args []Term, ec *EvalContext) (*Leaf, error) {
	if len(args) <= 0 {
		return nil, ErrInvalidArgLen{
			op:       "All",
			required: "at least 1 argument",
			actual:   len(args),
		}
	}

	for _, arg := range args {
		leaf, err := arg.Evaluate(ec)
		if err != nil {
			return nil, err
		}

		v, err := leaf.AsBool()
		if err != nil {
			return nil, err
		}

		if !v {
			return NewLeafBool(false), nil
		}
	}

	return NewLeafBool(true), nil
}

func doOpAny(args []Term, ec *EvalContext) (*Leaf, error) {
	if len(args) <= 0 {
		return nil, ErrInvalidArgLen{
			op:       "Any",
			required: "at least 1 argument",
			actual:   len(args),
		}
	}

	for _, arg := range args {
		leaf, err := arg.Evaluate(ec)
		if err != nil {
			return nil, err
		}

		v, err := leaf.AsBool()
		if err != nil {
			return nil, err
		}

		if v {
			return NewLeafBool(true), nil
		}
	}

	return NewLeafBool(false), nil
}

// --- Evaluate helper functions

func equals(l *Leaf, r *Leaf) (*Leaf, error) {
	if l.vtype == r.vtype {
		return NewLeafBool(l.numVal == r.numVal && l.strVal == r.strVal), nil
	}

	lv, err := l.AsString()
	if err != nil {
		return nil, err
	}

	rv, err := r.AsString()
	if err != nil {
		return nil, err
	}

	return NewLeafBool(lv == rv), nil
}

func notEquals(l *Leaf, r *Leaf) (*Leaf, error) {
	if l.vtype == r.vtype {
		return NewLeafBool(l.numVal != r.numVal || l.strVal != r.strVal), nil
	}

	lv, err := l.AsString()
	if err != nil {
		return nil, err
	}

	rv, err := r.AsString()
	if err != nil {
		return nil, err
	}

	return NewLeafBool(lv != rv), nil
}

func (n *Node) opString() string {
	s, ok := opToEntry[n.Op]
	if !ok {
		return "Invalid"
	}

	return s.name
}
