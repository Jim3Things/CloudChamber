package inventory

import (
	"context"
	"flag"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Jim3Things/CloudChamber/internal/clients/store"
	"github.com/Jim3Things/CloudChamber/internal/config"
	"github.com/Jim3Things/CloudChamber/internal/tracing/exporters"
	pb "github.com/Jim3Things/CloudChamber/pkg/protos/inventory"
)

type testSuiteCore struct {

	suite.Suite

	baseURI string

	utf *exporters.Exporter

	cfg *config.GlobalConfig

	store *store.Store
}

func (ts *testSuiteCore) rootName(suffix string)   string { return "StandardRoot-" + suffix }
func (ts *testSuiteCore) regionName(suffix string) string { return "REG-PNW-"      + suffix }
func (ts *testSuiteCore) zoneName(suffix string)   string { return "Zone-01-"      + suffix }
func (ts *testSuiteCore) rackName(suffix string)   string { return "Rack-01-"      + suffix }
func (ts *testSuiteCore) pduID(ID int64)      int64  { return int64(ID)}
func (ts *testSuiteCore) torID(ID int64)      int64  { return int64(ID)}
func (ts *testSuiteCore) bladeID(ID int64)    int64  { return int64(ID)}

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

	stdSuffix := "TestNewRoot"

	ctx := context.Background()

	root, err := NewRoot (ctx, ts.store, DefinitionTable)
	require.NoError(err)

	root.SetDetails(ctx, &pb.RootDetails{
		Name: ts.rootName(stdSuffix),
		Notes: "root for inventory definition test",
	})
}

func (ts *testSuiteCore) TestNewRegion() {
	assert := ts.Assert()
	require := ts.Require()

	stdSuffix := "TestNewRegion"

	ctx := context.Background()

	r, err := NewRegion(ctx, ts.store, DefinitionTable, ts.regionName(stdSuffix))
	require.NoError(err)

	r.SetDetails(ctx, &pb.RegionDetails{
		Name:     ts.regionName(stdSuffix),
		State:    pb.State_in_service,
		Location: "Pacific NW",
		Notes:    "region for inventory definition test",
	})

	rev, err := r.Create(ctx)

	require.NoError(err)
	assert.NotEqual(store.RevisionInvalid, rev)
}

func (ts *testSuiteCore) TestNewChildRegion() {
	assert := ts.Assert()
	require := ts.Require()

	stdSuffix := "TestNewChildRegion"

	regionName := ts.regionName(stdSuffix)

	ctx := context.Background()

	root, err := NewRoot (ctx, ts.store, DefinitionTable)
	require.NoError(err)

	root.SetDetails(ctx, &pb.RootDetails{
		Name:  ts.rootName(stdSuffix),
		Notes: "root for inventory definition test",
	})

	region, err := root.NewChild(ctx, regionName)
	require.NoError(err)

	region.SetDetails(ctx, &pb.RegionDetails{
		Name:     regionName,
		State:    pb.State_in_service,
		Location: "Pacific NW",
		Notes:    "region for inventory definition test (" + stdSuffix + ")",
	})

	rev, err := region.Create(ctx)
	require.NoError(err)
	assert.NotEqual(store.RevisionInvalid, rev)
}

func (ts *testSuiteCore) TestNewZone() {
	assert := ts.Assert()
	require := ts.Require()

	stdSuffix := "TestNewZone"

	ctx := context.Background()

	zone, err := NewZone(ctx, ts.store, DefinitionTable, ts.regionName(stdSuffix), ts.zoneName(stdSuffix))
	require.NoError(err)

	zone.SetDetails(ctx, &pb.ZoneDetails{
		Enabled:  true,
		State:    pb.State_in_service,
		Location: "Pacific NW",
		Notes:    "zone for inventory definition test",
	})

	rev, err := zone.Create(ctx)

	require.NoError(err)
	assert.NotEqual(store.RevisionInvalid, rev)
}

func (ts *testSuiteCore) TestNewChildZone() {
	assert := ts.Assert()
	require := ts.Require()

	stdSuffix := "TestNewChildZone"

	regionName := ts.regionName(stdSuffix)
	zoneName   := ts.zoneName(stdSuffix)

	ctx := context.Background()

	root, err := NewRoot (ctx, ts.store, DefinitionTable)
	require.NoError(err)

	region, err := root.NewChild(ctx, regionName)
	require.NoError(err)

	zone, err := region.NewChild(ctx, zoneName)
	require.NoError(err)

	zone.SetDetails(ctx, &pb.ZoneDetails{
		Enabled:  true,
		State:    pb.State_in_service,
		Location: "Pacific NW",
		Notes:    "zone for inventory definition test (" + stdSuffix + ")",
	})

	rev, err := zone.Create(ctx)
	require.NoError(err)
	assert.NotEqual(store.RevisionInvalid, rev)
}

func (ts *testSuiteCore) TestNewRack() {
	assert := ts.Assert()
	require := ts.Require()

	stdSuffix := "TestNewRack"

	ctx := context.Background()

	rack, err := NewRack(
		ctx,
		ts.store,
		DefinitionTable,
		ts.regionName(stdSuffix),
		ts.zoneName(stdSuffix),
		ts.rackName(stdSuffix),
	)
	require.NoError(err)

	rack.SetDetails(ctx, &pb.RackDetails{
		Enabled:   true,
		Condition: pb.Condition_operational,
		Location:  "Pacific NW",
		Notes:     "rack for inventory definition test",
	})

	rev, err := rack.Create(ctx)

	require.NoError(err)
	assert.NotEqual(store.RevisionInvalid, rev)
}

func (ts *testSuiteCore) TestNewChildRack() {
	assert := ts.Assert()
	require := ts.Require()

	stdSuffix := "TestNewChildRack"

	regionName := ts.regionName(stdSuffix)
	zoneName   := ts.zoneName(stdSuffix)
	rackName   := ts.rackName(stdSuffix)

	ctx := context.Background()

	root, err := NewRoot (ctx, ts.store, DefinitionTable)
	require.NoError(err)

	region, err := root.NewChild(ctx, regionName)
	require.NoError(err)

	zone, err := region.NewChild(ctx, zoneName)
	require.NoError(err)

	rack, err := zone.NewChild(ctx, rackName)
	require.NoError(err)

	rack.SetDetails(ctx, &pb.RackDetails{
		Enabled:   true,
		Condition: pb.Condition_operational,
		Location:  "Pacific NW",
		Notes:     "rack for inventory definition test (" + stdSuffix + ")",
	})

	rev, err := rack.Create(ctx)
	require.NoError(err)
	assert.NotEqual(store.RevisionInvalid, rev)
}

func (ts *testSuiteCore) TestNewPdu() {
	assert := ts.Assert()
	require := ts.Require()

	stdSuffix := "TestNewPdu"
	ports     := make(map[int64]*pb.PowerPort)

	ctx := context.Background()

	pdu, err := NewPdu(
		ctx,
		ts.store,
		DefinitionTable,
		ts.regionName(stdSuffix),
		ts.zoneName(stdSuffix),
		ts.rackName(stdSuffix),
		ts.pduID(1),
	)
	require.NoError(err)

	pdu.SetDetails(ctx, &pb.PduDetails{
		Enabled:   true,
		Condition: pb.Condition_operational,
	})

	pdu.SetPorts(ctx, &ports)

	rev, err := pdu.Create(ctx)

	require.NoError(err)
	assert.NotEqual(store.RevisionInvalid, rev)
}

func (ts *testSuiteCore) TestNewChildPdu() {
	assert := ts.Assert()
	require := ts.Require()

	stdSuffix := "TestNewChildRack"

	regionName := ts.regionName(stdSuffix)
	zoneName   := ts.zoneName(stdSuffix)
	rackName   := ts.rackName(stdSuffix)
	pduID      := ts.pduID(1)
	ports      := make(map[int64]*pb.PowerPort)

	ctx := context.Background()

	root, err := NewRoot (ctx, ts.store, DefinitionTable)
	require.NoError(err)

	region, err := root.NewChild(ctx, regionName)
	require.NoError(err)

	zone, err := region.NewChild(ctx, zoneName)
	require.NoError(err)

	rack, err := zone.NewChild(ctx, rackName)
	require.NoError(err)

	pdu, err := rack.NewPdu(ctx, pduID)
	require.NoError(err)

	pdu.SetDetails(ctx, &pb.PduDetails{
		Enabled:   true,
		Condition: pb.Condition_operational,
	})

	pdu.SetPorts(ctx, &ports)

	rev, err := pdu.Create(ctx)
	require.NoError(err)
	assert.NotEqual(store.RevisionInvalid, rev)
}

func (ts *testSuiteCore) TestNewBlade() {
	assert := ts.Assert()
	require := ts.Require()

	stdSuffix := "TestNewBlade"

	details := &pb.BladeDetails{
		Enabled:   true,
		Condition: pb.Condition_operational,
	}

	capacity := &pb.BladeCapacity{
		Cores:                  16,
		MemoryInMb:             1024,
		DiskInGb:               32,
		NetworkBandwidthInMbps: 1024,
		Arch:                   "amd64",
	}

	bootInfo := &pb.BladeBootInfo{
		Source:     pb.BladeBootInfo_local,
		Image:      "test-image.vhdx",
		Version:    "20201225-0000",
		Parameters: "-param1=val1 -param2=val2",
	}

	ctx := context.Background()

	blade, err := NewBlade(
		ctx,
		ts.store,
		DefinitionTable,
		ts.regionName(stdSuffix),
		ts.zoneName(stdSuffix),
		ts.rackName(stdSuffix),
		ts.bladeID(1),
	)
	require.NoError(err)

	blade.SetDetails(ctx,details)
	blade.SetCapacity(ctx, capacity)
	blade.SetBootInfo(ctx, true, bootInfo)

	rev, err := blade.Create(ctx)

	require.NoError(err)
	assert.NotEqual(store.RevisionInvalid, rev)
}

func (ts *testSuiteCore) TestNewChildBlade() {
	assert := ts.Assert()
	require := ts.Require()

	stdSuffix := "TestNewChildBlade"

	regionName := ts.regionName(stdSuffix)
	zoneName   := ts.zoneName(stdSuffix)
	rackName   := ts.rackName(stdSuffix)
	bladeID    := ts.bladeID(1)

	details := &pb.BladeDetails{
		Enabled:   true,
		Condition: pb.Condition_operational,
	}

	capacity := &pb.BladeCapacity{
		Cores: 16,
		MemoryInMb: 1024,
		DiskInGb: 32,
		NetworkBandwidthInMbps: 1024,
		Arch: "amd64",
	}

	bootInfo := &pb.BladeBootInfo{
		Source:     pb.BladeBootInfo_local,
		Image:      "test-image.vhdx",
		Version:    "20201225-0000",
		Parameters: "-param1=val1 -param2=val2",
	}


	ctx := context.Background()

	root, err := NewRoot (ctx, ts.store, DefinitionTable)
	require.NoError(err)

	region, err := root.NewChild(ctx, regionName)
	require.NoError(err)

	zone, err := region.NewChild(ctx, zoneName)
	require.NoError(err)

	rack, err := zone.NewChild(ctx, rackName)
	require.NoError(err)

	blade, err := rack.NewBlade(ctx, bladeID)
	require.NoError(err)

	blade.SetDetails(ctx, details)
	blade.SetCapacity(ctx, capacity)
	blade.SetBootInfo(ctx, true, bootInfo)

	rev, err := blade.Create(ctx)
	require.NoError(err)
	assert.NotEqual(store.RevisionInvalid, rev)
}

func (ts *testSuiteCore) TestRegionReadDetails() {
	assert := ts.Assert()
	require := ts.Require()

	stdSuffix := "TestRegionReadDetails"

	stdDetails := &pb.RegionDetails{
		Name:     ts.regionName(stdSuffix),
		State:    pb.State_in_service,
		Location: "Pacific NW",
		Notes:    "region for inventory definition test",
	}

	ctx := context.Background()

	r, err := NewRegion(ctx, ts.store, DefinitionTable, ts.regionName(stdSuffix))
	require.NoError(err)

	r.SetDetails(ctx, stdDetails)

	rev, err := r.Create(ctx)
	require.NoError(err)
	assert.NotEqual(store.RevisionInvalid, rev)

	rev2 := r.GetRevision(ctx)
	assert.Equal(rev, rev2)


	rRead, err := NewRegion(ctx, ts.store, DefinitionTable, ts.regionName(stdSuffix))
	require.NoError(err)

	revRead, err := rRead.Read(ctx)
	require.NoError(err)
	assert.Equal(rev, revRead)
	assert.Equal(revRead, rRead.GetRevision(ctx))

	revDet, detRead := rRead.GetDetails(ctx)
	require.NoError(err)
	assert.Equal(rev, revDet)
	assert.Equal(stdDetails, detRead)
}



func TestInventoryTestSuite(t *testing.T) {
	suite.Run(t, new(testSuiteCore))
}
