package inventory

import (
	"context"
	"flag"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Jim3Things/CloudChamber/internal/clients/store"
	"github.com/Jim3Things/CloudChamber/internal/config"
	"github.com/Jim3Things/CloudChamber/internal/tracing/exporters"
	"github.com/Jim3Things/CloudChamber/pkg/protos/inventory"
	pb "github.com/Jim3Things/CloudChamber/pkg/protos/inventory"
)

type testSuiteCore struct {

	suite.Suite

	baseURI string

	utf *exporters.Exporter

	cfg *config.GlobalConfig

	store *store.Store
}

func (ts *testSuiteCore) rootName()   string { return "StandardRoot"}
func (ts *testSuiteCore) regionName() string { return "REG-PNW" }
func (ts *testSuiteCore) zoneName()   string { return "Zone-001" }
func (ts *testSuiteCore) rackName()   string { return "Rack-00-15" }
func (ts *testSuiteCore) pduID()      int64  { return int64(1)}
func (ts *testSuiteCore) torID()      int64  { return int64(1)}
func (ts *testSuiteCore) bladeID()    int64  { return int64(1)}

func (ts *testSuiteCore) SetupSuite() {
	require := ts.Require()

	configPath := flag.String("config", "./testdata", "path to the configuration file")
	flag.Parse()

	cfg, err := config.ReadGlobalConfig(*configPath)
	require.NoError(err, "failed to process the global configuration")

	ts.utf = exporters.NewExporter(exporters.NewUTForwarder())
	exporters.ConnectToProvider(ts.utf)

	store.Initialize(cfg)

	ts.cfg   = cfg
	ts.store = store.NewStore()
	}	

func (ts *testSuiteCore) SetupTest() {
	require := ts.Require()

	_ = ts.utf.Open(ts.T())

	err := ts.store.Connect()
	require.NoError(err)
}

func (ts *testSuiteCore) TearDownTest() {
	ts.store.Disconnect()
	ts.utf.Close()
}

func (ts *testSuiteCore) TestNewRoot() {
	require := ts.Require()

	ctx := context.Background()

	root, err := NewRoot (ctx, ts.store, DefinitionTable)
	require.NoError(err)

	err = root.SetDetails(ctx, &pb.RootDetails{
		Name: ts.rootName(),
		Notes: "root for inventory definition test",
	})
	require.NoError(err)
}

func (ts *testSuiteCore) TestNewRegion() {
	assert := ts.Assert()
	require := ts.Require()

	ctx := context.Background()

	r, err := NewRegion (ctx, ts.store, DefinitionTable, ts.regionName())
	require.NoError(err)

	err = r.SetDetails(ctx, &inventory.RegionDetails{
		Name:     ts.regionName(),
		State:    pb.State_in_service,
		Location: "Pacific NW",
		Notes:    "region for inventory definition test",
	})

	require.NoError(err)

	rev, err := r.Create(ctx)

	require.NoError(err)
	assert.NotEqual(store.RevisionInvalid, rev)
}

func (ts *testSuiteCore) TestNewRegionRelative() {
	assert := ts.Assert()
	require := ts.Require()

	regionName := ts.regionName() + "Relative"

	ctx := context.Background()

	root, err := NewRoot (ctx, ts.store, DefinitionTable)
	require.NoError(err)

	err = root.SetDetails(ctx, &pb.RootDetails{
		Name:  ts.rootName(),
		Notes: "root for inventory definition test",
	})

	require.NoError(err)

	region, err := root.NewChild(ctx, regionName)
	require.NoError(err)

	err = region.SetDetails(ctx, &pb.RegionDetails{
		Name:     regionName,
		State:    pb.State_in_service,
		Location: "Pacific NW",
		Notes:    "region for inventory definition test (relative)",
	})
	require.NoError(err)

	rev, err := region.Create(ctx)
	require.NoError(err)
	assert.NotEqual(store.RevisionInvalid, rev)
}



func TestInventoryTestSuite(t *testing.T) {
	suite.Run(t, new(testSuiteCore))
}
