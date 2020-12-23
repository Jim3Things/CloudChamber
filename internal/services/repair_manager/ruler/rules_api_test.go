package ruler

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type RulesApiTestSuite struct {
	suite.Suite
}

func (ts *RulesApiTestSuite) testLeaf(l Term, vt ValueType) {
	require := ts.Require()

	s, ok := l.(*Leaf)
	require.True(ok)
	require.Equal(vt, s.vtype)
}

func (ts *RulesApiTestSuite) TestV() {
	ts.testLeaf(V("test"), ValueString)
	ts.testLeaf(V(3), ValueInt32)
	ts.testLeaf(V(true), ValueBool)
	ts.testLeaf(V("foo/bar/%baz%"), ValueString)
	ts.testLeaf(V(int32(1)), ValueInt32)
	ts.testLeaf(V(int64(1)), ValueInt64)
}

func TestRulesApiTest(t *testing.T) {
	suite.Run(t, new(RulesApiTestSuite))
}
