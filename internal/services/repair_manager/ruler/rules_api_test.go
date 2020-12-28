package ruler

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Jim3Things/CloudChamber/pkg/errors"
)

type MockTable struct {
	rows map[string]map[string]interface{}
}

func (m *MockTable) GetValue(key *Key) (interface{}, error) {
	if len(key.Nodes) != 1 {
		return nil, ErrInvalidArgLen{
			op:       fmt.Sprintf("GetValue from table %s", key.Table),
			required: "exactly 1 node in the path",
			actual:   len(key.Nodes),
		}
	}

	row, ok := m.rows[key.Nodes[0]]
	if !ok {
		return nil, errors.ErrStoreKeyNotFound(key.Nodes[0])
	}

	v, ok := row[key.Field]
	if !ok {
		return nil, errors.ErrStoreKeyNotFound(key.Field)
	}

	return v, nil
}

type MockTables struct {
	tables map[string]*MockTable
}

func (mt *MockTables) GetTable(key *Key) (Table, error) {
	r, ok := mt.tables[key.Table]
	if !ok {
		return nil, errors.ErrStoreKeyNotFound(key.Table)
	}

	return r, nil
}

type RulesApiTestSuite struct {
	suite.Suite
}

func (ts *RulesApiTestSuite) testLeaf(l Term, vt ValueType) {
	require := ts.Require()

	s, ok := l.(*Leaf)
	require.True(ok)
	require.Equal(vt, s.vtype)
}

func (ts *RulesApiTestSuite) buildMockTable(rowCount int) *MockTable {
	rows := make(map[string]map[string]interface{})

	for i := 0; i < rowCount; i++ {
		kv := make(map[string]interface{})
		kv["f1"] = i
		kv["s2"] = fmt.Sprintf("test%d", i)
		kv["b3"] = i/2 * 2 == i
		rows[fmt.Sprintf("row%d", i)] = kv
	}

	return &MockTable{rows: rows}
}

func (ts *RulesApiTestSuite) buildMockTables(tableCount int, rowCount int) *MockTables {
	t := make(map[string]*MockTable)
	for i := 0; i < tableCount; i++ {
		t[fmt.Sprintf("table%d", i)] = ts.buildMockTable(rowCount)
	}

	return &MockTables{tables: t}
}

func (ts *RulesApiTestSuite) TestSimple() {
	require := ts.Require()

	tables := ts.buildMockTables(2, 2)

	vars := make(map[string]string)
	vars["%table%"] = "table1"

	r := []Rule{
		{
			Where:  Match(N("%table%/row1.f1"), V(1)),
			Reason: "f1 matched expectations",
			Choices: []RuleChoice{
				{
					Assuming: V(true),
					Chosen:   "always chosen",
					Rejected: "never fails",
					With:     nil,
					Call: func(args []Arg) (*Proposal, error) {
						return &Proposal{}, nil
					},
				},
				{
					Assuming: V(true),
					Chosen:   "should not be chosen",
					Rejected: "should not get here",
					With:     nil,
					Call: func(args []Arg) (*Proposal, error) {
						require.Fail("should not get here")
						return &Proposal{}, nil
					},
				},
			},
		},
	}

	props, err := Process(r, tables, vars)
	require.NoError(err)
	require.NotNil(props)
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
