package inventory

import (
	"context"
	"flag"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Jim3Things/CloudChamber/internal/clients/store"
	"github.com/Jim3Things/CloudChamber/internal/config"
	"github.com/Jim3Things/CloudChamber/internal/tracing/exporters"
	"github.com/Jim3Things/CloudChamber/pkg/errors"
	pb "github.com/Jim3Things/CloudChamber/pkg/protos/inventory"
)

type testSuiteCore struct {

	suite.Suite

	baseURI string

	utf *exporters.Exporter

	cfg *config.GlobalConfig

	store *store.Store

	regionCount    int
	zonesPerRegion int
	racksPerZone   int
	pdusPerRack    int
	torsPerRack    int
	bladesPerRack  int

	portsPerPdu    int
	portsPerTor    int
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
		Location: "StdRegion: DC-PNW-" + suffix,
		Notes:    "region for inventory definition test",
	}
}

func (ts *testSuiteCore) stdZoneDetails(suffix string) *pb.ZoneDetails {
	return &pb.ZoneDetails{
		Enabled:  true,
		State:    pb.State_in_service,
		Location: "StdZone: DC-PNW-" + suffix,
		Notes:    "zone for inventory definition test",
	}
}

func (ts *testSuiteCore) stdRackDetails(suffix string) *pb.RackDetails {
	return &pb.RackDetails{
		Enabled:   true,
		Condition: pb.Condition_operational,
		Location:  "StdRack: DC-PNW-" + suffix,
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

func (ts *testSuiteCore)createStandardInventory(ctx context.Context) error {

	err := ts.createInventory(
		ctx,
		DefinitionTableStdTest,
		ts.regionCount,
		ts.zonesPerRegion,
		ts.racksPerZone,
		ts.pdusPerRack,
		ts.torsPerRack,
		ts.bladesPerRack)

	return err
}

func (ts *testSuiteCore)verifyStandardInventoryRegionDetails(name string, details *pb.RegionDetails) {
	assert := ts.Assert()

	check := ts.stdRegionDetails(name)

	assert.Equal(check.Name,     details.Name)
	assert.Equal(check.State,    details.State)
	assert.Equal(check.Location, details.Location)
	assert.Equal(check.Notes,    details.Notes)
}

func (ts *testSuiteCore)verifyStandardInventoryZoneDetails(name string, details *pb.ZoneDetails) {
	assert := ts.Assert()

	check := ts.stdZoneDetails(name)

	assert.Equal(check.Enabled,  details.Enabled)
	assert.Equal(check.State,    details.State)
	assert.Equal(check.Location, details.Location)
	assert.Equal(check.Notes,    details.Notes)
}

func (ts *testSuiteCore)verifyStandardInventoryRackDetails(name string, details *pb.RackDetails) {
	assert := ts.Assert()

	check := ts.stdRackDetails(name)

	assert.Equal(check.Enabled,   details.Enabled)
	assert.Equal(check.Condition, details.Condition)
	assert.Equal(check.Location,  details.Location)
	assert.Equal(check.Notes,     details.Notes)
}

func (ts *testSuiteCore) verifyStandardInventoryPdu(index int64, pdu Pdu) {
	assert := ts.Assert()

	details := ts.stdPduDetails(index)

	assert.Equal(details.Enabled,   pdu.details.Enabled)
	assert.Equal(details.Condition, pdu.details.Condition)

	ports := ts.stdPowerPorts(ts.portsPerPdu)

	assert.Equal(len(*ports), len(*pdu.ports))

	for k, v := range *ports {
		p := (*pdu.ports)[k]

		assert.Equal(v.Wired,     p.Wired)
		assert.Equal(v.Item.Type, p.Item.Type)
		assert.Equal(v.Item.Id,   p.Item.Id)
		assert.Equal(v.Item.Port, p.Item.Port)
	}
}

func (ts *testSuiteCore) verifyStandardInventoryTor(index int64, tor *Tor) {
	assert := ts.Assert()

	details := ts.stdTorDetails()

	assert.Equal(details.Enabled,   tor.details.Enabled)
	assert.Equal(details.Condition, tor.details.Condition)

	ports := ts.stdNetworkPorts(ts.portsPerTor)

	assert.Equal(len(*ports), len(*tor.ports))

	for k, v := range *ports {
		p := (*tor.ports)[k]

		assert.Equal(v.Wired,     p.Wired)
		assert.Equal(v.Item.Type, p.Item.Type)
		assert.Equal(v.Item.Id,   p.Item.Id)
		assert.Equal(v.Item.Port, p.Item.Port)
	}
}

func (ts *testSuiteCore) verifyStandardInventoryBlade(index int64, blade *Blade) {
	assert := ts.Assert()

	details := ts.stdTorDetails()

	assert.Equal(details.Enabled, blade.details.Enabled)
	assert.Equal(details.Condition, blade.details.Condition)

	capacity := ts.stdBladeCapacity()

	assert.Equal(capacity.Cores,                  blade.capacity.Cores)
	assert.Equal(capacity.MemoryInMb,             blade.capacity.MemoryInMb)
	assert.Equal(capacity.DiskInGb,               blade.capacity.DiskInGb)
	assert.Equal(capacity.NetworkBandwidthInMbps, blade.capacity.NetworkBandwidthInMbps)
	assert.Equal(capacity.Arch,                   blade.capacity.Arch)
	assert.Equal(capacity.Accelerators,           blade.capacity.Accelerators)

	bootInfo := ts.stdBladeBootInfo()
	assert.Equal(bootInfo.Source,     blade.bootInfo.Source)
	assert.Equal(bootInfo.Image,      blade.bootInfo.Image)
	assert.Equal(bootInfo.Version,    blade.bootInfo.Version)
	assert.Equal(bootInfo.Parameters, blade.bootInfo.Parameters)

	assert.True(blade.bootOnPowerOn)
}

func (ts *testSuiteCore)createInventory(
	ctx context.Context,
	table string,
	regions int,
	zonesPerRegion int,
	racksPerZone int,
	pdusPerRack int,
	torsPerRack int,
	bladesPerRack int) error {

	root, err := NewRoot(ctx, ts.store, table)
	if err != nil {
		return err
	}

	for i := 1; i <= regions; i++ {
		regionName := fmt.Sprintf("Region-%d", i)

		region, err := root.NewChild(ctx, regionName)

		region.SetDetails(ctx, ts.stdRegionDetails(regionName))

		_, err = region.Create(ctx)
		if err != nil {
			return err
		}
	
		for j := 1; j <= zonesPerRegion; j++ {
			zoneName := fmt.Sprintf("Zone-%d-%d", i, j)

			zone, err := region.NewChild(ctx, zoneName)
			zone.SetDetails(ctx, ts.stdZoneDetails(zoneName))
	
			_, err = zone.Create(ctx)
			if err != nil {
				return err
			}

			for k := 1; k <= racksPerZone; k++ {
				rackName := fmt.Sprintf("Rack-%d-%d-%d", i, j, k)

				rack, err := zone.NewChild(ctx, rackName)
		
				rack.SetDetails(ctx, ts.stdRackDetails(rackName))
		
				_, err = rack.Create(ctx)
				if err != nil {
					return err
				}

				for p := 0; p < pdusPerRack; p++ {
					pdu, err := rack.NewPdu(ctx, int64(p))
					if err != nil {
						return err
					}

					pdu.SetDetails(ctx, ts.stdPduDetails(1))
					pdu.SetPorts(ctx, ts.stdPowerPorts(ts.portsPerPdu))
	
					_, err = pdu.Create(ctx)
					if err != nil {
						return err
					}
				}

				for t := 0; t < torsPerRack; t++ {
					tor, err := rack.NewTor(ctx, int64(t))
					if err != nil {
						return err
					}

					tor.SetDetails(ctx, ts.stdTorDetails())
					tor.SetPorts(ctx, ts.stdNetworkPorts(ts.portsPerTor))
	
					_, err = tor.Create(ctx)
					if err != nil {
						return err
					}
				}

				for b := 0; b < bladesPerRack; b++ {
					blade, err := rack.NewBlade(ctx, int64(b))
					if err != nil {
						return err
					}

					blade.SetDetails(ctx, ts.stdBladeDetails())
					blade.SetCapacity(ctx, ts.stdBladeCapacity())
					blade.SetBootInfo(ctx, true, ts.stdBladeBootInfo())
	
					_, err = blade.Create(ctx)
					if err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

func (ts *testSuiteCore) SetupSuite() {
	require := ts.Require()

	ctx := context.Background()

	configPath := flag.String("config", "./testdata", "path to the configuration file")
	flag.Parse()

	cfg, err := config.ReadGlobalConfig(*configPath)
	require.NoError(err, "failed to process the global configuration")

	ts.utf = exporters.NewExporter(exporters.NewUTForwarder())
	exporters.ConnectToProvider(ts.utf)

	store.Initialize(cfg)

	ts.cfg   = cfg
	ts.store = store.NewStore()

	// These values are relatively arbitrary. The only criteria is that different
	// constants were chosen to help separate different multiples of different
	// object types where possible and not to have values which are too large
	// to avoid lots of IO when setting up the test suite.
	//
	ts.regionCount    = 2
	ts.zonesPerRegion = 3
	ts.racksPerZone   = 4
	ts.pdusPerRack    = 1
	ts.torsPerRack    = 1
	ts.bladesPerRack  = 5

	ts.portsPerPdu    = ts.torsPerRack + ts.bladesPerRack
	ts.portsPerTor    = ts.pdusPerRack + ts.torsPerRack + ts.bladesPerRack

	require.NoError(ts.utf.Open(ts.T()))
	require.NoError(ts.store.Connect())

	err = ts.createStandardInventory(ctx)
	require.NoError(err, "failed to create standard inventory")

	ts.store.Disconnect()
	ts.utf.Close()
}

func (ts *testSuiteCore) SetupTest() {
	require := ts.Require()

	require.NoError(ts.utf.Open(ts.T()))

	require.NoError(ts.store.Connect())
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
	rev := root.GetRevision()
	assert.Equal(store.RevisionInvalid, rev)

	rev = root.GetRevisionRecord()
	assert.Equal(store.RevisionInvalid, rev)

	rev = root.GetRevisionStore()
	assert.Equal(store.RevisionInvalid, rev)

	// Now try the various combinations of setting and clearing details.
	//
	details := root.GetDetails(ctx)
	assert.Nil(details)

	root.SetDetails(ctx, stdDetails)

	details = root.GetDetails(ctx)
	require.NotNil(details)
	assert.Equal(stdDetails, details)

	root.SetDetails(ctx, nil)

	details = root.GetDetails(ctx)
	require.Nil(details)

	rev, err = root.Read(ctx)
	require.Error(err)
	assert.Equal(errors.ErrFunctionNotAvailable, err)
	assert.Equal(store.RevisionInvalid, rev)

	rev, err = root.Update(ctx, false)
	require.Error(err)
	assert.Equal(errors.ErrFunctionNotAvailable, err)
	assert.Equal(store.RevisionInvalid, rev)

	rev, err = root.Update(ctx, true)
	require.Error(err)
	assert.Equal(errors.ErrFunctionNotAvailable, err)
	assert.Equal(store.RevisionInvalid, rev)

	rev, err = root.Delete(ctx, false)
	require.Error(err)
	assert.Equal(errors.ErrFunctionNotAvailable, err)
	assert.Equal(store.RevisionInvalid, rev)

	rev, err = root.Delete(ctx, true)
	require.Error(err)
	assert.Equal(errors.ErrFunctionNotAvailable, err)
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
	rev := region.GetRevision()
	assert.Equal(store.RevisionInvalid, rev)

	rev = region.GetRevisionRecord()
	assert.Equal(store.RevisionInvalid, rev)

	rev = region.GetRevisionStore()
	assert.Equal(store.RevisionInvalid, rev)

	// Now try the various combinations of setting and clearing details.
	//
	details := region.GetDetails(ctx)
	assert.Nil(details)

	region.SetDetails(ctx, stdDetails)

	details = region.GetDetails(ctx)
	require.NotNil(details)
	assert.Equal(stdDetails, details)

	region.SetDetails(ctx, nil)

	details = region.GetDetails(ctx)
	require.Nil(details)

	// This will actually attempt to read the region from the store. Since
	// we have yet to create the region, we expect to see a "not found"
	// type error.
	//
	rev, err = region.Read(ctx)
	require.Error(err)
	assert.Equal(errors.ErrRegionNotFound{Region: region.Region}, err)
	assert.Equal(store.RevisionInvalid, rev)

	rev, err = region.Update(ctx, false)
	require.Error(err)
	assert.Equal(errors.ErrDetailsNotAvailable("region"), err)
	assert.Equal(store.RevisionInvalid, rev)

	rev, err = region.Update(ctx, true)
	require.Error(err)
	assert.Equal(errors.ErrDetailsNotAvailable("region"), err)
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
	// assert.Equal(errors.ErrfRegionNotFound(region.Region), err)
	// assert.Equal(store.RevisionInvalid, rev)

	// rev, err = region.Delete(ctx, true)
	// require.Error(err)
	// assert.Equal(errors.ErrfRegionNotFound(region.Region), err)
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
	rev := zone.GetRevision()
	assert.Equal(store.RevisionInvalid, rev)

	rev = zone.GetRevisionRecord()
	assert.Equal(store.RevisionInvalid, rev)

	rev = zone.GetRevisionStore()
	assert.Equal(store.RevisionInvalid, rev)

	// Now try the various combinations of setting and clearing details.
	//
	details := zone.GetDetails(ctx)
	assert.Nil(details)

	zone.SetDetails(ctx, stdDetails)

	details = zone.GetDetails(ctx)
	require.NotNil(details)
	assert.Equal(stdDetails, details)

	zone.SetDetails(ctx, nil)

	details = zone.GetDetails(ctx)
	require.Nil(details)

	// This will actually attempt to read the zone from the store. Since
	// we have yet to create the zone, we expect to see a "not found"
	// type error.
	//
	rev, err = zone.Read(ctx)
	require.Error(err)
	assert.Equal(errors.ErrZoneNotFound{Region: zone.Region, Zone: zone.Zone}, err)
	assert.Equal(store.RevisionInvalid, rev)

	rev, err = zone.Update(ctx, false)
	require.Error(err)
	assert.Equal(errors.ErrDetailsNotAvailable("zone"), err)
	assert.Equal(store.RevisionInvalid, rev)

	rev, err = zone.Update(ctx, true)
	require.Error(err)
	assert.Equal(errors.ErrDetailsNotAvailable("zone"), err)
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
	// assert.Equal(errors.ErrfZoneNotFound(zone.Region, zone.Zone), err)
	// assert.Equal(store.RevisionInvalid, rev)

	// rev, err = zone.Delete(ctx, true)
	// require.Error(err)
	// assert.Equal(errors.ErrfZoneNotFound(zone.Region, zone.Zone), err)
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
	rev := rack.GetRevision()
	assert.Equal(store.RevisionInvalid, rev)

	rev = rack.GetRevisionRecord()
	assert.Equal(store.RevisionInvalid, rev)

	rev = rack.GetRevisionStore()
	assert.Equal(store.RevisionInvalid, rev)

	// Now try the various combinations of setting and clearing details.
	//
	details := rack.GetDetails(ctx)
	assert.Nil(details)

	rack.SetDetails(ctx, stdDetails)

	details = rack.GetDetails(ctx)
	require.NotNil(details)
	assert.Equal(stdDetails, details)

	rack.SetDetails(ctx, nil)

	details = rack.GetDetails(ctx)
	require.Nil(details)

	// This will actually attempt to read the rack from the store. Since
	// we have yet to create the rack, we expect to see a "not found"
	// type error.
	//
	rev, err = rack.Read(ctx)
	require.Error(err)
	assert.Equal(errors.ErrRackNotFound{Region: rack.Region, Zone: rack.Zone, Rack: rack.Rack}, err)
	assert.Equal(store.RevisionInvalid, rev)

	rev, err = rack.Update(ctx, false)
	require.Error(err)
	assert.Equal(errors.ErrDetailsNotAvailable("rack"), err)
	assert.Equal(store.RevisionInvalid, rev)

	rev, err = rack.Update(ctx, true)
	require.Error(err)
	assert.Equal(errors.ErrDetailsNotAvailable("rack"), err)
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
	// assert.Equal(errors.ErrfRackNotFound(rack.Region, rack.Zone, rack.Rack), err)
	// assert.Equal(store.RevisionInvalid, rev)

	// rev, err = rack.Delete(ctx, true)
	// require.Error(err)
	// assert.Equal(errors.ErrfRackNotFound(rack.Region, rack.Zone, rack.Rack), err)
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
	rev := pdu.GetRevision()
	assert.Equal(store.RevisionInvalid, rev)

	rev = pdu.GetRevisionRecord()
	assert.Equal(store.RevisionInvalid, rev)

	rev = pdu.GetRevisionStore()
	assert.Equal(store.RevisionInvalid, rev)

	// Now try the various combinations of setting and clearing
	// details and ports.
	//
	details := pdu.GetDetails(ctx)
	assert.Nil(details)

	ports := pdu.GetPorts(ctx)
	require.Nil(ports)


	// Now set just the details
	//
	pdu.SetDetails(ctx, stdDetails)

	details = pdu.GetDetails(ctx)
	require.NotNil(details)
	assert.Equal(stdDetails, details)

	ports = pdu.GetPorts(ctx)
	require.Nil(ports)


	// Also set the ports
	//
	pdu.SetPorts(ctx, stdPorts)

	details = pdu.GetDetails(ctx)
	require.NotNil(details)
	assert.Equal(stdDetails, details)

	ports = pdu.GetPorts(ctx)
	require.NotNil(ports)
	assert.Equal(stdPorts, ports)


	// Clear just the details
	//
	pdu.SetDetails(ctx, nil)

	details = pdu.GetDetails(ctx)
	assert.Nil(details)

	ports = pdu.GetPorts(ctx)
	require.NotNil(ports)
	assert.Equal(stdPorts, ports)


	// Now also clear the ports
	//
	pdu.SetPorts(ctx, nil)

	details = pdu.GetDetails(ctx)
	assert.Nil(details)

	ports = pdu.GetPorts(ctx)
	require.Nil(ports)


	// And then once agains, set both details and ports
	//
	pdu.SetDetails(ctx, stdDetails)
	pdu.SetPorts(ctx, stdPorts)

	details = pdu.GetDetails(ctx)
	require.NotNil(details)
	assert.Equal(stdDetails, details)

	ports = pdu.GetPorts(ctx)
	require.NotNil(ports)
	assert.Equal(stdPorts, ports)

	// This will actually attempt to read the pdu from the store. Since
	// we have yet to create the pdu, we expect to see a "not found"
	// type error.
	//
	rev, err = pdu.Read(ctx)
	require.Error(err)
	assert.Equal(errors.ErrPduNotFound{Region: pdu.Region, Zone: pdu.Zone, Rack: pdu.Rack, Pdu: pdu.ID}, err)
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
	require.Equal(errors.ErrPortsNotAvailable("pdu"), err)
	assert.Equal(store.RevisionInvalid, rev)

	rev, err = pdu.Update(ctx, true)
	require.Equal(errors.ErrPortsNotAvailable("pdu"), err)
	assert.Equal(store.RevisionInvalid, rev)

	pdu.SetDetails(ctx, nil)

	rev, err = pdu.Update(ctx, false)
	require.Error(err)
	assert.Equal(errors.ErrDetailsNotAvailable("pdu"), err)
	assert.Equal(store.RevisionInvalid, rev)

	rev, err = pdu.Update(ctx, true)
	require.Error(err)
	assert.Equal(errors.ErrDetailsNotAvailable("pdu"), err)
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
	// assert.Equal(errors.ErrfPduNotFound(pdu.Region, pdu.Zone, pdu.Rack, pdu.ID), err)
	// assert.Equal(store.RevisionInvalid, rev)

	// rev, err = pdu.Delete(ctx, true)
	// require.Error(err)
	// assert.Equal(errors.ErrfPduNotFound(pdu.Region, pdu.Zone, pdu.Rack, pdu.ID), err)
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
	rev := tor.GetRevision()
	assert.Equal(store.RevisionInvalid, rev)

	rev = tor.GetRevisionRecord()
	assert.Equal(store.RevisionInvalid, rev)

	rev = tor.GetRevisionStore()
	assert.Equal(store.RevisionInvalid, rev)

	// Now try the various combinations of setting and clearing
	// details and ports.
	//
	details := tor.GetDetails(ctx)
	assert.Nil(details)

	ports := tor.GetPorts(ctx)
	require.Nil(ports)

	// Now set just the details
	//
	tor.SetDetails(ctx, stdDetails)

	details = tor.GetDetails(ctx)
	require.NotNil(details)
	assert.Equal(stdDetails, details)

	ports = tor.GetPorts(ctx)
	require.Nil(ports)

	// Also set the ports
	//
	tor.SetPorts(ctx, stdPorts)

	details = tor.GetDetails(ctx)
	require.NotNil(details)
	assert.Equal(stdDetails, details)

	ports = tor.GetPorts(ctx)
	require.NotNil(ports)
	assert.Equal(stdPorts, ports)

	// Clear just the details
	//
	tor.SetDetails(ctx, nil)

	details = tor.GetDetails(ctx)
	assert.Nil(details)

	ports = tor.GetPorts(ctx)
	require.NotNil(ports)
	assert.Equal(stdPorts, ports)

	// Now also clear the ports
	//
	tor.SetPorts(ctx, nil)

	details = tor.GetDetails(ctx)
	assert.Nil(details)

	ports = tor.GetPorts(ctx)
	require.Nil(ports)

	// And then once again, set both details and ports
	//
	tor.SetDetails(ctx, stdDetails)
	tor.SetPorts(ctx, stdPorts)

	details = tor.GetDetails(ctx)
	require.NotNil(details)
	assert.Equal(stdDetails, details)

	ports = tor.GetPorts(ctx)
	require.NotNil(ports)
	assert.Equal(stdPorts, ports)

	// This will actually attempt to read the tor from the store. Since
	// we have yet to create the zone, we expect to see a "not found"
	// type error.
	//
	rev, err = tor.Read(ctx)
	require.Error(err)
	assert.Equal(errors.ErrTorNotFound{Region: tor.Region, Zone: tor.Zone, Rack: tor.Rack, Tor: tor.ID}, err)
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
	require.Equal(errors.ErrPortsNotAvailable("tor"), err)
	assert.Equal(store.RevisionInvalid, rev)

	rev, err = tor.Update(ctx, true)
	require.Equal(errors.ErrPortsNotAvailable("tor"), err)
	assert.Equal(store.RevisionInvalid, rev)

	tor.SetDetails(ctx, nil)

	rev, err = tor.Update(ctx, false)
	require.Error(err)
	assert.Equal(errors.ErrDetailsNotAvailable("tor"), err)
	assert.Equal(store.RevisionInvalid, rev)

	rev, err = tor.Update(ctx, true)
	require.Error(err)
	assert.Equal(errors.ErrDetailsNotAvailable("tor"), err)
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
	// assert.Equal(errors.ErrfTorNotFound(tor.Region, tor.Zone, tor.Rack, tor.ID), err)
	// assert.Equal(store.RevisionInvalid, rev)

	// rev, err = tor.Delete(ctx, true)
	// require.Error(err)
	// assert.Equal(errors.ErrfTorNotFound(tor.Region, tor.Zone, tor.Rack, tor.ID), err)
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
	rev := blade.GetRevision()
	assert.Equal(store.RevisionInvalid, rev)

	rev = blade.GetRevisionRecord()
	assert.Equal(store.RevisionInvalid, rev)

	rev = blade.GetRevisionStore()
	assert.Equal(store.RevisionInvalid, rev)

	// Now try the various combinations of setting and clearing
	// details and ports.
	//
	details := blade.GetDetails(ctx)
	require.Nil(details)

	capacity := blade.GetCapacity(ctx)
	require.Nil(capacity)

	bootOnPowerOn, bootInfo := blade.GetBootInfo(ctx)
	assert.False(bootOnPowerOn)
	require.Nil(bootInfo)

	// Now set just the details
	//
	blade.SetDetails(ctx, stdDetails)

	details = blade.GetDetails(ctx)
	require.NotNil(details)
	assert.Equal(stdDetails, details)

	capacity = blade.GetCapacity(ctx)
	require.Nil(capacity)

	bootOnPowerOn, bootInfo = blade.GetBootInfo(ctx)
	require.Nil(bootInfo)
	assert.False(bootOnPowerOn)

	// Then set the capacity
	//
	blade.SetCapacity(ctx, stdCapacity)

	details = blade.GetDetails(ctx)
	require.NotNil(details)
	assert.Equal(stdDetails, details)

	capacity = blade.GetCapacity(ctx)
	require.NotNil(capacity)
	assert.Equal(stdCapacity, capacity)

	bootOnPowerOn, bootInfo = blade.GetBootInfo(ctx)
	require.Nil(bootInfo)
	assert.False(bootOnPowerOn)

	// Finally, also set the bootInfo
	//
	blade.SetBootInfo(ctx, stdBootOnPowerOn, stdBootInfo)

	details = blade.GetDetails(ctx)
	require.NotNil(details)
	assert.Equal(stdDetails, details)

	capacity = blade.GetCapacity(ctx)
	require.NotNil(capacity)
	assert.Equal(stdCapacity, capacity)

	bootOnPowerOn, bootInfo = blade.GetBootInfo(ctx)
	require.NotNil(bootInfo)
	assert.Equal(stdBootOnPowerOn, bootOnPowerOn)
	assert.Equal(stdBootInfo, bootInfo)

	// This will actually attempt to read the blade from the store. Since
	// we have yet to create the blade, we expect to see a "not found"
	// type error.
	//
	rev, err = blade.Read(ctx)
	require.Error(err)
	assert.Equal(errors.ErrBladeNotFound{Region: blade.Region, Zone: blade.Zone, Rack: blade.Rack, Blade: blade.ID}, err)
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
	require.Equal(errors.ErrBootInfoNotAvailable("blade"), err)
	assert.Equal(store.RevisionInvalid, rev)

	rev, err = blade.Update(ctx, true)
	require.Equal(errors.ErrBootInfoNotAvailable("blade"), err)
	assert.Equal(store.RevisionInvalid, rev)

	blade.SetCapacity(ctx, nil)

	rev, err = blade.Update(ctx, false)
	require.Equal(errors.ErrCapacityNotAvailable("blade"), err)
	assert.Equal(store.RevisionInvalid, rev)

	rev, err = blade.Update(ctx, true)
	require.Equal(errors.ErrCapacityNotAvailable("blade"), err)
	assert.Equal(store.RevisionInvalid, rev)

	blade.SetDetails(ctx, nil)

	rev, err = blade.Update(ctx, false)
	require.Error(err)
	assert.Equal(errors.ErrDetailsNotAvailable("blade"), err)
	assert.Equal(store.RevisionInvalid, rev)

	rev, err = blade.Update(ctx, true)
	require.Error(err)
	assert.Equal(errors.ErrDetailsNotAvailable("blade"), err)
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
	// assert.Equal(errors.ErrfBladeNotFound(blade.Region, blade.Zone, blade.Rack, blade.ID), err)
	// assert.Equal(store.RevisionInvalid, rev)

	// rev, err = blade.Delete(ctx, true)
	// require.Error(err)
	// assert.Equal(errors.ErrfBladeNotFound(blade.Region, blade.Zone, blade.Rack, blade.ID), err)
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

	rev2 := r.GetRevision()
	assert.Equal(rev, rev2)

	// Read the region back using the direct constructor
	//
	rRead, err := NewRegion(ctx, ts.store, DefinitionTable, regionName)
	require.NoError(err)

	revRead, err := rRead.Read(ctx)
	require.NoError(err)
	assert.Equal(rev, revRead)
	assert.Equal(revRead, rRead.GetRevision())

	detRead := rRead.GetDetails(ctx)
	require.NoError(err)
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
	assert.Equal(revRead, cr.GetRevision())

	crDet := cr.GetDetails(ctx)
	require.NoError(err)
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

	rev2 := z.GetRevision()
	assert.Equal(rev, rev2)


	// Read the zone back using the direct constructor
	//
	rRead, err := NewZone(ctx, ts.store, DefinitionTable, regionName, zoneName)
	require.NoError(err)

	revRead, err := rRead.Read(ctx)
	require.NoError(err)
	assert.Equal(rev, revRead)
	assert.Equal(revRead, rRead.GetRevision())

	detRead := rRead.GetDetails(ctx)
	require.NoError(err)
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
	assert.Equal(zoneRev, zone.GetRevision())

	zoneDet := zone.GetDetails(ctx)
	require.NoError(err)
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

	rev2 := r.GetRevision()
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
	assert.Equal(revRead, rRead.GetRevision())

	detRead := rRead.GetDetails(ctx)
	require.NoError(err)
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
	assert.Equal(rackRev, rack.GetRevision())

	rackDet := rack.GetDetails(ctx)
	require.NoError(err)
	assert.Equal(stdDetails, rackDet)
}

func (ts *testSuiteCore) TestPduReadDetails() {
	assert := ts.Assert()
	require := ts.Require()

	stdSuffix := "TestPduReadDetails"

	regionName := ts.regionName(stdSuffix)
	zoneName   := ts.zoneName(stdSuffix)
	rackName   := ts.rackName(stdSuffix)
	pduID      := int64(2)

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

	rev2 := p.GetRevision()
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
	assert.Equal(p2Rev, p2.GetRevision())

	p2Det := p2.GetDetails(ctx)
	require.NoError(err)
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
	assert.Equal(pduRev, pdu.GetRevision())

	pduDet := pdu.GetDetails(ctx)
	require.NoError(err)
	assert.Equal(stdDetails, pduDet)

	pduPorts := pdu.GetPorts(ctx)
	require.NoError(err)
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

	rev2 := t.GetRevision()
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
	assert.Equal(t2Rev, t2.GetRevision())

	p2Det := t2.GetDetails(ctx)
	require.NoError(err)
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
	assert.Equal(torRev, tor.GetRevision())

	pduDet := tor.GetDetails(ctx)
	require.NoError(err)
	assert.Equal(stdDetails, pduDet)

	torPorts := tor.GetPorts(ctx)
	require.NoError(err)
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

	rev2 := b.GetRevision()
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
	assert.Equal(t2Rev, b2.GetRevision())

	p2Det := b2.GetDetails(ctx)
	require.NoError(err)
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
	assert.Equal(bladeRev, blade.GetRevision())

	bladeDet := blade.GetDetails(ctx)
	require.NoError(err)
	assert.Equal(stdDetails, bladeDet)

	bladeCapacity := blade.GetCapacity(ctx)
	require.NoError(err)
	assert.Equal(stdCapacity, bladeCapacity)

	bladeBootOnPowerOn, bladeBootInfo := blade.GetBootInfo(ctx)
	require.NoError(err)
	assert.Equal(stdBootOnPowerOn, bladeBootOnPowerOn)
	assert.Equal(stdBootInfo, bladeBootInfo)
}

func (ts *testSuiteCore) TestRegionUpdateDetails() {

	assert := ts.Assert()
	require := ts.Require()

	stdSuffix := "TestRegionUpdateDetails"

	regionName := ts.regionName(stdSuffix)

	stdDetails := ts.stdRegionDetails(stdSuffix)

	ctx := context.Background()

	r, err := NewRegion(ctx, ts.store, DefinitionTable, regionName)
	require.NoError(err)

	r.SetDetails(ctx, stdDetails)

	revCreate, err := r.Create(ctx)
	require.NoError(err)
	assert.NotEqual(store.RevisionInvalid, revCreate)

	// Read the region back using the relative constructor
	//
	root, err := NewRoot (ctx, ts.store, DefinitionTable)
	require.NoError(err)

	cr, err := root.NewChild(ctx, regionName)
	require.NoError(err)

	revRead, err := cr.Read(ctx)
	require.NoError(err)
	assert.Equal(revCreate, revRead)

	details := cr.GetDetails(ctx)
	require.NotNil(details)

	details.State = pb.State_out_of_service
	details.Notes += " (out of service)"

	// Update the region using the direct constructor, conditional on revision
	//
	r.SetDetails(ctx, details)

	revUpdate, err := r.Update(ctx, false)
	require.NoError(err)
	assert.Less(revRead, revUpdate)

	// Verify update using relative constructor
	//
	revVerify, err := cr.Read(ctx)
	require.NoError(err)
	assert.Equal(revUpdate, revVerify)

	detailsVerify := cr.GetDetails(ctx)
	require.NotNil(detailsVerify)

	// Compare new details with original + deltas
	//
	assert.Equal(details, detailsVerify)
}

func (ts *testSuiteCore) TestZoneUpdateDetails() {

	assert := ts.Assert()
	require := ts.Require()

	stdSuffix := "TestZoneUpdateDetails"

	regionName := ts.regionName(stdSuffix)
	zoneName   := ts.zoneName(stdSuffix)

	stdDetails := ts.stdZoneDetails(stdSuffix)

	ctx := context.Background()

	z, err := NewZone(ctx, ts.store, DefinitionTable, regionName, zoneName)
	require.NoError(err)

	z.SetDetails(ctx, stdDetails)

	revCreate, err := z.Create(ctx)
	require.NoError(err)
	assert.NotEqual(store.RevisionInvalid, revCreate)

	// Read the zone back using the relative constructor
	//
	root, err := NewRoot (ctx, ts.store, DefinitionTable)
	require.NoError(err)

	region, err := root.NewChild(ctx, regionName)
	require.NoError(err)

	cz, err := region.NewChild(ctx, zoneName)
	require.NoError(err)

	revRead, err := cz.Read(ctx)
	require.NoError(err)
	assert.Equal(revCreate, revRead)

	details := cz.GetDetails(ctx)
	require.NotNil(details)

	details.State = pb.State_out_of_service
	details.Notes += " (out of service)"

	// Update the record using the direct constructor, conditional on revision
	//
	z.SetDetails(ctx, details)

	revUpdate, err := z.Update(ctx, false)
	require.NoError(err)
	assert.Less(revRead, revUpdate)

	// Verify update using relative constructor
	//
	revVerify, err := cz.Read(ctx)
	require.NoError(err)
	assert.Equal(revUpdate, revVerify)

	detailsVerify := cz.GetDetails(ctx)
	require.NotNil(detailsVerify)

	// Compare new details with original + deltas
	//
	assert.Equal(details, detailsVerify)
}

func (ts *testSuiteCore) TestRackUpdateDetails() {

	assert := ts.Assert()
	require := ts.Require()

	stdSuffix := "TestRackUpdateDetails"

	regionName := ts.regionName(stdSuffix)
	zoneName   := ts.zoneName(stdSuffix)
	rackName   := ts.rackName(stdSuffix)

	stdDetails := ts.stdRackDetails(stdSuffix)

	ctx := context.Background()

	r, err := NewRack(ctx, ts.store, DefinitionTable, regionName, zoneName, rackName)
	require.NoError(err)

	r.SetDetails(ctx, stdDetails)

	revCreate, err := r.Create(ctx)
	require.NoError(err)
	assert.NotEqual(store.RevisionInvalid, revCreate)

	// Read the rack back using the relative constructor
	//
	root, err := NewRoot (ctx, ts.store, DefinitionTable)
	require.NoError(err)

	region, err := root.NewChild(ctx, regionName)
	require.NoError(err)

	zone, err := region.NewChild(ctx, zoneName)
	require.NoError(err)

	cr, err := zone.NewChild(ctx, rackName)
	require.NoError(err)

	revRead, err := cr.Read(ctx)
	require.NoError(err)
	assert.Equal(revCreate, revRead)

	details := cr.GetDetails(ctx)
	require.NotNil(details)

	details.Enabled   = false
	details.Condition = pb.Condition_not_in_service
	details.Notes += " (out of service)"

	// Update the record using the direct constructor, conditional on revision
	//
	r.SetDetails(ctx, details)

	revUpdate, err := r.Update(ctx, false)
	require.NoError(err)
	assert.Less(revRead, revUpdate)

	// Verify update using relative constructor
	//
	revVerify, err := cr.Read(ctx)
	require.NoError(err)
	assert.Equal(revUpdate, revVerify)

	detailsVerify := cr.GetDetails(ctx)
	require.NotNil(detailsVerify)

	// Compare new details with original + deltas
	//
	assert.Equal(details, detailsVerify)
}

func (ts *testSuiteCore) TestPduUpdateDetails() {

	assert := ts.Assert()
	require := ts.Require()

	stdSuffix := "TestPduUpdateDetails"

	regionName := ts.regionName(stdSuffix)
	zoneName   := ts.zoneName(stdSuffix)
	rackName   := ts.rackName(stdSuffix)
	pduID      := ts.pduID(1)

	stdDetails := ts.stdPduDetails(pduID)
	stdPorts   := ts.stdPowerPorts(4)

	ctx := context.Background()

	p, err := NewPdu(ctx, ts.store, DefinitionTable, regionName, zoneName, rackName, pduID)
	require.NoError(err)

	p.SetDetails(ctx, stdDetails)
	p.SetPorts(ctx, stdPorts)

	revCreate, err := p.Create(ctx)
	require.NoError(err)
	assert.NotEqual(store.RevisionInvalid, revCreate)

	// Read the rack back using the relative constructor
	//
	root, err := NewRoot (ctx, ts.store, DefinitionTable)
	require.NoError(err)

	region, err := root.NewChild(ctx, regionName)
	require.NoError(err)

	zone, err := region.NewChild(ctx, zoneName)
	require.NoError(err)

	rack, err := zone.NewChild(ctx, rackName)
	require.NoError(err)

	cp, err := rack.NewPdu(ctx, pduID)
	require.NoError(err)

	revRead, err := cp.Read(ctx)
	require.NoError(err)
	assert.Equal(revCreate, revRead)

	details := cp.GetDetails(ctx)
	require.NotNil(details)

	details.Enabled   = false
	details.Condition = pb.Condition_not_in_service

	// Update the record using the direct constructor, conditional on revision
	//
	p.SetDetails(ctx, details)

	revUpdate, err := p.Update(ctx, false)
	require.NoError(err)
	assert.Less(revRead, revUpdate)

	// Verify update using relative constructor
	//
	revVerify, err := cp.Read(ctx)
	require.NoError(err)
	assert.Equal(revUpdate, revVerify)

	detailsVerify := cp.GetDetails(ctx)
	require.NotNil(detailsVerify)

	portsVerify := cp.GetPorts(ctx)
	require.NotNil(portsVerify)

	// Compare new details with original + deltas
	//
	assert.Equal(details, detailsVerify)
	assert.Equal(stdPorts, portsVerify)
}

func (ts *testSuiteCore) TestTorUpdateDetails() {

	assert := ts.Assert()
	require := ts.Require()

	stdSuffix := "TestTorUpdateDetails"

	regionName := ts.regionName(stdSuffix)
	zoneName   := ts.zoneName(stdSuffix)
	rackName   := ts.rackName(stdSuffix)
	torID      := ts.torID(1)

	stdDetails := ts.stdTorDetails()
	stdPorts   := ts.stdNetworkPorts(4)

	ctx := context.Background()

	t, err := NewTor(ctx, ts.store, DefinitionTable, regionName, zoneName, rackName, torID)
	require.NoError(err)

	t.SetDetails(ctx, stdDetails)
	t.SetPorts(ctx, stdPorts)

	revCreate, err := t.Create(ctx)
	require.NoError(err)
	assert.NotEqual(store.RevisionInvalid, revCreate)

	// Read the rack back using the relative constructor
	//
	root, err := NewRoot (ctx, ts.store, DefinitionTable)
	require.NoError(err)

	region, err := root.NewChild(ctx, regionName)
	require.NoError(err)

	zone, err := region.NewChild(ctx, zoneName)
	require.NoError(err)

	rack, err := zone.NewChild(ctx, rackName)
	require.NoError(err)

	ct, err := rack.NewTor(ctx, torID)
	require.NoError(err)

	revRead, err := ct.Read(ctx)
	require.NoError(err)
	assert.Equal(revCreate, revRead)

	details := ct.GetDetails(ctx)
	require.NotNil(details)

	details.Enabled   = false
	details.Condition = pb.Condition_not_in_service

	// Update the record using the direct constructor, conditional on revision
	//
	t.SetDetails(ctx, details)

	revUpdate, err := t.Update(ctx, false)
	require.NoError(err)
	assert.Less(revRead, revUpdate)

	// Verify update using relative constructor
	//
	revVerify, err := ct.Read(ctx)
	require.NoError(err)
	assert.Equal(revUpdate, revVerify)

	detailsVerify := ct.GetDetails(ctx)
	require.NotNil(detailsVerify)

	portsVerify := ct.GetPorts(ctx)
	require.NotNil(portsVerify)

	// Compare new details with original + deltas
	//
	assert.Equal(details, detailsVerify)
	assert.Equal(stdPorts, portsVerify)
}

func (ts *testSuiteCore) TestBladeUpdateDetails() {

	assert := ts.Assert()
	require := ts.Require()

	stdSuffix := "TestBladeUpdateDetails"

	regionName       := ts.regionName(stdSuffix)
	zoneName         := ts.zoneName(stdSuffix)
	rackName         := ts.rackName(stdSuffix)
	bladeID          := ts.bladeID(1)

	stdDetails       := ts.stdBladeDetails()
	stdCapacity      := ts.stdBladeCapacity()
	stdBootInfo      := ts.stdBladeBootInfo()
	stdBootOnPowerOn := true

	ctx := context.Background()

	b, err := NewBlade(ctx, ts.store, DefinitionTable, regionName, zoneName, rackName, bladeID)
	require.NoError(err)

	b.SetDetails(ctx, stdDetails)
	b.SetCapacity(ctx, stdCapacity)
	b.SetBootInfo(ctx, stdBootOnPowerOn, stdBootInfo)

	revCreate, err := b.Create(ctx)
	require.NoError(err)
	assert.NotEqual(store.RevisionInvalid, revCreate)

	// Read the rack back using the relative constructor
	//
	root, err := NewRoot (ctx, ts.store, DefinitionTable)
	require.NoError(err)

	region, err := root.NewChild(ctx, regionName)
	require.NoError(err)

	zone, err := region.NewChild(ctx, zoneName)
	require.NoError(err)

	rack, err := zone.NewChild(ctx, rackName)
	require.NoError(err)

	cb, err := rack.NewBlade(ctx, bladeID)
	require.NoError(err)

	revRead, err := cb.Read(ctx)
	require.NoError(err)
	assert.Equal(revCreate, revRead)

	details := cb.GetDetails(ctx)
	require.NotNil(details)

	details.Enabled   = false
	details.Condition = pb.Condition_not_in_service

	// Update the record using the direct constructor, conditional on revision
	//
	b.SetDetails(ctx, details)

	revUpdate, err := b.Update(ctx, false)
	require.NoError(err)
	assert.Less(revRead, revUpdate)

	// Verify update using relative constructor
	//
	revVerify, err := cb.Read(ctx)
	require.NoError(err)
	assert.Equal(revUpdate, revVerify)

	detailsVerify := cb.GetDetails(ctx)
	require.NotNil(detailsVerify)

	capacityVerify := cb.GetCapacity(ctx)
	require.NotNil(detailsVerify)

	bootOnPowerOnVerify, bootInfoVerify := cb.GetBootInfo(ctx)
	require.NotNil(detailsVerify)

	// Compare new details with original + deltas
	//
	assert.Equal(details, detailsVerify)
	assert.Equal(stdCapacity, capacityVerify)
	assert.Equal(stdBootInfo, bootInfoVerify)
	assert.Equal(stdBootOnPowerOn, bootOnPowerOnVerify)
}

func (ts *testSuiteCore) TestRootListChildren() {
	assert := ts.Assert()
	require := ts.Require()

	ctx := context.Background()

	root, err := NewRoot (ctx, ts.store, DefinitionTableStdTest)
	require.NoError(err)

	_, regions, err := root.ListChildren(ctx)
	require.NoError(err)
	assert.Equal(ts.regionCount, len(regions))
}

func (ts *testSuiteCore) TestRegionListChildren() {
	assert := ts.Assert()
	require := ts.Require()

	ctx := context.Background()

	root, err := NewRoot (ctx, ts.store, DefinitionTableStdTest)
	require.NoError(err)

	_, regions, err := root.ListChildren(ctx)
	require.NoError(err)
	assert.Equal(ts.regionCount, len(regions))

	for _, v := range regions {
		region, err := root.NewChild(ctx, v)

		_, zones, err := region.ListChildren(ctx)
		require.NoError(err)
		assert.Equal(ts.zonesPerRegion, len(zones))
	}
}

func (ts *testSuiteCore) TestZoneListChildren() {
	assert := ts.Assert()
	require := ts.Require()

	ctx := context.Background()

	root, err := NewRoot (ctx, ts.store, DefinitionTableStdTest)
	require.NoError(err)

	_, regions, err := root.ListChildren(ctx)
	require.NoError(err)
	assert.Equal(ts.regionCount, len(regions))

	for _, v := range regions {
		region, err := root.NewChild(ctx, v)

		_, zones, err := region.ListChildren(ctx)
		require.NoError(err)
		assert.Equal(ts.zonesPerRegion, len(zones))

		for _, v := range zones {
			zone, err := region.NewChild(ctx, v)
	
			_, racks, err := zone.ListChildren(ctx)
			require.NoError(err)
			assert.Equal(ts.racksPerZone, len(racks))
		}
	}
}

func (ts *testSuiteCore) TestRackListChildren() {
	assert := ts.Assert()
	require := ts.Require()

	ctx := context.Background()

	root, err := NewRoot (ctx, ts.store, DefinitionTableStdTest)
	require.NoError(err)

	_, regions, err := root.ListChildren(ctx)
	require.NoError(err)
	assert.Equal(ts.regionCount, len(regions))

	for _, v := range regions {
		region, err := root.NewChild(ctx, v)

		_, zones, err := region.ListChildren(ctx)
		require.NoError(err)
		assert.Equal(ts.zonesPerRegion, len(zones))

		for _, v := range zones {
			zone, err := region.NewChild(ctx, v)
	
			_, racks, err := zone.ListChildren(ctx)
			require.NoError(err)
			assert.Equal(ts.racksPerZone, len(racks))

			for _, v := range racks {
				rack, err := zone.NewChild(ctx, v)

				_, pdus, err := rack.ListPdus(ctx)
				require.NoError(err)
				assert.Equal(ts.pdusPerRack, len(pdus))

				_, tors, err := rack.ListTors(ctx)
				require.NoError(err)
				assert.Equal(ts.torsPerRack, len(tors))

				_, blades, err := rack.ListBlades(ctx)
				require.NoError(err)
				assert.Equal(ts.bladesPerRack, len(blades))
			}
		}
	}
}

func (ts *testSuiteCore) TestRootFetchChildren() {
	assert := ts.Assert()
	require := ts.Require()

	ctx := context.Background()

	root, err := NewRoot (ctx, ts.store, DefinitionTableStdTest)
	require.NoError(err)

	_, regions, err := root.FetchChildren(ctx)
	require.NoError(err)
	assert.Equal(ts.regionCount, len(*regions))

	for n, v := range *regions {
		ts.verifyStandardInventoryRegionDetails(n, v.GetDetails(ctx))
	}
}

func (ts *testSuiteCore) TestRegionFetchChildren() {
	assert := ts.Assert()
	require := ts.Require()

	ctx := context.Background()

	root, err := NewRoot (ctx, ts.store, DefinitionTableStdTest)
	require.NoError(err)

	_, regions, err := root.FetchChildren(ctx)
	require.NoError(err)
	assert.Equal(ts.regionCount, len(*regions))

	for n, v := range *regions {
		ts.verifyStandardInventoryRegionDetails(n, v.GetDetails(ctx))

		region, err := root.NewChild(ctx, n)

		_, zones, err := region.FetchChildren(ctx)
		require.NoError(err)
		assert.Equal(ts.zonesPerRegion, len(*zones))

		for n, v := range *zones {
			ts.verifyStandardInventoryZoneDetails(n, v.GetDetails(ctx))
		}
	}
}

func (ts *testSuiteCore) TestZoneFetchChildren() {
	assert := ts.Assert()
	require := ts.Require()

	ctx := context.Background()

	root, err := NewRoot (ctx, ts.store, DefinitionTableStdTest)
	require.NoError(err)

	_, regions, err := root.FetchChildren(ctx)
	require.NoError(err)
	assert.Equal(ts.regionCount, len(*regions))

	for n, v := range *regions {
		ts.verifyStandardInventoryRegionDetails(n, v.GetDetails(ctx))

		region, err := root.NewChild(ctx, n)

		_, zones, err := region.FetchChildren(ctx)
		require.NoError(err)
		assert.Equal(ts.zonesPerRegion, len(*zones))

		for n, v := range *zones {
			ts.verifyStandardInventoryZoneDetails(n, v.GetDetails(ctx))

			zone, err := region.NewChild(ctx, n)

			_, racks, err := zone.FetchChildren(ctx)
			require.NoError(err)
			assert.Equal(ts.zonesPerRegion, len(*zones))
	
			for n, v := range *racks {
				ts.verifyStandardInventoryRackDetails(n, v.GetDetails(ctx))
			}
		}
	}
}

func (ts *testSuiteCore) TestRackFetchChildren() {
	assert := ts.Assert()
	require := ts.Require()

	ctx := context.Background()

	root, err := NewRoot (ctx, ts.store, DefinitionTableStdTest)
	require.NoError(err)

	_, regions, err := root.FetchChildren(ctx)
	require.NoError(err)
	assert.Equal(ts.regionCount, len(*regions))

	for n, v := range *regions {
		ts.verifyStandardInventoryRegionDetails(n, v.GetDetails(ctx))

		region, err := root.NewChild(ctx, n)

		_, zones, err := region.FetchChildren(ctx)
		require.NoError(err)
		assert.Equal(ts.zonesPerRegion, len(*zones))

		for n, v := range *zones {
			ts.verifyStandardInventoryZoneDetails(n, v.GetDetails(ctx))

			zone, err := region.NewChild(ctx, n)

			_, racks, err := zone.FetchChildren(ctx)
			require.NoError(err)
			assert.Equal(ts.zonesPerRegion, len(*zones))
	
			for n, v := range *racks {
				ts.verifyStandardInventoryRackDetails(n, v.GetDetails(ctx))

				rack, err := zone.NewChild(ctx, n)

				// There should be no implementation of the generic FetchChildren()
				// function for a rack. Instead it has specific FetchXxx() functions
				// for each type of sub-element.
				//
				rev, children, err := rack.FetchChildren(ctx)
				require.ErrorIs(err, errors.ErrFunctionNotAvailable)
				assert.Nil(children)
				assert.Equal(store.RevisionInvalid, rev)

				_, pdus, err := rack.FetchPdus(ctx)
				require.NoError(err)
				assert.Equal(ts.pdusPerRack, len(*pdus))

				_, tors, err := rack.FetchTors(ctx)
				require.NoError(err)
				assert.Equal(ts.torsPerRack, len(*tors))

				_, blades, err := rack.FetchBlades(ctx)
				require.NoError(err)
				assert.Equal(ts.bladesPerRack, len(*blades))

				for n, v := range *pdus {
					ts.verifyStandardInventoryPdu(n, v)
				}

				for n, v := range *tors {
					ts.verifyStandardInventoryTor(n, &v)
				}

				for n, v := range *blades {
					ts.verifyStandardInventoryBlade(n, &v)
				}
			}
		}
	}
}

func TestInventoryTestSuite(t *testing.T) {
	suite.Run(t, new(testSuiteCore))
}
