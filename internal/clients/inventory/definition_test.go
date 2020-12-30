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


func (ts *testSuiteCore) stdRootDetails(suffix string) *pb.RootDetails {
	return &pb.RootDetails{
		Name: ts.rootName(suffix),
		Notes: "root for inventory definition test",

	}
}

func (ts *testSuiteCore) stdRegionDetails(suffix string) *pb.RegionDetails {
	return &pb.RegionDetails{
		Name:     ts.regionName(suffix),
		State:    pb.State_in_service,
		Location: "Pacific NW",
		Notes:    "region for inventory definition test",
	}
}

func (ts *testSuiteCore) stdZoneDetails(suffix string) *pb.ZoneDetails {
	return &pb.ZoneDetails{
		Enabled:  true,
		State:    pb.State_in_service,
		Location: "Pacific NW",
		Notes:    "zone for inventory definition test",
	}
}

func (ts *testSuiteCore) stdRackDetails(suffix string) *pb.RackDetails {
	return &pb.RackDetails{
		Enabled:   true,
		Condition: pb.Condition_operational,
		Location:  "Pacific NW",
		Notes:     "rack for inventory definition test",
	}
}

func (ts *testSuiteCore) stdPduDetails(ID int64) *pb.PduDetails {
	return &pb.PduDetails{
		Enabled:   true,
		Condition: pb.Condition_operational,
	}
}

func (ts *testSuiteCore) stdPowerPorts(count int) *map[int64]*pb.PowerPort {
	ports := make(map[int64]*pb.PowerPort)

	for i := 0; i < count; i++ {
		ports[int64(i)] = &pb.PowerPort{
			Wired: true,
			Item: &pb.Hardware{
				Type: pb.Hardware_blade,
				Id: int64(i),
				Port: int64(i),
			},
		}
	}

	return &ports
}

func (ts *testSuiteCore) stdTorDetails() *pb.TorDetails {
	return &pb.TorDetails{
		Enabled:   true,
		Condition: pb.Condition_operational,
	}
}

func (ts *testSuiteCore) stdNetworkPorts(count int) *map[int64]*pb.NetworkPort {
	ports := make(map[int64]*pb.NetworkPort)

	for i := 0; i < count; i++ {
		ports[int64(i)] = &pb.NetworkPort{
			Wired: true,
			Item: &pb.Hardware{
				Type: pb.Hardware_blade,
				Id: int64(i),
				Port: int64(1),
			},
		}
	}

	return &ports
}

func (ts *testSuiteCore) stdBladeDetails() *pb.BladeDetails {
	return &pb.BladeDetails{
		Enabled:   true,
		Condition: pb.Condition_operational,
	}
}

func (ts *testSuiteCore) stdBladeCapacity() *pb.BladeCapacity {
	return &pb.BladeCapacity{
		Cores:                  16,
		MemoryInMb:             1024,
		DiskInGb:               32,
		NetworkBandwidthInMbps: 1024,
		Arch:                   "amd64",
	}
}

func (ts *testSuiteCore) stdBladeBootInfo() *pb.BladeBootInfo {
	return &pb.BladeBootInfo{
		Source:     pb.BladeBootInfo_local,
		Image:      "test-image.vhdx",
		Version:    "20201225-0000",
		Parameters: "-param1=val1 -param2=val2",
	}
}


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
	assert := ts.Assert()
	require := ts.Require()

	stdSuffix := "TestNewRoot"
	stdDetails := ts.stdRootDetails(stdSuffix)

	ctx := context.Background()

	root, err := NewRoot (ctx, ts.store, DefinitionTable)
	require.NoError(err)

	// We only expect real revision values once there has been a create 
	// or update to the store.
	//
	rev := root.GetRevision(ctx)
	assert.Equal(store.RevisionInvalid, rev)

	rev = root.GetRevisionRecord(ctx)
	assert.Equal(store.RevisionInvalid, rev)

	rev = root.GetRevisionStore(ctx)
	assert.Equal(store.RevisionInvalid, rev)

	// Now try the various combinations of setting and clearing details.
	//
	rev, details := root.GetDetails(ctx)
	assert.Equal(store.RevisionInvalid, rev)
	assert.Nil(details)

	root.SetDetails(ctx, stdDetails)

	rev, details = root.GetDetails(ctx)
	assert.Equal(store.RevisionInvalid, rev)
	require.NotNil(details)
	assert.Equal(stdDetails, details)

	root.SetDetails(ctx, nil)

	rev, details = root.GetDetails(ctx)
	assert.Equal(store.RevisionInvalid, rev)
	require.Nil(details)

	rev, err = root.Read(ctx)
	require.Error(err)
	assert.Equal(ErrFunctionNotAvailable, err)
	assert.Equal(store.RevisionInvalid, rev)

	rev, err = root.Update(ctx, false)
	require.Error(err)
	assert.Equal(ErrFunctionNotAvailable, err)
	assert.Equal(store.RevisionInvalid, rev)

	rev, err = root.Update(ctx, true)
	require.Error(err)
	assert.Equal(ErrFunctionNotAvailable, err)
	assert.Equal(store.RevisionInvalid, rev)

	rev, err = root.Delete(ctx, false)
	require.Error(err)
	assert.Equal(ErrFunctionNotAvailable, err)
	assert.Equal(store.RevisionInvalid, rev)

	rev, err = root.Delete(ctx, true)
	require.Error(err)
	assert.Equal(ErrFunctionNotAvailable, err)
	assert.Equal(store.RevisionInvalid, rev)
}

func (ts *testSuiteCore) TestNewRegion() {
	assert := ts.Assert()
	require := ts.Require()

	stdSuffix := "TestNewRegion"
	stdDetails := ts.stdRegionDetails(stdSuffix)

	ctx := context.Background()

	region, err := NewRegion(ctx, ts.store, DefinitionTable, ts.regionName(stdSuffix))
	require.NoError(err)

	// We only expect real revision values once there has been a create 
	// or update to the store.
	//
	rev := region.GetRevision(ctx)
	assert.Equal(store.RevisionInvalid, rev)

	rev = region.GetRevisionRecord(ctx)
	assert.Equal(store.RevisionInvalid, rev)

	rev = region.GetRevisionStore(ctx)
	assert.Equal(store.RevisionInvalid, rev)

	// Now try the various combinations of setting and clearing details.
	//
	rev, details := region.GetDetails(ctx)
	assert.Equal(store.RevisionInvalid, rev)
	assert.Nil(details)

	region.SetDetails(ctx, stdDetails)

	rev, details = region.GetDetails(ctx)
	assert.Equal(store.RevisionInvalid, rev)
	require.NotNil(details)
	assert.Equal(stdDetails, details)

	region.SetDetails(ctx, nil)

	rev, details = region.GetDetails(ctx)
	assert.Equal(store.RevisionInvalid, rev)
	require.Nil(details)

	// This will actually attempt to read the region from the store. Since
	// we have yet to create the region, we expect to see a "not found"
	// type error.
	//
	rev, err = region.Read(ctx)
	require.Error(err)
	assert.Equal(ErrfRegionNotFound(region.Region), err)
	assert.Equal(store.RevisionInvalid, rev)

	rev, err = region.Update(ctx, false)
	require.Error(err)
	assert.Equal(ErrDetailsNotAvailable("region"), err)
	assert.Equal(store.RevisionInvalid, rev)

	rev, err = region.Update(ctx, true)
	require.Error(err)
	assert.Equal(ErrDetailsNotAvailable("region"), err)
	assert.Equal(store.RevisionInvalid, rev)

	// This will actually attempt to delete the region from the store. Since
	// we have yet to create the region, we expect to see a "not found"
	// type error.
	//

	// Currently, the delete of a non-existent k,v pair is succeeding.
	// I suspect this is because the delete is effectively an
	// unconditional delete as a result of the revision field being
	// set to store.RevisionInvalid. Alternatively, this could be an
	// issue in the store layer in the Delete() function where the
	// response from Etcd is being parsed. May require further
	// investigation.

	// rev, err = region.Delete(ctx, false)
	// require.Error(err)
	// assert.Equal(ErrfRegionNotFound(region.Region), err)
	// assert.Equal(store.RevisionInvalid, rev)

	// rev, err = region.Delete(ctx, true)
	// require.Error(err)
	// assert.Equal(ErrfRegionNotFound(region.Region), err)
	// assert.Equal(store.RevisionInvalid, rev)
}

func (ts *testSuiteCore) TestNewZone() {
	assert := ts.Assert()
	require := ts.Require()

	stdSuffix  := "TestNewZone"
	stdDetails := ts.stdZoneDetails(stdSuffix)

	ctx := context.Background()

	zone, err := NewZone(ctx, ts.store, DefinitionTable, ts.regionName(stdSuffix), ts.zoneName(stdSuffix))
	require.NoError(err)

	// We only expect real revision values once there has been a create 
	// or update to the store.
	//
	rev := zone.GetRevision(ctx)
	assert.Equal(store.RevisionInvalid, rev)

	rev = zone.GetRevisionRecord(ctx)
	assert.Equal(store.RevisionInvalid, rev)

	rev = zone.GetRevisionStore(ctx)
	assert.Equal(store.RevisionInvalid, rev)

	// Now try the various combinations of setting and clearing details.
	//
	rev, details := zone.GetDetails(ctx)
	assert.Equal(store.RevisionInvalid, rev)
	assert.Nil(details)

	zone.SetDetails(ctx, stdDetails)

	rev, details = zone.GetDetails(ctx)
	assert.Equal(store.RevisionInvalid, rev)
	require.NotNil(details)
	assert.Equal(stdDetails, details)

	zone.SetDetails(ctx, nil)

	rev, details = zone.GetDetails(ctx)
	assert.Equal(store.RevisionInvalid, rev)
	require.Nil(details)

	// This will actually attempt to read the zone from the store. Since
	// we have yet to create the zone, we expect to see a "not found"
	// type error.
	//
	rev, err = zone.Read(ctx)
	require.Error(err)
	assert.Equal(ErrfZoneNotFound(zone.Region, zone.Zone), err)
	assert.Equal(store.RevisionInvalid, rev)

	rev, err = zone.Update(ctx, false)
	require.Error(err)
	assert.Equal(ErrDetailsNotAvailable("zone"), err)
	assert.Equal(store.RevisionInvalid, rev)

	rev, err = zone.Update(ctx, true)
	require.Error(err)
	assert.Equal(ErrDetailsNotAvailable("zone"), err)
	assert.Equal(store.RevisionInvalid, rev)

	// This will actually attempt to delete the zone from the store. Since
	// we have yet to create the zone, we expect to see a "not found"
	// type error.
	//

	// Currently, the delete of a non-existent k,v pair is succeeding.
	// I suspect this is because the delete is effectively an
	// unconditional delete as a result of the revision field being
	// set to store.RevisionInvalid. Alternatively, this could be an
	// issue in the store layer in the Delete() function where the
	// response from Etcd is being parsed. May require further
	// investigation.

	// rev, err = zone.Delete(ctx, false)
	// require.Error(err)
	// assert.Equal(ErrfZoneNotFound(zone.Region, zone.Zone), err)
	// assert.Equal(store.RevisionInvalid, rev)

	// rev, err = zone.Delete(ctx, true)
	// require.Error(err)
	// assert.Equal(ErrfZoneNotFound(zone.Region, zone.Zone), err)
	// assert.Equal(store.RevisionInvalid, rev)
}

func (ts *testSuiteCore) TestNewRack() {
	assert := ts.Assert()
	require := ts.Require()

	stdSuffix  := "TestNewRack"
	stdDetails := ts.stdRackDetails(stdSuffix)

	ctx := context.Background()

	rack, err := NewRack(
		ctx,
		ts.store,
		DefinitionTable,
		ts.rackName(stdSuffix),
		ts.zoneName(stdSuffix),
		ts.rackName(stdSuffix),
	)
	require.NoError(err)

	// We only expect real revision values once there has been a create 
	// or update to the store.
	//
	rev := rack.GetRevision(ctx)
	assert.Equal(store.RevisionInvalid, rev)

	rev = rack.GetRevisionRecord(ctx)
	assert.Equal(store.RevisionInvalid, rev)

	rev = rack.GetRevisionStore(ctx)
	assert.Equal(store.RevisionInvalid, rev)

	// Now try the various combinations of setting and clearing details.
	//
	rev, details := rack.GetDetails(ctx)
	assert.Equal(store.RevisionInvalid, rev)
	assert.Nil(details)

	rack.SetDetails(ctx, stdDetails)

	rev, details = rack.GetDetails(ctx)
	assert.Equal(store.RevisionInvalid, rev)
	require.NotNil(details)
	assert.Equal(stdDetails, details)

	rack.SetDetails(ctx, nil)

	rev, details = rack.GetDetails(ctx)
	assert.Equal(store.RevisionInvalid, rev)
	require.Nil(details)

	// This will actually attempt to read the rack from the store. Since
	// we have yet to create the rack, we expect to see a "not found"
	// type error.
	//
	rev, err = rack.Read(ctx)
	require.Error(err)
	assert.Equal(ErrfRackNotFound(rack.Region, rack.Zone, rack.Rack), err)
	assert.Equal(store.RevisionInvalid, rev)

	rev, err = rack.Update(ctx, false)
	require.Error(err)
	assert.Equal(ErrDetailsNotAvailable("rack"), err)
	assert.Equal(store.RevisionInvalid, rev)

	rev, err = rack.Update(ctx, true)
	require.Error(err)
	assert.Equal(ErrDetailsNotAvailable("rack"), err)
	assert.Equal(store.RevisionInvalid, rev)

	// This will actually attempt to delete the rack from the store. Since
	// we have yet to create the rack, we expect to see a "not found"
	// type error.
	//

	// Currently, the delete of a non-existent k,v pair is succeeding.
	// I suspect this is because the delete is effectively an
	// unconditional delete as a result of the revision field being
	// set to store.RevisionInvalid. Alternatively, this could be an
	// issue in the store layer in the Delete() function where the
	// response from Etcd is being parsed. May require further
	// investigation.

	// rev, err = rack.Delete(ctx, false)
	// require.Error(err)
	// assert.Equal(ErrfRackNotFound(rack.Region, rack.Zone, rack.Rack), err)
	// assert.Equal(store.RevisionInvalid, rev)

	// rev, err = rack.Delete(ctx, true)
	// require.Error(err)
	// assert.Equal(ErrfRackNotFound(rack.Region, rack.Zone, rack.Rack), err)
	// assert.Equal(store.RevisionInvalid, rev)
}

func (ts *testSuiteCore) TestNewPdu() {
	assert := ts.Assert()
	require := ts.Require()

	stdSuffix  := "TestNewPdu"
	stdDetails := ts.stdPduDetails(1)
	stdPorts   := ts.stdPowerPorts(8)

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

	// We only expect real revision values once there has been a create 
	// or update to the store.
	//
	rev := pdu.GetRevision(ctx)
	assert.Equal(store.RevisionInvalid, rev)

	rev = pdu.GetRevisionRecord(ctx)
	assert.Equal(store.RevisionInvalid, rev)

	rev = pdu.GetRevisionStore(ctx)
	assert.Equal(store.RevisionInvalid, rev)

	// Now try the various combinations of setting and clearing
	// details and ports.
	//
	rev, details := pdu.GetDetails(ctx)
	assert.Equal(store.RevisionInvalid, rev)
	assert.Nil(details)

	rev, ports := pdu.GetPorts(ctx)
	assert.Equal(store.RevisionInvalid, rev)
	require.Nil(ports)


	// Now set just the details
	//
	pdu.SetDetails(ctx, stdDetails)

	rev, details = pdu.GetDetails(ctx)
	assert.Equal(store.RevisionInvalid, rev)
	require.NotNil(details)
	assert.Equal(stdDetails, details)

	rev, ports = pdu.GetPorts(ctx)
	assert.Equal(store.RevisionInvalid, rev)
	require.Nil(ports)


	// Also set the ports
	//
	pdu.SetPorts(ctx, stdPorts)

	rev, details = pdu.GetDetails(ctx)
	assert.Equal(store.RevisionInvalid, rev)
	require.NotNil(details)
	assert.Equal(stdDetails, details)

	rev, ports = pdu.GetPorts(ctx)
	assert.Equal(store.RevisionInvalid, rev)
	require.NotNil(ports)
	assert.Equal(stdPorts, ports)


	// Clear just the details
	//
	pdu.SetDetails(ctx, nil)

	rev, details = pdu.GetDetails(ctx)
	assert.Equal(store.RevisionInvalid, rev)
	assert.Nil(details)

	rev, ports = pdu.GetPorts(ctx)
	assert.Equal(store.RevisionInvalid, rev)
	require.NotNil(ports)
	assert.Equal(stdPorts, ports)


	// Now also clear the ports
	//
	pdu.SetPorts(ctx, nil)

	rev, details = pdu.GetDetails(ctx)
	assert.Equal(store.RevisionInvalid, rev)
	assert.Nil(details)

	rev, ports = pdu.GetPorts(ctx)
	assert.Equal(store.RevisionInvalid, rev)
	require.Nil(ports)


	// And then once agains, set both details and ports
	//
	pdu.SetDetails(ctx, stdDetails)
	pdu.SetPorts(ctx, stdPorts)

	rev, details = pdu.GetDetails(ctx)
	assert.Equal(store.RevisionInvalid, rev)
	require.NotNil(details)
	assert.Equal(stdDetails, details)

	rev, ports = pdu.GetPorts(ctx)
	assert.Equal(store.RevisionInvalid, rev)
	require.NotNil(ports)
	assert.Equal(stdPorts, ports)

	// This will actually attempt to read the pdu from the store. Since
	// we have yet to create the pdu, we expect to see a "not found"
	// type error.
	//
	rev, err = pdu.Read(ctx)
	require.Error(err)
	assert.Equal(ErrfPduNotFound(pdu.Region, pdu.Zone, pdu.Rack, pdu.ID), err)
	assert.Equal(store.RevisionInvalid, rev)


	// Clear the ports and check the update fails
	//
	// Note: the ordering of this and the subsequent statements assume
	//       the Update() call checks for the details being present before
	//       checking for the ports being present. Any change in the
	//       ordering may result in the tests neding amendment.
	//
	pdu.SetPorts(ctx, nil)

	rev, err = pdu.Update(ctx, false)
	require.Equal(ErrPortsNotAvailable("pdu"), err)
	assert.Equal(store.RevisionInvalid, rev)

	rev, err = pdu.Update(ctx, true)
	require.Equal(ErrPortsNotAvailable("pdu"), err)
	assert.Equal(store.RevisionInvalid, rev)

	pdu.SetDetails(ctx, nil)

	rev, err = pdu.Update(ctx, false)
	require.Error(err)
	assert.Equal(ErrDetailsNotAvailable("pdu"), err)
	assert.Equal(store.RevisionInvalid, rev)

	rev, err = pdu.Update(ctx, true)
	require.Error(err)
	assert.Equal(ErrDetailsNotAvailable("pdu"), err)
	assert.Equal(store.RevisionInvalid, rev)


	// This will actually attempt to delete the zone from the store. Since
	// we have yet to create the zone, we expect to see a "not found"
	// type error.
	//

	// Currently, the delete of a non-existent k,v pair is succeeding.
	// I suspect this is because the delete is effectively an
	// unconditional delete as a result of the revision field being
	// set to store.RevisionInvalid. Alternatively, this could be an
	// issue in the store layer in the Delete() function where the
	// response from Etcd is being parsed. May require further
	// investigation.

	// rev, err = pdu.Delete(ctx, false)
	// require.Error(err)
	// assert.Equal(ErrfPduNotFound(pdu.Region, pdu.Zone, pdu.Rack, pdu.ID), err)
	// assert.Equal(store.RevisionInvalid, rev)

	// rev, err = pdu.Delete(ctx, true)
	// require.Error(err)
	// assert.Equal(ErrfPduNotFound(pdu.Region, pdu.Zone, pdu.Rack, pdu.ID), err)
	// assert.Equal(store.RevisionInvalid, rev)
}

func (ts *testSuiteCore) TestNewTor() {
	assert := ts.Assert()
	require := ts.Require()

	stdSuffix  := "TestNewTor"
	stdDetails := ts.stdTorDetails()
	stdPorts   := ts.stdNetworkPorts(8)

	ctx := context.Background()

	tor, err := NewTor(
		ctx,
		ts.store,
		DefinitionTable,
		ts.regionName(stdSuffix),
		ts.zoneName(stdSuffix),
		ts.rackName(stdSuffix),
		ts.pduID(1),
	)
	require.NoError(err)

	// We only expect real revision values once there has been a create 
	// or update to the store.
	//
	rev := tor.GetRevision(ctx)
	assert.Equal(store.RevisionInvalid, rev)

	rev = tor.GetRevisionRecord(ctx)
	assert.Equal(store.RevisionInvalid, rev)

	rev = tor.GetRevisionStore(ctx)
	assert.Equal(store.RevisionInvalid, rev)

	// Now try the various combinations of setting and clearing
	// details and ports.
	//
	rev, details := tor.GetDetails(ctx)
	assert.Equal(store.RevisionInvalid, rev)
	assert.Nil(details)

	rev, ports := tor.GetPorts(ctx)
	assert.Equal(store.RevisionInvalid, rev)
	require.Nil(ports)


	// Now set just the details
	//
	tor.SetDetails(ctx, stdDetails)

	rev, details = tor.GetDetails(ctx)
	assert.Equal(store.RevisionInvalid, rev)
	require.NotNil(details)
	assert.Equal(stdDetails, details)

	rev, ports = tor.GetPorts(ctx)
	assert.Equal(store.RevisionInvalid, rev)
	require.Nil(ports)


	// Also set the ports
	//
	tor.SetPorts(ctx, stdPorts)

	rev, details = tor.GetDetails(ctx)
	assert.Equal(store.RevisionInvalid, rev)
	require.NotNil(details)
	assert.Equal(stdDetails, details)

	rev, ports = tor.GetPorts(ctx)
	assert.Equal(store.RevisionInvalid, rev)
	require.NotNil(ports)
	assert.Equal(stdPorts, ports)


	// Clear just the details
	//
	tor.SetDetails(ctx, nil)

	rev, details = tor.GetDetails(ctx)
	assert.Equal(store.RevisionInvalid, rev)
	assert.Nil(details)

	rev, ports = tor.GetPorts(ctx)
	assert.Equal(store.RevisionInvalid, rev)
	require.NotNil(ports)
	assert.Equal(stdPorts, ports)


	// Now also clear the ports
	//
	tor.SetPorts(ctx, nil)

	rev, details = tor.GetDetails(ctx)
	assert.Equal(store.RevisionInvalid, rev)
	assert.Nil(details)

	rev, ports = tor.GetPorts(ctx)
	assert.Equal(store.RevisionInvalid, rev)
	require.Nil(ports)


	// And then once agains, set both details and ports
	//
	tor.SetDetails(ctx, stdDetails)
	tor.SetPorts(ctx, stdPorts)

	rev, details = tor.GetDetails(ctx)
	assert.Equal(store.RevisionInvalid, rev)
	require.NotNil(details)
	assert.Equal(stdDetails, details)

	rev, ports = tor.GetPorts(ctx)
	assert.Equal(store.RevisionInvalid, rev)
	require.NotNil(ports)
	assert.Equal(stdPorts, ports)

	// This will actually attempt to read the tor from the store. Since
	// we have yet to create the zone, we expect to see a "not found"
	// type error.
	//
	rev, err = tor.Read(ctx)
	require.Error(err)
	assert.Equal(ErrfTorNotFound(tor.Region, tor.Zone, tor.Rack, tor.ID), err)
	assert.Equal(store.RevisionInvalid, rev)


	// Clear the ports and check the update fails
	//
	// Note: the ordering of this and the subsequent statements assume
	//       the Update() call checks for the details being present before
	//       checking for the ports being present. Any change in the
	//       ordering may result in the tests neding amendment.
	//
	tor.SetPorts(ctx, nil)

	rev, err = tor.Update(ctx, false)
	require.Equal(ErrPortsNotAvailable("tor"), err)
	assert.Equal(store.RevisionInvalid, rev)

	rev, err = tor.Update(ctx, true)
	require.Equal(ErrPortsNotAvailable("tor"), err)
	assert.Equal(store.RevisionInvalid, rev)

	tor.SetDetails(ctx, nil)

	rev, err = tor.Update(ctx, false)
	require.Error(err)
	assert.Equal(ErrDetailsNotAvailable("tor"), err)
	assert.Equal(store.RevisionInvalid, rev)

	rev, err = tor.Update(ctx, true)
	require.Error(err)
	assert.Equal(ErrDetailsNotAvailable("tor"), err)
	assert.Equal(store.RevisionInvalid, rev)

	// This will actually attempt to delete the tor from the store. Since
	// we have yet to create the tor, we expect to see a "not found"
	// type error.
	//

	// Currently, the delete of a non-existent k,v pair is succeeding.
	// I suspect this is because the delete is effectively an
	// unconditional delete as a result of the revision field being
	// set to store.RevisionInvalid. Alternatively, this could be an
	// issue in the store layer in the Delete() function where the
	// response from Etcd is being parsed. May require further
	// investigation.

	// rev, err = tor.Delete(ctx, false)
	// require.Error(err)
	// assert.Equal(ErrfTorNotFound(tor.Region, tor.Zone, tor.Rack, tor.ID), err)
	// assert.Equal(store.RevisionInvalid, rev)

	// rev, err = tor.Delete(ctx, true)
	// require.Error(err)
	// assert.Equal(ErrfTorNotFound(tor.Region, tor.Zone, tor.Rack, tor.ID), err)
	// assert.Equal(store.RevisionInvalid, rev)
}

func (ts *testSuiteCore) TestNewBlade() {
	assert := ts.Assert()
	require := ts.Require()

	stdSuffix   := "TestNewBlade"
	stdDetails  := ts.stdBladeDetails()
	stdCapacity := ts.stdBladeCapacity()
	stdBootInfo := ts.stdBladeBootInfo()
	
	stdBootOnPowerOn := true

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

	// We only expect real revision values once there has been a create 
	// or update to the store.
	//
	rev := blade.GetRevision(ctx)
	assert.Equal(store.RevisionInvalid, rev)

	rev = blade.GetRevisionRecord(ctx)
	assert.Equal(store.RevisionInvalid, rev)

	rev = blade.GetRevisionStore(ctx)
	assert.Equal(store.RevisionInvalid, rev)

	// Now try the various combinations of setting and clearing
	// details and ports.
	//
	rev, details := blade.GetDetails(ctx)
	assert.Equal(store.RevisionInvalid, rev)
	require.Nil(details)

	rev, capacity := blade.GetCapacity(ctx)
	assert.Equal(store.RevisionInvalid, rev)
	require.Nil(capacity)

	rev, bootOnPowerOn, bootInfo := blade.GetBootInfo(ctx)
	assert.Equal(store.RevisionInvalid, rev)
	assert.False(bootOnPowerOn)
	require.Nil(bootInfo)


	// Now set just the details
	//
	blade.SetDetails(ctx, stdDetails)

	rev, details = blade.GetDetails(ctx)
	assert.Equal(store.RevisionInvalid, rev)
	require.NotNil(details)
	assert.Equal(stdDetails, details)

	rev, capacity = blade.GetCapacity(ctx)
	assert.Equal(store.RevisionInvalid, rev)
	require.Nil(capacity)

	rev, bootOnPowerOn, bootInfo = blade.GetBootInfo(ctx)
	assert.Equal(store.RevisionInvalid, rev)
	require.Nil(bootInfo)
	assert.False(bootOnPowerOn)


	// Then set the capacity
	//
	blade.SetCapacity(ctx, stdCapacity)

	rev, details = blade.GetDetails(ctx)
	assert.Equal(store.RevisionInvalid, rev)
	require.NotNil(details)
	assert.Equal(stdDetails, details)

	rev, capacity = blade.GetCapacity(ctx)
	assert.Equal(store.RevisionInvalid, rev)
	require.NotNil(capacity)
	assert.Equal(stdCapacity, capacity)

	rev, bootOnPowerOn, bootInfo = blade.GetBootInfo(ctx)
	assert.Equal(store.RevisionInvalid, rev)
	require.Nil(bootInfo)
	assert.False(bootOnPowerOn)


	// Finally, also set the bootInfo
	//
	blade.SetBootInfo(ctx, stdBootOnPowerOn, stdBootInfo)

	rev, details = blade.GetDetails(ctx)
	assert.Equal(store.RevisionInvalid, rev)
	require.NotNil(details)
	assert.Equal(stdDetails, details)

	rev, capacity = blade.GetCapacity(ctx)
	assert.Equal(store.RevisionInvalid, rev)
	require.NotNil(capacity)
	assert.Equal(stdCapacity, capacity)

	rev, bootOnPowerOn, bootInfo = blade.GetBootInfo(ctx)
	assert.Equal(store.RevisionInvalid, rev)
	require.NotNil(bootInfo)
	assert.Equal(stdBootOnPowerOn, bootOnPowerOn)
	assert.Equal(stdBootInfo, bootInfo)


	// This will actually attempt to read the blade from the store. Since
	// we have yet to create the blade, we expect to see a "not found"
	// type error.
	//
	rev, err = blade.Read(ctx)
	require.Error(err)
	assert.Equal(ErrfPduNotFound(blade.Region, blade.Zone, blade.Rack, blade.ID), err)
	assert.Equal(store.RevisionInvalid, rev)


	// Clear the ports and check the update fails
	//
	// Note: the ordering of this and the subsequent statements assume
	//       the Update() call checks for the details being present before
	//       checking for the capacity or bootInfo being present. Any
	//	     change in the ordering may result in the tests neding amendment.
	//
	blade.SetBootInfo(ctx, false, nil)

	rev, err = blade.Update(ctx, false)
	require.Equal(ErrBootInfoNotAvailable("blade"), err)
	assert.Equal(store.RevisionInvalid, rev)

	rev, err = blade.Update(ctx, true)
	require.Equal(ErrBootInfoNotAvailable("blade"), err)
	assert.Equal(store.RevisionInvalid, rev)


	blade.SetCapacity(ctx, nil)

	rev, err = blade.Update(ctx, false)
	require.Equal(ErrCapacityNotAvailable("blade"), err)
	assert.Equal(store.RevisionInvalid, rev)

	rev, err = blade.Update(ctx, true)
	require.Equal(ErrCapacityNotAvailable("blade"), err)
	assert.Equal(store.RevisionInvalid, rev)


	blade.SetDetails(ctx, nil)

	rev, err = blade.Update(ctx, false)
	require.Error(err)
	assert.Equal(ErrDetailsNotAvailable("blade"), err)
	assert.Equal(store.RevisionInvalid, rev)

	rev, err = blade.Update(ctx, true)
	require.Error(err)
	assert.Equal(ErrDetailsNotAvailable("blade"), err)
	assert.Equal(store.RevisionInvalid, rev)


	// This will actually attempt to delete the blade from the store. Since
	// we have yet to create the blade, we expect to see a "not found"
	// type error.
	//

	// Currently, the delete of a non-existent k,v pair is succeeding.
	// I suspect this is because the delete is effectively an
	// unconditional delete as a result of the revision field being
	// set to store.RevisionInvalid. Alternatively, this could be an
	// issue in the store layer in the Delete() function where the
	// response from Etcd is being parsed. May require further
	// investigation.

	// rev, err = blade.Delete(ctx, false)
	// require.Error(err)
	// assert.Equal(ErrfBladeNotFound(blade.Region, blade.Zone, blade.Rack, blade.ID), err)
	// assert.Equal(store.RevisionInvalid, rev)

	// rev, err = blade.Delete(ctx, true)
	// require.Error(err)
	// assert.Equal(ErrfBladeNotFound(blade.Region, blade.Zone, blade.Rack, blade.ID), err)
	// assert.Equal(store.RevisionInvalid, rev)
}

func (ts *testSuiteCore) TestNewRegionWithCreate() {
	assert := ts.Assert()
	require := ts.Require()

	stdSuffix  := "TestNewRegionWithCreate"
	stdDetails := ts.stdRegionDetails(stdSuffix)

	ctx := context.Background()

	r, err := NewRegion(ctx, ts.store, DefinitionTable, ts.regionName(stdSuffix))
	require.NoError(err)

	r.SetDetails(ctx, stdDetails)

	rev, err := r.Create(ctx)

	require.NoError(err)
	assert.NotEqual(store.RevisionInvalid, rev)
}

func (ts *testSuiteCore) TestNewZoneWithCreate() {
	assert := ts.Assert()
	require := ts.Require()

	stdSuffix  := "TestNewZoneWithCreate"
	stdDedails := ts.stdZoneDetails(stdSuffix)

	ctx := context.Background()

	zone, err := NewZone(ctx, ts.store, DefinitionTable, ts.regionName(stdSuffix), ts.zoneName(stdSuffix))
	require.NoError(err)

	zone.SetDetails(ctx, stdDedails)

	rev, err := zone.Create(ctx)

	require.NoError(err)
	assert.NotEqual(store.RevisionInvalid, rev)
}

func (ts *testSuiteCore) TestNewRackWithCreate() {
	assert := ts.Assert()
	require := ts.Require()

	stdSuffix := "TestNewRackWithCreate"
	stdDetails := ts.stdRackDetails(stdSuffix)

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

	rack.SetDetails(ctx, stdDetails)

	rev, err := rack.Create(ctx)

	require.NoError(err)
	assert.NotEqual(store.RevisionInvalid, rev)
}

func (ts *testSuiteCore) TestNewPduWithCreate() {
	assert := ts.Assert()
	require := ts.Require()

	stdSuffix  := "TestNewPduWithCreate"
	stdDetails := ts.stdPduDetails(4)
	stdPorts   := ts.stdPowerPorts(4)

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

	pdu.SetDetails(ctx, stdDetails)
	pdu.SetPorts(ctx, stdPorts)

	rev, err := pdu.Create(ctx)

	require.NoError(err)
	assert.NotEqual(store.RevisionInvalid, rev)
}

func (ts *testSuiteCore) TestNewTorWithCreate() {
	assert := ts.Assert()
	require := ts.Require()

	stdSuffix  := "TestNewTorWithCreate"
	stdDetails := ts.stdTorDetails()
	stdPorts   := ts.stdNetworkPorts(4)

	ctx := context.Background()

	tor, err := NewTor(
		ctx,
		ts.store,
		DefinitionTable,
		ts.regionName(stdSuffix),
		ts.zoneName(stdSuffix),
		ts.rackName(stdSuffix),
		ts.pduID(1),
	)
	require.NoError(err)

	tor.SetDetails(ctx,stdDetails)
	tor.SetPorts(ctx, stdPorts)

	rev, err := tor.Create(ctx)

	require.NoError(err)
	assert.NotEqual(store.RevisionInvalid, rev)
}

func (ts *testSuiteCore) TestNewBladeWithCreate() {
	assert := ts.Assert()
	require := ts.Require()

	stdSuffix        := "TestNewBladeWithCreate"
	stdDetails       := ts.stdBladeDetails()
	stdCapacity      := ts.stdBladeCapacity()
	stdBootInfo      := ts.stdBladeBootInfo()
	stdBootOnPowerOn := true

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

	blade.SetDetails(ctx, stdDetails)
	blade.SetCapacity(ctx, stdCapacity)
	blade.SetBootInfo(ctx, stdBootOnPowerOn, stdBootInfo)

	rev, err := blade.Create(ctx)

	require.NoError(err)
	assert.NotEqual(store.RevisionInvalid, rev)
}

func (ts *testSuiteCore) TestRootNewChild() {
	assert := ts.Assert()
	require := ts.Require()

	stdSuffix := "TestRootNewChild"

	regionName := ts.regionName(stdSuffix)

	ctx := context.Background()

	root, err := NewRoot (ctx, ts.store, DefinitionTable)
	require.NoError(err)

	root.SetDetails(ctx, ts.stdRootDetails(stdSuffix))

	region, err := root.NewChild(ctx, regionName)
	require.NoError(err)

	region.SetDetails(ctx, ts.stdRegionDetails(stdSuffix))

	rev, err := region.Create(ctx)
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

	zone.SetDetails(ctx,ts.stdZoneDetails(stdSuffix))

	rev, err := zone.Create(ctx)
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

	rack.SetDetails(ctx, ts.stdRackDetails(stdSuffix))

	rev, err := rack.Create(ctx)
	require.NoError(err)
	assert.NotEqual(store.RevisionInvalid, rev)
}

func (ts *testSuiteCore) TestNewChildPdu() {
	assert := ts.Assert()
	require := ts.Require()

	stdSuffix := "TestNewChildPdu"

	regionName := ts.regionName(stdSuffix)
	zoneName   := ts.zoneName(stdSuffix)
	rackName   := ts.rackName(stdSuffix)
	pduID      := ts.pduID(1)

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

	pdu.SetDetails(ctx, ts.stdPduDetails(pduID))
	pdu.SetPorts(ctx, ts.stdPowerPorts(8))

	rev, err := pdu.Create(ctx)
	require.NoError(err)
	assert.NotEqual(store.RevisionInvalid, rev)
}

func (ts *testSuiteCore) TestNewChildTor() {
	assert := ts.Assert()
	require := ts.Require()

	stdSuffix := "TestNewChildTor"

	regionName := ts.regionName(stdSuffix)
	zoneName   := ts.zoneName(stdSuffix)
	rackName   := ts.rackName(stdSuffix)
	pduID      := ts.pduID(1)

	ctx := context.Background()

	root, err := NewRoot (ctx, ts.store, DefinitionTable)
	require.NoError(err)

	region, err := root.NewChild(ctx, regionName)
	require.NoError(err)

	zone, err := region.NewChild(ctx, zoneName)
	require.NoError(err)

	rack, err := zone.NewChild(ctx, rackName)
	require.NoError(err)

	tor, err := rack.NewTor(ctx, pduID)
	require.NoError(err)

	tor.SetDetails(ctx, ts.stdTorDetails())
	tor.SetPorts(ctx, ts.stdNetworkPorts(8))

	rev, err := tor.Create(ctx)
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

	stdDetails  := ts.stdBladeDetails()
	stdCapacity := ts.stdBladeCapacity()
	stdBootInfo := ts.stdBladeBootInfo()
	
	stdBootOnPowerOn := true


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

	blade.SetDetails(ctx, stdDetails)
	blade.SetCapacity(ctx, stdCapacity)
	blade.SetBootInfo(ctx, stdBootOnPowerOn, stdBootInfo)

	rev, err := blade.Create(ctx)
	require.NoError(err)
	assert.NotEqual(store.RevisionInvalid, rev)
}

func (ts *testSuiteCore) TestRegionReadDetails() {
	assert := ts.Assert()
	require := ts.Require()

	stdSuffix := "TestRegionReadDetails"

	regionName := ts.regionName(stdSuffix)

	stdDetails := ts.stdRegionDetails(stdSuffix)

	ctx := context.Background()

	r, err := NewRegion(ctx, ts.store, DefinitionTable, regionName)
	require.NoError(err)

	r.SetDetails(ctx, stdDetails)

	rev, err := r.Create(ctx)
	require.NoError(err)
	assert.NotEqual(store.RevisionInvalid, rev)

	rev2 := r.GetRevision(ctx)
	assert.Equal(rev, rev2)

	// Read the region back using the direct constructor
	//
	rRead, err := NewRegion(ctx, ts.store, DefinitionTable, regionName)
	require.NoError(err)

	revRead, err := rRead.Read(ctx)
	require.NoError(err)
	assert.Equal(rev, revRead)
	assert.Equal(revRead, rRead.GetRevision(ctx))

	revDet, detRead := rRead.GetDetails(ctx)
	require.NoError(err)
	assert.Equal(rev, revDet)
	assert.Equal(stdDetails, detRead)


	// Read the region back using the relative constructor
	//
	root, err := NewRoot (ctx, ts.store, DefinitionTable)
	require.NoError(err)

	cr, err := root.NewChild(ctx, regionName)
	require.NoError(err)

	crRev, err := cr.Read(ctx)
	require.NoError(err)
	assert.Equal(rev, crRev)
	assert.Equal(revRead, cr.GetRevision(ctx))

	crRevDet, crDet := cr.GetDetails(ctx)
	require.NoError(err)
	assert.Equal(rev, crRevDet)
	assert.Equal(stdDetails, crDet)
}

func (ts *testSuiteCore) TestZoneReadDetails() {
	assert := ts.Assert()
	require := ts.Require()

	stdSuffix := "TestZoneReadDetails"

	regionName := ts.regionName(stdSuffix)
	zoneName   := ts.zoneName(stdSuffix)

	stdDetails := ts.stdZoneDetails(stdSuffix)

	ctx := context.Background()

	z, err := NewZone(ctx, ts.store, DefinitionTable, regionName, zoneName)
	require.NoError(err)

	z.SetDetails(ctx, stdDetails)

	rev, err := z.Create(ctx)
	require.NoError(err)
	assert.NotEqual(store.RevisionInvalid, rev)

	rev2 := z.GetRevision(ctx)
	assert.Equal(rev, rev2)


	// Read the zone back using the direct constructor
	//
	rRead, err := NewZone(ctx, ts.store, DefinitionTable, regionName, zoneName)
	require.NoError(err)

	revRead, err := rRead.Read(ctx)
	require.NoError(err)
	assert.Equal(rev, revRead)
	assert.Equal(revRead, rRead.GetRevision(ctx))

	revDet, detRead := rRead.GetDetails(ctx)
	require.NoError(err)
	assert.Equal(rev, revDet)
	assert.Equal(stdDetails, detRead)


	// Read the zone back using the relative constructor
	//
	root, err := NewRoot (ctx, ts.store, DefinitionTable)
	require.NoError(err)

	region, err := root.NewChild(ctx, regionName)
	require.NoError(err)

	zone, err := region.NewChild(ctx, zoneName)
	require.NoError(err)

	zoneRev, err := zone.Read(ctx)
	require.NoError(err)
	assert.Equal(rev, zoneRev)
	assert.Equal(zoneRev, zone.GetRevision(ctx))

	zoneDetRev, zoneDet := zone.GetDetails(ctx)
	require.NoError(err)
	assert.Equal(rev, zoneDetRev)
	assert.Equal(stdDetails, zoneDet)
}

func (ts *testSuiteCore) TestRackReadDetails() {
	assert := ts.Assert()
	require := ts.Require()

	stdSuffix := "TestRackReadDetails"

	regionName := ts.regionName(stdSuffix)
	zoneName   := ts.zoneName(stdSuffix)
	rackName   := ts.rackName(stdSuffix)

	stdDetails := ts.stdRackDetails(stdSuffix)

	ctx := context.Background()

	r, err := NewRack(
		ctx,
		ts.store,
		DefinitionTable,
		regionName,
		zoneName,
		rackName,
	)
	require.NoError(err)

	r.SetDetails(ctx, stdDetails)

	rev, err := r.Create(ctx)
	require.NoError(err)
	assert.NotEqual(store.RevisionInvalid, rev)

	rev2 := r.GetRevision(ctx)
	assert.Equal(rev, rev2)


	// Read the region back using the direct constructor
	//
	rRead, err := NewRack(
		ctx,
		ts.store,
		DefinitionTable,
		regionName,
		zoneName,
		rackName,
	)
	require.NoError(err)

	revRead, err := rRead.Read(ctx)
	require.NoError(err)
	assert.Equal(rev, revRead)
	assert.Equal(revRead, rRead.GetRevision(ctx))

	revDet, detRead := rRead.GetDetails(ctx)
	require.NoError(err)
	assert.Equal(rev, revDet)
	assert.Equal(stdDetails, detRead)


	// Read the zone back using the relative constructor
	//
	root, err := NewRoot (ctx, ts.store, DefinitionTable)
	require.NoError(err)

	region, err := root.NewChild(ctx, regionName)
	require.NoError(err)

	zone, err := region.NewChild(ctx, zoneName)
	require.NoError(err)

	rack, err := zone.NewChild(ctx, rackName)
	require.NoError(err)

	rackRev, err := rack.Read(ctx)
	require.NoError(err)
	assert.Equal(rev, rackRev)
	assert.Equal(rackRev, rack.GetRevision(ctx))

	rackDetRev, rackDet := rack.GetDetails(ctx)
	require.NoError(err)
	assert.Equal(rev, rackDetRev)
	assert.Equal(stdDetails, rackDet)
}

func (ts *testSuiteCore) TestPduReadDetails() {
	assert := ts.Assert()
	require := ts.Require()

	stdSuffix := "TestPduReadDetails"

	regionName := ts.regionName(stdSuffix)
	zoneName   := ts.zoneName(stdSuffix)
	rackName   := ts.rackName(stdSuffix)
	pduID      := int64(4)

	stdDetails := ts.stdPduDetails(pduID)
	stdPorts   := ts.stdPowerPorts(4)

	ctx := context.Background()

	p, err := NewPdu(
		ctx,
		ts.store,
		DefinitionTable,
		regionName,
		zoneName,
		rackName,
		pduID,
	)
	require.NoError(err)

	p.SetDetails(ctx, stdDetails)
	p.SetPorts(ctx, stdPorts)

	rev, err := p.Create(ctx)
	require.NoError(err)
	assert.NotEqual(store.RevisionInvalid, rev)

	rev2 := p.GetRevision(ctx)
	assert.Equal(rev, rev2)


	// Read the region back using the direct constructor
	//
	p2, err := NewPdu(
		ctx,
		ts.store,
		DefinitionTable,
		regionName,
		zoneName,
		rackName,
		pduID,
	)
	require.NoError(err)

	p2Rev, err := p2.Read(ctx)
	require.NoError(err)
	assert.Equal(rev, p2Rev)
	assert.Equal(p2Rev, p2.GetRevision(ctx))

	p2DetRev, p2Det := p2.GetDetails(ctx)
	require.NoError(err)
	assert.Equal(rev, p2DetRev)
	assert.Equal(stdDetails, p2Det)


	// Read the zone back using the relative constructor
	//
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

	pduRev, err := pdu.Read(ctx)
	require.NoError(err)
	assert.Equal(rev, pduRev)
	assert.Equal(pduRev, pdu.GetRevision(ctx))

	pduDetRev, pduDet := pdu.GetDetails(ctx)
	require.NoError(err)
	assert.Equal(rev, pduDetRev)
	assert.Equal(stdDetails, pduDet)

	pduPortsRev, pduPorts := pdu.GetPorts(ctx)
	require.NoError(err)
	assert.Equal(rev, pduPortsRev)
	assert.Equal(stdPorts, pduPorts)
}

func (ts *testSuiteCore) TestTorReadDetails() {
	assert := ts.Assert()
	require := ts.Require()

	stdSuffix := "TestTorReadDetails"

	regionName := ts.regionName(stdSuffix)
	zoneName   := ts.zoneName(stdSuffix)
	rackName   := ts.rackName(stdSuffix)
	torID      := int64(1)

	stdDetails := ts.stdTorDetails()
	stdPorts   := ts.stdNetworkPorts(4)

	ctx := context.Background()

	t, err := NewTor(
		ctx,
		ts.store,
		DefinitionTable,
		regionName,
		zoneName,
		rackName,
		torID,
	)
	require.NoError(err)

	t.SetDetails(ctx, stdDetails)
	t.SetPorts(ctx, stdPorts)

	rev, err := t.Create(ctx)
	require.NoError(err)
	assert.NotEqual(store.RevisionInvalid, rev)

	rev2 := t.GetRevision(ctx)
	assert.Equal(rev, rev2)


	// Read the region back using the direct constructor
	//
	t2, err := NewTor(
		ctx,
		ts.store,
		DefinitionTable,
		regionName,
		zoneName,
		rackName,
		torID,
	)
	require.NoError(err)

	t2Rev, err := t2.Read(ctx)
	require.NoError(err)
	assert.Equal(rev, t2Rev)
	assert.Equal(t2Rev, t2.GetRevision(ctx))

	t2DetRev, p2Det := t2.GetDetails(ctx)
	require.NoError(err)
	assert.Equal(rev, t2DetRev)
	assert.Equal(stdDetails, p2Det)


	// Read the zone back using the relative constructor
	//
	root, err := NewRoot (ctx, ts.store, DefinitionTable)
	require.NoError(err)

	region, err := root.NewChild(ctx, regionName)
	require.NoError(err)

	zone, err := region.NewChild(ctx, zoneName)
	require.NoError(err)

	rack, err := zone.NewChild(ctx, rackName)
	require.NoError(err)

	tor, err := rack.NewTor(ctx, torID)
	require.NoError(err)

	torRev, err := tor.Read(ctx)
	require.NoError(err)
	assert.Equal(rev, torRev)
	assert.Equal(torRev, tor.GetRevision(ctx))

	torDetRev, pduDet := tor.GetDetails(ctx)
	require.NoError(err)
	assert.Equal(rev, torDetRev)
	assert.Equal(stdDetails, pduDet)

	torPortsRev, torPorts := tor.GetPorts(ctx)
	require.NoError(err)
	assert.Equal(rev, torPortsRev)
	assert.Equal(stdPorts, torPorts)
}

func (ts *testSuiteCore) TestBladeReadDetails() {
	assert := ts.Assert()
	require := ts.Require()

	stdSuffix        := "TestBladeReadDetails"

	regionName       := ts.regionName(stdSuffix)
	zoneName         := ts.zoneName(stdSuffix)
	rackName         := ts.rackName(stdSuffix)
	bladeID          := int64(5)

	stdDetails       := ts.stdBladeDetails()
	stdCapacity      := ts.stdBladeCapacity()
	stdBootInfo      := ts.stdBladeBootInfo()
	stdBootOnPowerOn := true

	ctx := context.Background()

	b, err := NewBlade(
		ctx,
		ts.store,
		DefinitionTable,
		regionName,
		zoneName,
		rackName,
		bladeID,
	)
	require.NoError(err)

	b.SetDetails(ctx, stdDetails)
	b.SetCapacity(ctx, stdCapacity)
	b.SetBootInfo(ctx, stdBootOnPowerOn, stdBootInfo)

	rev, err := b.Create(ctx)
	require.NoError(err)
	assert.NotEqual(store.RevisionInvalid, rev)

	rev2 := b.GetRevision(ctx)
	assert.Equal(rev, rev2)


	// Read the region back using the direct constructor
	//
	b2, err := NewBlade(
		ctx,
		ts.store,
		DefinitionTable,
		regionName,
		zoneName,
		rackName,
		bladeID,
	)
	require.NoError(err)

	t2Rev, err := b2.Read(ctx)
	require.NoError(err)
	assert.Equal(rev, t2Rev)
	assert.Equal(t2Rev, b2.GetRevision(ctx))

	t2DetRev, p2Det := b2.GetDetails(ctx)
	require.NoError(err)
	assert.Equal(rev, t2DetRev)
	assert.Equal(stdDetails, p2Det)


	// Read the zone back using the relative constructor
	//
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

	bladeRev, err := blade.Read(ctx)
	require.NoError(err)
	assert.Equal(rev, bladeRev)
	assert.Equal(bladeRev, blade.GetRevision(ctx))

	bladeDetRev, bladeDet := blade.GetDetails(ctx)
	require.NoError(err)
	assert.Equal(rev, bladeDetRev)
	assert.Equal(stdDetails, bladeDet)

	bladeCapacityRev, bladeCapacity := blade.GetCapacity(ctx)
	require.NoError(err)
	assert.Equal(rev, bladeCapacityRev)
	assert.Equal(stdCapacity, bladeCapacity)

	bladeBootinfoRev, bladeBootOnPowerOn, bladeBootInfo := blade.GetBootInfo(ctx)
	require.NoError(err)
	assert.Equal(rev, bladeBootinfoRev)
	assert.Equal(stdBootOnPowerOn, bladeBootOnPowerOn)
	assert.Equal(stdBootInfo, bladeBootInfo)
}

func TestInventoryTestSuite(t *testing.T) {
	suite.Run(t, new(testSuiteCore))
}
