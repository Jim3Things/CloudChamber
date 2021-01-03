package ruler

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Jim3Things/CloudChamber/pkg/errors"
)

type NodeTestSuite struct {
	suite.Suite
}

func (ts *NodeTestSuite) TestOpType() {
	require := ts.Require()

	n := NewNodeFetch(NewLeafString("test"))
	require.Equal(OpFetch, n.Op)

	n = NewNodeMatch(NewLeafInt64(1), NewLeafInt64(2))
	require.Equal(OpMatch, n.Op)

	n = NewNodeNotMatch(NewLeafInt64(1), NewLeafInt64(2))
	require.Equal(OpNotMatch, n.Op)

	n = NewNodeAll(NewLeafInt32(1), NewLeafInt32(1))
	require.Equal(OpAll, n.Op)

	n = NewNodeAny(NewLeafInt32(1), NewLeafInt32(1))
	require.Equal(OpAny, n.Op)
}

func (ts *NodeTestSuite) TestFetchFormat() {
	assert := ts.Assert()

	n := NewNodeFetch(NewLeafString("test"))
	s := n.Format("")
	assert.Equal(
		"Fetch\n"+
			"    [Leaf: string] \"test\"",
		s)
}

func (ts *NodeTestSuite) TestMatchFormat() {
	assert := ts.Assert()

	n := NewNodeMatch(NewLeafInt32(1), NewLeafInt32(2))
	s := n.Format("")
	assert.Equal(
		"Match\n"+
			"    [Leaf: int32] \"1\"\n"+
			"    [Leaf: int32] \"2\"",
		s)
}

func (ts *NodeTestSuite) TestNotMatchFormat() {
	assert := ts.Assert()

	n := NewNodeNotMatch(NewLeafInt32(1), NewLeafInt32(2))
	s := n.Format("")
	assert.Equal(
		"Not Match\n"+
			"    [Leaf: int32] \"1\"\n"+
			"    [Leaf: int32] \"2\"",
		s)
}

func (ts *NodeTestSuite) TestComplexMatchFormat() {
	assert := ts.Assert()

	n := NewNodeMatch(
		NewNodeFetch(NewLeafString("%test%")),
		NewNodeMatch(
			NewNodeFetch(NewLeafString("%testing%")),
			NewLeafInt32(1)))
	s := n.Format("")
	assert.Equal(
		"Match\n"+
			"    Fetch\n"+
			"        [Leaf: string] \"%test%\"\n"+
			"    Match\n"+
			"        Fetch\n"+
			"            [Leaf: string] \"%testing%\"\n"+
			"        [Leaf: int32] \"1\"",
		s)
}

func (ts *NodeTestSuite) TestAllFormat() {
	assert := ts.Assert()

	n := NewNodeAll(
		NewLeafInt32(1),
		NewLeafBool(true),
		NewLeafInt64(100))
	s := n.Format("")
	assert.Equal(
		"All Match\n"+
			"    [Leaf: int32] \"1\"\n"+
			"    [Leaf: bool] \"true\"\n"+
			"    [Leaf: int64] \"100\"",
		s)
}

func (ts *NodeTestSuite) TestAnyFormat() {
	assert := ts.Assert()

	n := NewNodeAny(
		NewLeafInt32(1),
		NewLeafBool(true),
		NewLeafInt64(100))
	s := n.Format("")
	assert.Equal(
		"At Least One Matches\n"+
			"    [Leaf: int32] \"1\"\n"+
			"    [Leaf: bool] \"true\"\n"+
			"    [Leaf: int64] \"100\"",
		s)
}

func (ts *NodeTestSuite) TestInvalidFormat() {
	assert := ts.Assert()

	n := &Node{
		Op: OpInvalid,
		Args: []Term{
			NewLeafString("test"),
		},
	}

	s := n.Format("")
	assert.Equal(
		"Invalid\n"+
			"    [Leaf: string] \"test\"",
		s)
}

func (ts *NodeTestSuite) TestMatchEvaluate() {
	assert := ts.Assert()
	require := ts.Require()

	ec := &EvalContext{Replacements: []string{}}

	n := NewNodeMatch(NewLeafInt32(1), NewLeafInt32(2))

	l, err := n.Evaluate(ec)
	require.NoError(err)
	v, err := l.AsBool()
	require.NoError(err)
	assert.False(v)

	n = NewNodeMatch(NewLeafInt32(1), NewLeafInt64(1))

	l, err = n.Evaluate(ec)
	require.NoError(err)
	v, err = l.AsBool()
	require.NoError(err)
	assert.True(v)
}

func (ts *NodeTestSuite) TestNotMatchEvaluate() {
	assert := ts.Assert()
	require := ts.Require()

	ec := &EvalContext{Replacements: []string{}}

	n := NewNodeNotMatch(NewLeafInt32(1), NewLeafInt32(2))

	l, err := n.Evaluate(ec)
	require.NoError(err)
	v, err := l.AsBool()
	require.NoError(err)
	assert.True(v)

	n = NewNodeNotMatch(NewLeafInt32(1), NewLeafInt32(1))

	l, err = n.Evaluate(ec)
	require.NoError(err)
	v, err = l.AsBool()
	require.NoError(err)
	assert.False(v)
}

func (ts *NodeTestSuite) TestAllEvaluate() {
	assert := ts.Assert()
	require := ts.Require()

	ec := &EvalContext{Replacements: []string{}}

	n := NewNodeAll(
		NewLeafInt64(1),
		NewLeafBool(true),
		NewLeafInt32(100))

	l, err := n.Evaluate(ec)
	require.NoError(err)
	v, err := l.AsBool()
	require.NoError(err)
	assert.True(v)

	n = NewNodeAll(
		NewLeafInt64(1),
		NewLeafBool(true),
		NewLeafInt32(0))

	l, err = n.Evaluate(ec)
	require.NoError(err)
	v, err = l.AsBool()
	require.NoError(err)
	assert.False(v)

	n = NewNodeAll(
		NewLeafInt64(1),
		NewLeafBool(true),
		NewLeafString("test"))

	l, err = n.Evaluate(ec)
	require.Error(err)
	assert.Equal(errors.ErrInvalidType(ValueString), err)
}

func (ts *NodeTestSuite) TestAnyEvaluate() {
	assert := ts.Assert()
	require := ts.Require()

	ec := &EvalContext{Replacements: []string{}}

	n := NewNodeAny(
		NewLeafInt64(0),
		NewLeafBool(false),
		NewLeafInt32(100))

	l, err := n.Evaluate(ec)
	require.NoError(err)
	v, err := l.AsBool()
	require.NoError(err)
	assert.True(v)

	n = NewNodeAny(
		NewLeafInt64(0),
		NewLeafBool(false),
		NewLeafInt32(0))

	l, err = n.Evaluate(ec)
	require.NoError(err)
	v, err = l.AsBool()
	require.NoError(err)
	assert.False(v)

	n = NewNodeAny(
		NewLeafInt64(0),
		NewLeafBool(false),
		NewLeafString("test"))

	l, err = n.Evaluate(ec)
	require.Error(err)
	assert.Equal(errors.ErrInvalidType(ValueString), err)
}

func (ts *NodeTestSuite) TestInvalidEvaluate() {
	assert := ts.Assert()
	require := ts.Require()

	n := &Node{
		Op: OpInvalid,
		Args: []Term{
			NewLeafString("test"),
		},
	}

	ec := &EvalContext{Replacements: []string{}}

	_, err := n.Evaluate(ec)
	require.Error(err)
	assert.Equal(errors.ErrInvalidRuleOp(OpInvalid), err)
}

func TestNodeTestSuite(t *testing.T) {
	suite.Run(t, new(NodeTestSuite))
}
