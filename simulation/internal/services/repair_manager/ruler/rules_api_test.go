package ruler

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Jim3Things/CloudChamber/simulation/internal/common"
	"github.com/Jim3Things/CloudChamber/simulation/internal/tracing"
	"github.com/Jim3Things/CloudChamber/simulation/internal/tracing/exporters"
	"github.com/Jim3Things/CloudChamber/simulation/pkg/errors"
)

type MockTable struct {
	rows map[string]map[string]interface{}
}

func (m *MockTable) GetValue(key *Key) (interface{}, error) {
	if len(key.Nodes) != 1 {
		return nil, errors.ErrInvalidArgLen{
			Op:       fmt.Sprintf("GetValue from table %s", key.Table),
			Required: "exactly 1 node in the path",
			Actual:   len(key.Nodes),
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

	utf *exporters.Exporter
}

func (ts *RulesApiTestSuite) SetupSuite() {
	ts.utf = exporters.NewExporter(exporters.NewUTForwarder())
	exporters.ConnectToProvider(ts.utf)
}

func (ts *RulesApiTestSuite) SetupTest() {
	_ = ts.utf.Open(ts.T())
}

func (ts *RulesApiTestSuite) TearDownTest() {
	ts.utf.Close()
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
		kv := map[string]interface{}{
			"f1": i,
			"s2": fmt.Sprintf("test%d", i),
			"b3": i/2*2 == i,
		}
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

	vars := map[string]string{
		"%table%": "table1",
	}

	args := map[string]Term{
		"target": N("%table%/row1.s2"),
	}

	r := []Rule{
		{
			Name:   "test rule",
			Where:  Match(N("%table%/row1.f1"), V(1)),
			Reason: "f1 matched expectations",
			Choices: []RuleChoice{
				{
					Assuming: Match(N("%table%/row1.b3"), V(true)),
					Chosen:   "should not be chosen",
					Rejected: "b3 is not true",
					With:     nil,
					Call: func(ctx context.Context, args map[string]Term, ec *EvalContext) (*Proposal, error) {
						require.Fail("should not get here")
						return &Proposal{}, nil
					},
				},
				{
					Assuming: V(true),
					Chosen:   "this option is always chosen",
					Rejected: "never fails",
					With:     args,
					Call: func(ctx context.Context, args map[string]Term, ec *EvalContext) (*Proposal, error) {
						val, err := args["target"].Evaluate(ec)
						require.NoError(err)
						s, err := val.AsString()
						require.NoError(err)

						tracing.Info(ctx, "Processing with %q: %q", "target", s)
						return &Proposal{
							Path:  "target",
							Value: s,
						}, nil
					},
				},
				{
					Assuming: V(true),
					Chosen:   "should not be chosen",
					Rejected: "should not get here",
					With:     nil,
					Call: func(ctx context.Context, args map[string]Term, ec *EvalContext) (*Proposal, error) {
						require.Fail("should not get here")
						return &Proposal{}, nil
					},
				},
			},
		},
	}

	ctx := common.ContextWithTick(context.Background(), -1)
	props, err := Process(ctx, r, tables, vars)
	require.NoError(err)
	require.NotNil(props)
	require.Equal(1, len(props))
	require.Equal("target", props[0].Path)
	require.Equal("test1", props[0].Value)
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
