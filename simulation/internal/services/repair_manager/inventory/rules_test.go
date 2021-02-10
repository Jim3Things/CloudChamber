package inventory

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Jim3Things/CloudChamber/simulation/internal/common"
	"github.com/Jim3Things/CloudChamber/simulation/internal/services/repair_manager/inventory/rules"
	r "github.com/Jim3Things/CloudChamber/simulation/internal/services/repair_manager/ruler"
	"github.com/Jim3Things/CloudChamber/simulation/internal/tracing/exporters"
	"github.com/Jim3Things/CloudChamber/simulation/pkg/errors"
)

type cable map[string]interface{}

type mockPdu struct {
	cables map[string]cable
}

type mockTor struct {
	cables map[string]cable
}

type mockBlade struct {
	fields map[string]interface{}
}

type mockRack struct {
	pdu    mockPdu
	tor    mockTor
	blades map[string]*mockBlade
}

type mockTable struct {
	racks map[string]*mockRack
}

func (t *mockTable) GetValue(key *r.Key) (interface{}, error) {
	if len(key.Nodes) != 4 {
		return nil, errors.ErrMissingPath("path is too short")
	}
	if key.Nodes[0] != "racks" {
		return nil, errors.ErrMissingPath(key.Nodes[0])
	}

	rack, ok := t.racks[key.Nodes[1]]
	if !ok {
		return nil, errors.ErrMissingPath(key.Nodes[1])
	}

	switch key.Nodes[2] {
	case "blades":
		blade, ok := rack.blades[key.Nodes[3]]
		if !ok {
			return nil, errors.ErrMissingPath(key.Nodes[3])
		}

		f, ok := blade.fields[key.Field]
		if !ok {
			return nil, errors.ErrMissingFieldName(key.Field)
		}

		return f, nil

	case "pdu":
		cable, ok := rack.pdu.cables[key.Nodes[3]]
		if !ok {
			return nil, errors.ErrMissingPath(key.Nodes[3])
		}

		f, ok := cable[key.Field]
		if !ok {
			return nil, errors.ErrMissingFieldName(key.Field)
		}

		return f, nil

	case "tor":
		cable, ok := rack.tor.cables[key.Nodes[3]]
		if !ok {
			return nil, errors.ErrMissingPath(key.Nodes[3])
		}

		f, ok := cable[key.Field]
		if !ok {
			return nil, errors.ErrMissingFieldName(key.Field)
		}

		return f, nil

	default:
		return nil, errors.ErrMissingPath(key.Nodes[2])
	}

}

type mockTables struct {
	target   *mockTable
	observed *mockTable
}

func (t *mockTables) GetTable(key *r.Key) (r.Table, error) {
	switch key.Table {
	case "target":
		return t.target, nil

	case "observed":
		return t.observed, nil

	default:
		return nil, errors.ErrStoreKeyNotFound(key.Table)
	}
}

func createTables() *mockTables {
	return &mockTables{
		target: &mockTable{
			racks: map[string]*mockRack{
				"rack1": {
					pdu: mockPdu{},
					tor: mockTor{},
					blades: map[string]*mockBlade{
						"1": {
							fields: map[string]interface{}{
								"power":     true,
								"connected": false,
								"powering":  false,
							},
						},
						"2": {
							fields: map[string]interface{}{
								"power":     true,
								"connected": false,
								"powering":  false,
							},
						},
						"3": {
							fields: map[string]interface{}{
								"power":     false,
								"connected": false,
								"powering":  false,
							},
						},
					},
				},
				"rack2": {
					pdu: mockPdu{},
					tor: mockTor{},
					blades: map[string]*mockBlade{
						"1": {
							fields: map[string]interface{}{
								"power":     false,
								"connected": true,
								"powering":  false,
							},
						},
						"2": {
							fields: map[string]interface{}{
								"power":     false,
								"connected": true,
								"powering":  false,
							},
						},
						"3": {
							fields: map[string]interface{}{
								"power":     true,
								"connected": true,
								"powering":  false,
							},
						},
					},
				},
			}},
		observed: &mockTable{
			racks: map[string]*mockRack{
				"rack1": {
					pdu: mockPdu{
						cables: map[string]cable{
							"blade_1": {
								"faulted": false,
							},
							"blade_2": {
								"faulted": true,
							},
							"blade_3": {
								"faulted": false,
							},
						},
					},
					tor: mockTor{
						cables: map[string]cable{
							"blade_1": {
								"faulted": false,
							},
							"blade_2": {
								"faulted": false,
							},
							"blade_3": {
								"faulted": true,
							},
						},
					},
					blades: map[string]*mockBlade{
						"1": {
							fields: map[string]interface{}{
								"power":     false,
								"connected": false,
								"powering":  false,
							},
						},
						"2": {
							fields: map[string]interface{}{
								"power":     false,
								"connected": false,
								"powering":  false,
							},
						},
						"3": {
							fields: map[string]interface{}{
								"power":     false,
								"connected": false,
								"powering":  false,
							},
						},
					},
				},
				"rack2": {
					pdu: mockPdu{
						cables: map[string]cable{
							"blade_1": {
								"faulted": false,
							},
							"blade_2": {
								"faulted": false,
							},
							"blade_3": {
								"faulted": false,
							},
						},
					},
					tor: mockTor{
						cables: map[string]cable{
							"blade_1": {
								"faulted": false,
							},
							"blade_2": {
								"faulted": true,
							},
							"blade_3": {
								"faulted": false,
							},
						},
					},
					blades: map[string]*mockBlade{
						"1": {
							fields: map[string]interface{}{
								"power":     false,
								"connected": false,
								"powering":  false,
							},
						},
						"2": {
							fields: map[string]interface{}{
								"power":     false,
								"connected": false,
								"powering":  false,
							},
						},
						"3": {
							fields: map[string]interface{}{
								"power":     false,
								"connected": false,
								"powering":  false,
							},
						},
					},
				},
			},
		},
	}
}

type RulesTestSuite struct {
	suite.Suite

	utf *exporters.Exporter
}

func (ts *RulesTestSuite) SetupSuite() {
	ts.utf = exporters.NewExporter(exporters.NewUTForwarder())
	exporters.ConnectToProvider(ts.utf)
}

func (ts *RulesTestSuite) SetupTest() {
	_ = ts.utf.Open(ts.T())
}

func (ts *RulesTestSuite) TearDownTest() {
	ts.utf.Close()
}

func (ts *RulesTestSuite) TestPowerOn() {
	require := ts.Require()

	tables := createTables()

	vars := map[string]string{
		"%target%":   "target",
		"%observed%": "observed",
		"%rack%":     "rack1",
		"%blade%":    "1",
	}

	ctx := common.ContextWithTick(context.Background(), -1)
	props, err := r.Process(ctx, rules.Rules, tables, vars)
	require.NoError(err)

	require.NotNil(props)
	require.Equal(1, len(props))
	require.Equal("racks/rack1/blades/1.powering", props[0].Path)
	require.Equal(true, props[0].Value)
}

func (ts *RulesTestSuite) TestDuplicatePowerOn() {
	require := ts.Require()

	tables := createTables()

	tables.observed.racks["rack1"].blades["1"].fields["powering"] = true

	vars := map[string]string{
		"%target%":   "target",
		"%observed%": "observed",
		"%rack%":     "rack1",
		"%blade%":    "1",
	}

	ctx := common.ContextWithTick(context.Background(), -1)
	props, err := r.Process(ctx, rules.Rules, tables, vars)

	// No error, and no proposal
	require.NoError(err)
	require.Nil(props)
}

func (ts *RulesTestSuite) TestPowerOnHI() {
	require := ts.Require()

	tables := createTables()

	vars := map[string]string{
		"%target%":   "target",
		"%observed%": "observed",
		"%rack%":     "rack1",
		"%blade%":    "2",
	}

	ctx := common.ContextWithTick(context.Background(), -1)
	props, err := r.Process(ctx, rules.Rules, tables, vars)

	// No error, and no proposal
	require.NoError(err)

	require.NotNil(props)
	require.Equal(1, len(props))
	require.Equal("message", props[0].Path)
	require.Equal("fix the power connection for blade 2 in rack \"rack1\"", props[0].Value)
}

func (ts *RulesTestSuite) TestConnect() {
	require := ts.Require()

	tables := createTables()

	vars := map[string]string{
		"%target%":   "target",
		"%observed%": "observed",
		"%rack%":     "rack2",
		"%blade%":    "1",
	}

	ctx := common.ContextWithTick(context.Background(), -1)
	props, err := r.Process(ctx, rules.Rules, tables, vars)
	require.NoError(err)

	require.NotNil(props)
	require.Equal(1, len(props))
	require.Equal("racks/rack2/blades/1.connecting", props[0].Path)
	require.Equal(true, props[0].Value)
}

func (ts *RulesTestSuite) TestConnectHI() {
	require := ts.Require()

	tables := createTables()

	vars := map[string]string{
		"%target%":   "target",
		"%observed%": "observed",
		"%rack%":     "rack2",
		"%blade%":    "2",
	}

	ctx := common.ContextWithTick(context.Background(), -1)
	props, err := r.Process(ctx, rules.Rules, tables, vars)

	// No error, and no proposal
	require.NoError(err)

	require.NotNil(props)
	require.Equal(1, len(props))
	require.Equal("message", props[0].Path)
	require.Equal("fix the network cable connection for blade 2 in rack \"rack2\"", props[0].Value)
}

func (ts *RulesTestSuite) TestBothPowerOnAndConnect() {
	require := ts.Require()

	tables := createTables()

	vars := map[string]string{
		"%target%":   "target",
		"%observed%": "observed",
		"%rack%":     "rack2",
		"%blade%":    "3",
	}

	ctx := common.ContextWithTick(context.Background(), -1)
	props, err := r.Process(ctx, rules.Rules, tables, vars)

	// No error, and no proposal
	require.NoError(err)

	require.NotNil(props)
	require.Equal(2, len(props))
	require.Equal("racks/rack2/blades/3.powering", props[0].Path)
	require.Equal(true, props[0].Value)
	require.Equal("racks/rack2/blades/3.connecting", props[1].Path)
	require.Equal(true, props[0].Value)
}

func TestRulesTestSuite(t *testing.T) {
	suite.Run(t, new(RulesTestSuite))
}
