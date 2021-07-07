package inventory

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Jim3Things/CloudChamber/simulation/internal/clients/namespace"
	"github.com/Jim3Things/CloudChamber/simulation/internal/clients/store"
	"github.com/Jim3Things/CloudChamber/simulation/pkg/errors"
	pb "github.com/Jim3Things/CloudChamber/simulation/pkg/protos/inventory"
)

type definitionTestSuite struct {
	testSuiteCore

	inventory      *Inventory

	regionCount    int
	zonesPerRegion int
	racksPerZone   int
	pdusPerRack    int
	torsPerRack    int
	bladesPerRack  int

	portsPerPdu int
	portsPerTor int
}

func (ts *definitionTestSuite) rootName(suffix string)   string { return "StandardRoot-" + suffix }
func (ts *definitionTestSuite) regionName(suffix string) string { return "REG-PNW-"      + suffix }
func (ts *definitionTestSuite) zoneName(suffix string)   string { return "Zone-01-"      + suffix }
func (ts *definitionTestSuite) rackName(suffix string)   string { return "Rack-01-"      + suffix }
func (ts *definitionTestSuite) pduID(ID int64)      int64  { return int64(ID)}
func (ts *definitionTestSuite) torID(ID int64)      int64  { return int64(ID)}
func (ts *definitionTestSuite) bladeID(ID int64)    int64  { return int64(ID)}


func (ts *definitionTestSuite) stdRootDetails(suffix string) *pb.RootDetails {
	return &pb.RootDetails{
		Name:  ts.rootName(suffix),
		Notes: "root for inventory definition test",
	}
}

func (ts *definitionTestSuite) stdRegionDetails(suffix string) *pb.RegionDetails {
	return &pb.RegionDetails{
		Name:     ts.regionName(suffix),
		State:    pb.State_in_service,
		Location: "StdRegion: DC-PNW-" + suffix,
		Notes:    "region for inventory definition test",
	}
}

func (ts *definitionTestSuite) stdZoneDetails(suffix string) *pb.ZoneDetails {
	return &pb.ZoneDetails{
		Enabled:  true,
		State:    pb.State_in_service,
		Location: "StdZone: DC-PNW-" + suffix,
		Notes:    "zone for inventory definition test",
	}
}

func (ts *definitionTestSuite) stdRackDetails(suffix string) *pb.RackDetails {
	return &pb.RackDetails{
		Enabled:   true,
		Condition: pb.Condition_operational,
		Location:  "StdRack: DC-PNW-" + suffix,
		Notes:     "rack for inventory definition test",
	}
}

func (ts *definitionTestSuite) stdPduDetails(ID int64) *pb.PduDetails {
	return &pb.PduDetails{
		Enabled:   true,
		Condition: pb.Condition_operational,
	}
}

func (ts *definitionTestSuite) stdPowerPorts(count int) map[int64]*pb.PowerPort {
	ports := make(map[int64]*pb.PowerPort)

	for i := 0; i < count; i++ {
		ports[int64(i)] = &pb.PowerPort{
			Wired: true,
			Item: &pb.Hardware{
				Type: pb.Hardware_blade,
				Id:   int64(i),
				Port: int64(i),
			},
		}
	}

	return ports
}

func (ts *definitionTestSuite) stdTorDetails() *pb.TorDetails {
	return &pb.TorDetails{
		Enabled:   true,
		Condition: pb.Condition_operational,
	}
}

func (ts *definitionTestSuite) stdNetworkPorts(count int) map[int64]*pb.NetworkPort {
	ports := make(map[int64]*pb.NetworkPort)

	for i := 0; i < count; i++ {
		ports[int64(i)] = &pb.NetworkPort{
			Wired: true,
			Item: &pb.Hardware{
				Type: pb.Hardware_blade,
				Id:   int64(i),
				Port: int64(1),
			},
		}
	}

	return ports
}

func (ts *definitionTestSuite) stdBladeDetails() *pb.BladeDetails {
	return &pb.BladeDetails{
		Enabled:   true,
		Condition: pb.Condition_operational,
	}
}

func (ts *definitionTestSuite) stdBladeCapacity() *pb.BladeCapacity {
	return &pb.BladeCapacity{
		Cores:                  16,
		MemoryInMb:             1024,
		DiskInGb:               32,
		NetworkBandwidthInMbps: 1024,
		Arch:                   "amd64",
	}
}

func (ts *definitionTestSuite) stdBladeBootInfo() *pb.BladeBootInfo {
	return &pb.BladeBootInfo{
		Source:     pb.BladeBootInfo_local,
		Image:      "test-image.vhdx",
		Version:    "20201225-0000",
		Parameters: "-param1=val1 -param2=val2",
	}
}

func (ts *definitionTestSuite) stdBladeBootOnPowerOn() bool {
	return true
}

func (ts *definitionTestSuite)createStandardInventory(ctx context.Context) error {

	err := ts.createInventory(
		ctx,
		namespace.DefinitionTableStdTest,
		ts.regionCount,
		ts.zonesPerRegion,
		ts.racksPerZone,
		ts.pdusPerRack,
		ts.torsPerRack,
		ts.bladesPerRack)

	return err
}

func (ts *definitionTestSuite)verifyStandardInventoryRegionDetails(name string, details *pb.RegionDetails) {
	assert := ts.Assert()

	check := ts.stdRegionDetails(name)

	assert.Equal(check.Name, details.Name)
	assert.Equal(check.State, details.State)
	assert.Equal(check.Location, details.Location)
	assert.Equal(check.Notes, details.Notes)
}

func (ts *definitionTestSuite)verifyStandardInventoryZoneDetails(name string, details *pb.ZoneDetails) {
	assert := ts.Assert()

	check := ts.stdZoneDetails(name)

	assert.Equal(check.Enabled, details.Enabled)
	assert.Equal(check.State, details.State)
	assert.Equal(check.Location, details.Location)
	assert.Equal(check.Notes, details.Notes)
}

func (ts *definitionTestSuite)verifyStandardInventoryRackDetails(name string, details *pb.RackDetails) {
	assert := ts.Assert()

	check := ts.stdRackDetails(name)

	assert.Equal(check.Enabled, details.Enabled)
	assert.Equal(check.Condition, details.Condition)
	assert.Equal(check.Location, details.Location)
	assert.Equal(check.Notes, details.Notes)
}

func (ts *definitionTestSuite) verifyStandardInventoryPdu(index int64, pdu Pdu) {
	assert := ts.Assert()

	details := ts.stdPduDetails(index)

	assert.Equal(details.Enabled, pdu.details.Enabled)
	assert.Equal(details.Condition, pdu.details.Condition)

	ports := ts.stdPowerPorts(ts.portsPerPdu)

	assert.Equal(len(ports), len(pdu.ports))

	for k, v := range ports {
		p := (pdu.ports)[k]

		assert.Equal(v.Wired, p.Wired)
		assert.Equal(v.Item.Type, p.Item.Type)
		assert.Equal(v.Item.Id, p.Item.Id)
		assert.Equal(v.Item.Port, p.Item.Port)
	}
}

func (ts *definitionTestSuite) verifyStandardInventoryTor(index int64, tor *Tor) {
	assert := ts.Assert()

	details := ts.stdTorDetails()

	assert.Equal(details.Enabled, tor.details.Enabled)
	assert.Equal(details.Condition, tor.details.Condition)

	ports := ts.stdNetworkPorts(ts.portsPerTor)

	assert.Equal(len(ports), len(tor.ports))

	for k, v := range ports {
		p := (tor.ports)[k]

		assert.Equal(v.Wired, p.Wired)
		assert.Equal(v.Item.Type, p.Item.Type)
		assert.Equal(v.Item.Id, p.Item.Id)
		assert.Equal(v.Item.Port, p.Item.Port)
	}
}

func (ts *definitionTestSuite) verifyStandardInventoryBlade(index int64, blade *Blade) {
	assert := ts.Assert()

	details := ts.stdBladeDetails()

	assert.Equal(details.Enabled, blade.details.Enabled)
	assert.Equal(details.Condition, blade.details.Condition)

	capacity := ts.stdBladeCapacity()

	assert.Equal(capacity.Cores, blade.capacity.Cores)
	assert.Equal(capacity.MemoryInMb, blade.capacity.MemoryInMb)
	assert.Equal(capacity.DiskInGb, blade.capacity.DiskInGb)
	assert.Equal(capacity.NetworkBandwidthInMbps, blade.capacity.NetworkBandwidthInMbps)
	assert.Equal(capacity.Arch, blade.capacity.Arch)
	assert.Equal(capacity.Accelerators, blade.capacity.Accelerators)

	bootInfo := ts.stdBladeBootInfo()
	assert.Equal(bootInfo.Source, blade.bootInfo.Source)
	assert.Equal(bootInfo.Image, blade.bootInfo.Image)
	assert.Equal(bootInfo.Version, blade.bootInfo.Version)
	assert.Equal(bootInfo.Parameters, blade.bootInfo.Parameters)

	assert.True(blade.bootOnPowerOn)
}

func (ts *definitionTestSuite)createInventory(
	ctx context.Context,
	table namespace.TableName,
	regions int,
	zonesPerRegion int,
	racksPerZone int,
	pdusPerRack int,
	torsPerRack int,
	bladesPerRack int) error {

	root, err := ts.inventory.NewRoot(table)
	if err != nil {
		return err
	}

	for i := 1; i <= regions; i++ {
		regionName := fmt.Sprintf("Region-%d", i)

		region, err := root.NewChild(regionName)
		if err != nil {
			return err
		}


		region.SetDetails(ts.stdRegionDetails(regionName))

		_, err = region.Create(ctx, ViewDefinition)
		if err != nil {
			return err
		}

		for j := 1; j <= zonesPerRegion; j++ {
			zoneName := fmt.Sprintf("Zone-%d-%d", i, j)

			zone, err := region.NewChild(zoneName)
			if err != nil {
				return err
			}

				zone.SetDetails(ts.stdZoneDetails(zoneName))

			_, err = zone.Create(ctx, ViewDefinition)
			if err != nil {
				return err
			}

			for k := 1; k <= racksPerZone; k++ {
				rackName := fmt.Sprintf("Rack-%d-%d-%d", i, j, k)

				rack, err := zone.NewChild(rackName)
				if err != nil {
					return err
				}

				rack.SetDetails(ts.stdRackDetails(rackName))

				_, err = rack.Create(ctx, ViewDefinition)
				if err != nil {
					return err
				}

				for p := 0; p < pdusPerRack; p++ {
					pdu, err := rack.NewPdu(int64(p))
					if err != nil {
						return err
					}

					pdu.SetDetails(ts.stdPduDetails(1))
					pdu.SetPorts(ts.stdPowerPorts(ts.portsPerPdu))

					_, err = pdu.Create(ctx, ViewDefinition)
					if err != nil {
						return err
					}
				}

				for t := 0; t < torsPerRack; t++ {
					tor, err := rack.NewTor(int64(t))
					if err != nil {
						return err
					}

					tor.SetDetails(ts.stdTorDetails())
					tor.SetPorts(ts.stdNetworkPorts(ts.portsPerTor))

					_, err = tor.Create(ctx, ViewDefinition)
					if err != nil {
						return err
					}
				}

				for b := 0; b < bladesPerRack; b++ {
					blade, err := rack.NewBlade(int64(b))
					if err != nil {
						return err
					}

					blade.SetDetails(ts.stdBladeDetails())
					blade.SetCapacity(ts.stdBladeCapacity())
					blade.SetBootInfo(ts.stdBladeBootInfo())
					blade.SetBootPowerOn(true)

					_, err = blade.Create(ctx, ViewDefinition)
					if err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

func (ts *definitionTestSuite) SetupSuite() {
	require := ts.Require()

	ctx := context.Background()

	ts.testSuiteCore.SetupSuite()

	ts.inventory = NewInventory(ts.cfg, ts.store)

	// These values are relatively arbitrary. The only criteria is that different
	// constants were chosen to help separate different multiples of different
	// object types where possible and not to have values which are too large
	// to avoid lots of IO when setting up the test suite.
	//
	ts.regionCount = 2
	ts.zonesPerRegion = 3
	ts.racksPerZone = 4
	ts.pdusPerRack = 1
	ts.torsPerRack = 1
	ts.bladesPerRack = 5

	ts.portsPerPdu = ts.torsPerRack + ts.bladesPerRack
	ts.portsPerTor = ts.pdusPerRack + ts.torsPerRack + ts.bladesPerRack

	require.NoError(ts.utf.Open(ts.T()))
	require.NoError(ts.store.Connect())

	err := ts.createStandardInventory(ctx)
	require.NoError(err, "failed to create standard inventory")

	ts.store.Disconnect()
	ts.utf.Close()
}

func (ts *definitionTestSuite) TestNewRoot() {
	assert := ts.Assert()
	require := ts.Require()

	stdSuffix := "TestNewRoot"
	stdDetails := ts.stdRootDetails(stdSuffix)

	ctx := context.Background()

	root, err := ts.inventory.NewRoot(namespace.DefinitionTable)
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
	details := root.GetDetails()
	assert.Nil(details)

	root.SetDetails(stdDetails)

	details = root.GetDetails()
	require.NotNil(details)
	assert.Equal(stdDetails, details)

	root.SetDetails(nil)

	details = root.GetDetails()
	require.Nil(details)

	rev, err = root.Read(ctx, ViewDefinition)
	require.Error(err)
	assert.Equal(errors.ErrFunctionNotAvailable, err)
	assert.Equal(store.RevisionInvalid, rev)

	rev, err = root.Update(ctx, false, ViewDefinition)
	require.Error(err)
	assert.Equal(errors.ErrFunctionNotAvailable, err)
	assert.Equal(store.RevisionInvalid, rev)

	rev, err = root.Update(ctx, true, ViewDefinition)
	require.Error(err)
	assert.Equal(errors.ErrFunctionNotAvailable, err)
	assert.Equal(store.RevisionInvalid, rev)

	rev, err = root.Delete(ctx, false, ViewDefinition)
	require.Error(err)
	assert.Equal(errors.ErrFunctionNotAvailable, err)
	assert.Equal(store.RevisionInvalid, rev)

	rev, err = root.Delete(ctx, true, ViewDefinition)
	require.Error(err)
	assert.Equal(errors.ErrFunctionNotAvailable, err)
	assert.Equal(store.RevisionInvalid, rev)
}

func (ts *definitionTestSuite) TestNewRegion() {
	assert := ts.Assert()
	require := ts.Require()

	stdSuffix := "TestNewRegion"
	stdDetails := ts.stdRegionDetails(stdSuffix)

	ctx := context.Background()

	region, err := ts.inventory.NewRegion(namespace.DefinitionTable, ts.regionName(stdSuffix))
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
	details := region.GetDetails()
	assert.Nil(details)

	region.SetDetails(stdDetails)

	details = region.GetDetails()
	require.NotNil(details)
	assert.Equal(stdDetails, details)

	region.SetDetails(nil)

	details = region.GetDetails()
	require.Nil(details)

	// This will actually attempt to read the region from the store. Since
	// we have yet to create the region, we expect to see a "not found"
	// type error.
	//
	rev, err = region.Read(ctx, ViewDefinition)
	require.Error(err)
	assert.Equal(errors.ErrRegionNotFound{Region: region.Region}, err)
	assert.Equal(store.RevisionInvalid, rev)

	rev, err = region.Update(ctx, false, ViewDefinition)
	require.Error(err)
	assert.Equal(errors.ErrDetailsNotAvailable("region"), err)
	assert.Equal(store.RevisionInvalid, rev)

	rev, err = region.Update(ctx, true, ViewDefinition)
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

func (ts *definitionTestSuite) TestNewZone() {
	assert := ts.Assert()
	require := ts.Require()

	stdSuffix := "TestNewZone"
	stdDetails := ts.stdZoneDetails(stdSuffix)

	ctx := context.Background()

	zone, err := ts.inventory.NewZone(namespace.DefinitionTable, ts.regionName(stdSuffix), ts.zoneName(stdSuffix))
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
	details := zone.GetDetails()
	assert.Nil(details)

	zone.SetDetails(stdDetails)

	details = zone.GetDetails()
	require.NotNil(details)
	assert.Equal(stdDetails, details)

	zone.SetDetails(nil)

	details = zone.GetDetails()
	require.Nil(details)

	// This will actually attempt to read the zone from the store. Since
	// we have yet to create the zone, we expect to see a "not found"
	// type error.
	//
	rev, err = zone.Read(ctx, ViewDefinition)
	require.Error(err)
	assert.Equal(errors.ErrZoneNotFound{Region: zone.Region, Zone: zone.Zone}, err)
	assert.Equal(store.RevisionInvalid, rev)

	rev, err = zone.Update(ctx, false, ViewDefinition)
	require.Error(err)
	assert.Equal(errors.ErrDetailsNotAvailable("zone"), err)
	assert.Equal(store.RevisionInvalid, rev)

	rev, err = zone.Update(ctx, true, ViewDefinition)
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

func (ts *definitionTestSuite) TestNewRack() {
	assert := ts.Assert()
	require := ts.Require()

	stdSuffix := "TestNewRack"
	stdDetails := ts.stdRackDetails(stdSuffix)

	ctx := context.Background()

	rack, err := ts.inventory.NewRack(
		namespace.DefinitionTable,
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
	details := rack.GetDetails()
	assert.Nil(details)

	rack.SetDetails(stdDetails)

	details = rack.GetDetails()
	require.NotNil(details)
	assert.Equal(stdDetails, details)

	rack.SetDetails(nil)

	details = rack.GetDetails()
	require.Nil(details)

	// This will actually attempt to read the rack from the store. Since
	// we have yet to create the rack, we expect to see a "not found"
	// type error.
	//
	rev, err = rack.Read(ctx, ViewDefinition)
	require.Error(err)
	assert.Equal(errors.ErrRackNotFound{Region: rack.Region, Zone: rack.Zone, Rack: rack.Rack}, err)
	assert.Equal(store.RevisionInvalid, rev)

	rev, err = rack.Update(ctx, false, ViewDefinition)
	require.Error(err)
	assert.Equal(errors.ErrDetailsNotAvailable("rack"), err)
	assert.Equal(store.RevisionInvalid, rev)

	rev, err = rack.Update(ctx, true, ViewDefinition)
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

func (ts *definitionTestSuite) TestNewPdu() {
	assert := ts.Assert()
	require := ts.Require()

	stdSuffix := "TestNewPdu"
	stdDetails := ts.stdPduDetails(1)
	stdPorts := ts.stdPowerPorts(8)

	ctx := context.Background()

	pdu, err := ts.inventory.NewPdu(
		namespace.DefinitionTable,
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
	details := pdu.GetDetails()
	assert.Nil(details)

	ports := pdu.GetPorts()
	require.Nil(ports)

	// Now set just the details
	//
	pdu.SetDetails(stdDetails)

	details = pdu.GetDetails()
	require.NotNil(details)
	assert.Equal(stdDetails, details)

	ports = pdu.GetPorts()
	require.Nil(ports)

	// Also set the ports
	//
	pdu.SetPorts(stdPorts)

	details = pdu.GetDetails()
	require.NotNil(details)
	assert.Equal(stdDetails, details)

	ports = pdu.GetPorts()
	require.NotNil(ports)
	assert.Equal(stdPorts, ports)

	// Clear just the details
	//
	pdu.SetDetails(nil)

	details = pdu.GetDetails()
	assert.Nil(details)

	ports = pdu.GetPorts()
	require.NotNil(ports)
	assert.Equal(stdPorts, ports)

	// Now also clear the ports
	//
	pdu.SetPorts(nil)

	details = pdu.GetDetails()
	assert.Nil(details)

	ports = pdu.GetPorts()
	require.Nil(ports)

	// And then once again, set both details and ports
	//
	pdu.SetDetails(stdDetails)
	pdu.SetPorts(stdPorts)

	details = pdu.GetDetails()
	require.NotNil(details)
	assert.Equal(stdDetails, details)

	ports = pdu.GetPorts()
	require.NotNil(ports)
	assert.Equal(stdPorts, ports)

	// This will actually attempt to read the pdu from the store. Since
	// we have yet to create the pdu, we expect to see a "not found"
	// type error.
	//
	rev, err = pdu.Read(ctx, ViewDefinition)
	require.Error(err)
	assert.Equal(errors.ErrPduNotFound{Region: pdu.Region, Zone: pdu.Zone, Rack: pdu.Rack, Pdu: pdu.ID}, err)
	assert.Equal(store.RevisionInvalid, rev)

	// Clear the ports and check the update fails
	//
	// Note: the ordering of this and the subsequent statements assume
	//       the Update() call checks for the details being present before
	//       checking for the ports being present. Any change in the
	//       ordering may result in the tests needing amendment.
	//
	pdu.SetPorts(nil)

	rev, err = pdu.Update(ctx, false, ViewDefinition)
	require.Equal(errors.ErrPortsNotAvailable("pdu"), err)
	assert.Equal(store.RevisionInvalid, rev)

	rev, err = pdu.Update(ctx, true, ViewDefinition)
	require.Equal(errors.ErrPortsNotAvailable("pdu"), err)
	assert.Equal(store.RevisionInvalid, rev)

	pdu.SetDetails(nil)

	rev, err = pdu.Update(ctx, false, ViewDefinition)
	require.Error(err)
	assert.Equal(errors.ErrDetailsNotAvailable("pdu"), err)
	assert.Equal(store.RevisionInvalid, rev)

	rev, err = pdu.Update(ctx, true, ViewDefinition)
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

func (ts *definitionTestSuite) TestNewTor() {
	assert := ts.Assert()
	require := ts.Require()

	stdSuffix := "TestNewTor"
	stdDetails := ts.stdTorDetails()
	stdPorts := ts.stdNetworkPorts(8)

	ctx := context.Background()

	tor, err := ts.inventory.NewTor(
		namespace.DefinitionTable,
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
	details := tor.GetDetails()
	assert.Nil(details)

	ports := tor.GetPorts()
	require.Nil(ports)

	// Now set just the details
	//
	tor.SetDetails(stdDetails)

	details = tor.GetDetails()
	require.NotNil(details)
	assert.Equal(stdDetails, details)

	ports = tor.GetPorts()
	require.Nil(ports)

	// Also set the ports
	//
	tor.SetPorts(stdPorts)

	details = tor.GetDetails()
	require.NotNil(details)
	assert.Equal(stdDetails, details)

	ports = tor.GetPorts()
	require.NotNil(ports)
	assert.Equal(stdPorts, ports)

	// Clear just the details
	//
	tor.SetDetails(nil)

	details = tor.GetDetails()
	assert.Nil(details)

	ports = tor.GetPorts()
	require.NotNil(ports)
	assert.Equal(stdPorts, ports)

	// Now also clear the ports
	//
	tor.SetPorts(nil)

	details = tor.GetDetails()
	assert.Nil(details)

	ports = tor.GetPorts()
	require.Nil(ports)

	// And then once again, set both details and ports
	//
	tor.SetDetails(stdDetails)
	tor.SetPorts(stdPorts)

	details = tor.GetDetails()
	require.NotNil(details)
	assert.Equal(stdDetails, details)

	ports = tor.GetPorts()
	require.NotNil(ports)
	assert.Equal(stdPorts, ports)

	// This will actually attempt to read the tor from the store. Since
	// we have yet to create the zone, we expect to see a "not found"
	// type error.
	//
	rev, err = tor.Read(ctx, ViewDefinition)
	require.Error(err)
	assert.Equal(errors.ErrTorNotFound{Region: tor.Region, Zone: tor.Zone, Rack: tor.Rack, Tor: tor.ID}, err)
	assert.Equal(store.RevisionInvalid, rev)

	// Clear the ports and check the update fails
	//
	// Note: the ordering of this and the subsequent statements assume
	//       the Update() call checks for the details being present before
	//       checking for the ports being present. Any change in the
	//       ordering may result in the tests needing amendment.
	//
	tor.SetPorts(nil)

	rev, err = tor.Update(ctx, false, ViewDefinition)
	require.Equal(errors.ErrPortsNotAvailable("tor"), err)
	assert.Equal(store.RevisionInvalid, rev)

	rev, err = tor.Update(ctx, true, ViewDefinition)
	require.Equal(errors.ErrPortsNotAvailable("tor"), err)
	assert.Equal(store.RevisionInvalid, rev)

	tor.SetDetails(nil)

	rev, err = tor.Update(ctx, false, ViewDefinition)
	require.Error(err)
	assert.Equal(errors.ErrDetailsNotAvailable("tor"), err)
	assert.Equal(store.RevisionInvalid, rev)

	rev, err = tor.Update(ctx, true, ViewDefinition)
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

func (ts *definitionTestSuite) TestNewBlade() {
	assert := ts.Assert()
	require := ts.Require()

	stdSuffix := "TestNewBlade"
	stdDetails := ts.stdBladeDetails()
	stdCapacity := ts.stdBladeCapacity()
	stdBootInfo := ts.stdBladeBootInfo()

	stdBootOnPowerOn := true

	ctx := context.Background()

	blade, err := ts.inventory.NewBlade(
		namespace.DefinitionTable,
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
	details := blade.GetDetails()
	require.Nil(details)

	capacity := blade.GetCapacity()
	require.Nil(capacity)

	bootInfo := blade.GetBootInfo()
	require.Nil(bootInfo)

	bootOnPowerOn := blade.GetBootOnPowerOn()
	assert.False(bootOnPowerOn)

	// Now set just the details
	//
	blade.SetDetails(stdDetails)

	details = blade.GetDetails()
	require.NotNil(details)
	assert.Equal(stdDetails, details)

	capacity = blade.GetCapacity()
	require.Nil(capacity)

	bootInfo = blade.GetBootInfo()
	require.Nil(bootInfo)

	bootOnPowerOn = blade.GetBootOnPowerOn()
	assert.False(bootOnPowerOn)

	// Then set the capacity
	//
	blade.SetCapacity(stdCapacity)

	details = blade.GetDetails()
	require.NotNil(details)
	assert.Equal(stdDetails, details)

	capacity = blade.GetCapacity()
	require.NotNil(capacity)
	assert.Equal(stdCapacity, capacity)

	bootInfo = blade.GetBootInfo()
	require.Nil(bootInfo)

	bootOnPowerOn = blade.GetBootOnPowerOn()
	assert.False(bootOnPowerOn)

	// Finally, also set the bootInfo
	//
	blade.SetBootInfo(stdBootInfo)
	blade.SetBootPowerOn(stdBootOnPowerOn)

	details = blade.GetDetails()
	require.NotNil(details)
	assert.Equal(stdDetails, details)

	capacity = blade.GetCapacity()
	require.NotNil(capacity)
	assert.Equal(stdCapacity, capacity)

	bootInfo = blade.GetBootInfo()
	require.NotNil(bootInfo)
	assert.Equal(stdBootInfo, bootInfo)

	bootOnPowerOn = blade.GetBootOnPowerOn()
	assert.Equal(stdBootOnPowerOn, bootOnPowerOn)

	// This will actually attempt to read the blade from the store. Since
	// we have yet to create the blade, we expect to see a "not found"
	// type error.
	//
	rev, err = blade.Read(ctx, ViewDefinition)
	require.Error(err)
	assert.Equal(errors.ErrBladeNotFound{Region: blade.Region, Zone: blade.Zone, Rack: blade.Rack, Blade: blade.ID}, err)
	assert.Equal(store.RevisionInvalid, rev)

	// Clear the ports and check the update fails
	//
	// Note: the ordering of this and the subsequent statements assume
	//       the Update() call checks for the details being present before
	//       checking for the capacity or bootInfo being present. Any
	//	     change in the ordering may result in the tests needing amendment.
	//
	blade.SetBootInfo(nil)
	blade.SetBootPowerOn(false)

	rev, err = blade.Update(ctx, false, ViewDefinition)
	require.Equal(errors.ErrBootInfoNotAvailable("blade"), err)
	assert.Equal(store.RevisionInvalid, rev)

	rev, err = blade.Update(ctx, true, ViewDefinition)
	require.Equal(errors.ErrBootInfoNotAvailable("blade"), err)
	assert.Equal(store.RevisionInvalid, rev)

	blade.SetCapacity(nil)

	rev, err = blade.Update(ctx, false, ViewDefinition)
	require.Equal(errors.ErrCapacityNotAvailable("blade"), err)
	assert.Equal(store.RevisionInvalid, rev)

	rev, err = blade.Update(ctx, true, ViewDefinition)
	require.Equal(errors.ErrCapacityNotAvailable("blade"), err)
	assert.Equal(store.RevisionInvalid, rev)

	blade.SetDetails(nil)

	rev, err = blade.Update(ctx, false, ViewDefinition)
	require.Error(err)
	assert.Equal(errors.ErrDetailsNotAvailable("blade"), err)
	assert.Equal(store.RevisionInvalid, rev)

	rev, err = blade.Update(ctx, true, ViewDefinition)
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

func (ts *definitionTestSuite) TestNewRegionWithCreate() {
	assert := ts.Assert()
	require := ts.Require()

	stdSuffix := "TestNewRegionWithCreate"
	stdDetails := ts.stdRegionDetails(stdSuffix)

	ctx := context.Background()

	region, err := ts.inventory.NewRegion(namespace.DefinitionTable, ts.regionName(stdSuffix))
	require.NoError(err)

	region.SetDetails(stdDetails)

	rev, err := region.Create(ctx, ViewDefinition)

	require.NoError(err)
	assert.NotEqual(store.RevisionInvalid, rev)

	revDel, err := region.Delete(ctx, false, ViewDefinition)
	require.NoError(err)
	assert.Less(rev, revDel)
}

func (ts *definitionTestSuite) TestNewZoneWithCreate() {
	assert := ts.Assert()
	require := ts.Require()

	stdSuffix := "TestNewZoneWithCreate"
	stdDetails := ts.stdZoneDetails(stdSuffix)

	ctx := context.Background()

	zone, err := ts.inventory.NewZone(namespace.DefinitionTable, ts.regionName(stdSuffix), ts.zoneName(stdSuffix))
	require.NoError(err)

	zone.SetDetails(stdDetails)

	rev, err := zone.Create(ctx, ViewDefinition)

	require.NoError(err)
	assert.NotEqual(store.RevisionInvalid, rev)

	revDel, err := zone.Delete(ctx, false, ViewDefinition)
	require.NoError(err)
	assert.Less(rev, revDel)
}

func (ts *definitionTestSuite) TestNewRackWithCreate() {
	assert := ts.Assert()
	require := ts.Require()

	stdSuffix := "TestNewRackWithCreate"
	stdDetails := ts.stdRackDetails(stdSuffix)

	ctx := context.Background()

	rack, err := ts.inventory.NewRack(
		namespace.DefinitionTable,
		ts.regionName(stdSuffix),
		ts.zoneName(stdSuffix),
		ts.rackName(stdSuffix),
	)
	require.NoError(err)

	rack.SetDetails(stdDetails)

	rev, err := rack.Create(ctx, ViewDefinition)

	require.NoError(err)
	assert.NotEqual(store.RevisionInvalid, rev)

	revDel, err := rack.Delete(ctx, false, ViewDefinition)
	require.NoError(err)
	assert.Less(rev, revDel)
}

func (ts *definitionTestSuite) TestNewPduWithCreate() {
	assert := ts.Assert()
	require := ts.Require()

	stdSuffix := "TestNewPduWithCreate"
	stdDetails := ts.stdPduDetails(4)
	stdPorts := ts.stdPowerPorts(4)

	ctx := context.Background()

	pdu, err := ts.inventory.NewPdu(
		namespace.DefinitionTable,
		ts.regionName(stdSuffix),
		ts.zoneName(stdSuffix),
		ts.rackName(stdSuffix),
		ts.pduID(1),
	)
	require.NoError(err)

	pdu.SetDetails(stdDetails)
	pdu.SetPorts(stdPorts)

	rev, err := pdu.Create(ctx, ViewDefinition)

	require.NoError(err)
	assert.NotEqual(store.RevisionInvalid, rev)

	revDel, err := pdu.Delete(ctx, false, ViewDefinition)
	require.NoError(err)
	assert.Less(rev, revDel)
}

func (ts *definitionTestSuite) TestNewTorWithCreate() {
	assert := ts.Assert()
	require := ts.Require()

	stdSuffix := "TestNewTorWithCreate"
	stdDetails := ts.stdTorDetails()
	stdPorts := ts.stdNetworkPorts(4)

	ctx := context.Background()

	tor, err := ts.inventory.NewTor(
		namespace.DefinitionTable,
		ts.regionName(stdSuffix),
		ts.zoneName(stdSuffix),
		ts.rackName(stdSuffix),
		ts.pduID(1),
	)
	require.NoError(err)

	tor.SetDetails(stdDetails)
	tor.SetPorts(stdPorts)

	rev, err := tor.Create(ctx, ViewDefinition)

	require.NoError(err)
	assert.NotEqual(store.RevisionInvalid, rev)

	revDel, err := tor.Delete(ctx, false, ViewDefinition)
	require.NoError(err)
	assert.Less(rev, revDel)
}

func (ts *definitionTestSuite) TestNewBladeWithCreate() {
	assert := ts.Assert()
	require := ts.Require()

	stdSuffix := "TestNewBladeWithCreate"
	stdDetails := ts.stdBladeDetails()
	stdCapacity := ts.stdBladeCapacity()
	stdBootInfo := ts.stdBladeBootInfo()
	stdBootOnPowerOn := true

	ctx := context.Background()

	blade, err := ts.inventory.NewBlade(
		namespace.DefinitionTable,
		ts.regionName(stdSuffix),
		ts.zoneName(stdSuffix),
		ts.rackName(stdSuffix),
		ts.bladeID(1),
	)
	require.NoError(err)

	blade.SetDetails(stdDetails)
	blade.SetCapacity(stdCapacity)
	blade.SetBootInfo(stdBootInfo)
	blade.SetBootPowerOn(stdBootOnPowerOn)

	rev, err := blade.Create(ctx, ViewDefinition)

	require.NoError(err)
	assert.NotEqual(store.RevisionInvalid, rev)

	revDel, err := blade.Delete(ctx, false, ViewDefinition)
	require.NoError(err)
	assert.Less(rev, revDel)
}

func (ts *definitionTestSuite) TestRootNewChild() {
	assert := ts.Assert()
	require := ts.Require()

	stdSuffix := "TestRootNewChild"

	regionName := ts.regionName(stdSuffix)

	ctx := context.Background()

	root, err := ts.inventory.NewRoot(namespace.DefinitionTable)
	require.NoError(err)

	root.SetDetails(ts.stdRootDetails(stdSuffix))

	region, err := root.NewChild(regionName)
	require.NoError(err)

	region.SetDetails(ts.stdRegionDetails(stdSuffix))

	rev, err := region.Create(ctx, ViewDefinition)
	require.NoError(err)
	assert.NotEqual(store.RevisionInvalid, rev)

	revDel, err := region.Delete(ctx, false, ViewDefinition)
	require.NoError(err)
	assert.Less(rev, revDel)
}

func (ts *definitionTestSuite) TestNewChildZone() {
	assert := ts.Assert()
	require := ts.Require()

	stdSuffix := "TestNewChildZone"

	regionName := ts.regionName(stdSuffix)
	zoneName := ts.zoneName(stdSuffix)

	ctx := context.Background()

	root, err := ts.inventory.NewRoot(namespace.DefinitionTable)
	require.NoError(err)

	region, err := root.NewChild(regionName)
	require.NoError(err)

	zone, err := region.NewChild(zoneName)
	require.NoError(err)

	zone.SetDetails(ts.stdZoneDetails(stdSuffix))

	rev, err := zone.Create(ctx, ViewDefinition)
	require.NoError(err)
	assert.NotEqual(store.RevisionInvalid, rev)

	revDel, err := zone.Delete(ctx, false, ViewDefinition)
	require.NoError(err)
	assert.Less(rev, revDel)
}

func (ts *definitionTestSuite) TestNewChildRack() {
	assert := ts.Assert()
	require := ts.Require()

	stdSuffix := "TestNewChildRack"

	regionName := ts.regionName(stdSuffix)
	zoneName := ts.zoneName(stdSuffix)
	rackName := ts.rackName(stdSuffix)

	ctx := context.Background()

	root, err := ts.inventory.NewRoot(namespace.DefinitionTable)
	require.NoError(err)

	region, err := root.NewChild(regionName)
	require.NoError(err)

	zone, err := region.NewChild(zoneName)
	require.NoError(err)

	rack, err := zone.NewChild(rackName)
	require.NoError(err)

	rack.SetDetails(ts.stdRackDetails(stdSuffix))

	rev, err := rack.Create(ctx, ViewDefinition)
	require.NoError(err)
	assert.NotEqual(store.RevisionInvalid, rev)

	revDel, err := rack.Delete(ctx, false, ViewDefinition)
	require.NoError(err)
	assert.Less(rev, revDel)
}

func (ts *definitionTestSuite) TestNewChildPdu() {
	assert := ts.Assert()
	require := ts.Require()

	stdSuffix := "TestNewChildPdu"

	regionName := ts.regionName(stdSuffix)
	zoneName := ts.zoneName(stdSuffix)
	rackName := ts.rackName(stdSuffix)
	pduID := ts.pduID(1)

	ctx := context.Background()

	root, err := ts.inventory.NewRoot(namespace.DefinitionTable)
	require.NoError(err)

	region, err := root.NewChild(regionName)
	require.NoError(err)

	zone, err := region.NewChild(zoneName)
	require.NoError(err)

	rack, err := zone.NewChild(rackName)
	require.NoError(err)

	pdu, err := rack.NewPdu(pduID)
	require.NoError(err)

	pdu.SetDetails(ts.stdPduDetails(pduID))
	pdu.SetPorts(ts.stdPowerPorts(8))

	rev, err := pdu.Create(ctx, ViewDefinition)
	require.NoError(err)
	assert.NotEqual(store.RevisionInvalid, rev)

	revDel, err := pdu.Delete(ctx, false, ViewDefinition)
	require.NoError(err)
	assert.Less(rev, revDel)
}

func (ts *definitionTestSuite) TestNewChildTor() {
	assert := ts.Assert()
	require := ts.Require()

	stdSuffix := "TestNewChildTor"

	regionName := ts.regionName(stdSuffix)
	zoneName := ts.zoneName(stdSuffix)
	rackName := ts.rackName(stdSuffix)
	pduID := ts.pduID(1)

	ctx := context.Background()

	root, err := ts.inventory.NewRoot(namespace.DefinitionTable)
	require.NoError(err)

	region, err := root.NewChild(regionName)
	require.NoError(err)

	zone, err := region.NewChild(zoneName)
	require.NoError(err)

	rack, err := zone.NewChild(rackName)
	require.NoError(err)

	tor, err := rack.NewTor(pduID)
	require.NoError(err)

	tor.SetDetails(ts.stdTorDetails())
	tor.SetPorts(ts.stdNetworkPorts(8))

	rev, err := tor.Create(ctx, ViewDefinition)
	require.NoError(err)
	assert.NotEqual(store.RevisionInvalid, rev)

	revDel, err := tor.Delete(ctx, false, ViewDefinition)
	require.NoError(err)
	assert.Less(rev, revDel)
}

func (ts *definitionTestSuite) TestNewChildBlade() {
	assert := ts.Assert()
	require := ts.Require()

	stdSuffix := "TestNewChildBlade"

	regionName := ts.regionName(stdSuffix)
	zoneName := ts.zoneName(stdSuffix)
	rackName := ts.rackName(stdSuffix)
	bladeID := ts.bladeID(1)

	stdDetails := ts.stdBladeDetails()
	stdCapacity := ts.stdBladeCapacity()
	stdBootInfo := ts.stdBladeBootInfo()

	stdBootOnPowerOn := true

	ctx := context.Background()

	root, err := ts.inventory.NewRoot(namespace.DefinitionTable)
	require.NoError(err)

	region, err := root.NewChild(regionName)
	require.NoError(err)

	zone, err := region.NewChild(zoneName)
	require.NoError(err)

	rack, err := zone.NewChild(rackName)
	require.NoError(err)

	blade, err := rack.NewBlade(bladeID)
	require.NoError(err)

	blade.SetDetails(stdDetails)
	blade.SetCapacity(stdCapacity)
	blade.SetBootInfo(stdBootInfo)
	blade.SetBootPowerOn(stdBootOnPowerOn)

	rev, err := blade.Create(ctx, ViewDefinition)
	require.NoError(err)
	assert.NotEqual(store.RevisionInvalid, rev)

	revDel, err := blade.Delete(ctx, false, ViewDefinition)
	require.NoError(err)
	assert.Less(rev, revDel)
}

func (ts *definitionTestSuite) TestRegionReadDetails() {
	assert := ts.Assert()
	require := ts.Require()

	stdSuffix := "TestRegionReadDetails"

	regionName := ts.regionName(stdSuffix)

	stdDetails := ts.stdRegionDetails(stdSuffix)

	ctx := context.Background()

	r, err := ts.inventory.NewRegion(namespace.DefinitionTable, regionName)
	require.NoError(err)

	r.SetDetails(stdDetails)

	rev, err := r.Create(ctx, ViewDefinition)
	require.NoError(err)
	assert.NotEqual(store.RevisionInvalid, rev)

	rev2 := r.GetRevision()
	assert.Equal(rev, rev2)

	// Read the region back using the direct constructor
	//
	rRead, err := ts.inventory.NewRegion(namespace.DefinitionTable, regionName)
	require.NoError(err)

	revRead, err := rRead.Read(ctx, ViewDefinition)
	require.NoError(err)
	assert.Equal(rev, revRead)
	assert.Equal(revRead, rRead.GetRevision())

	detRead := rRead.GetDetails()
	require.NoError(err)
	assert.Equal(stdDetails, detRead)

	// Read the region back using the relative constructor
	//
	root, err := ts.inventory.NewRoot(namespace.DefinitionTable)
	require.NoError(err)

	cr, err := root.NewChild(regionName)
	require.NoError(err)

	crRev, err := cr.Read(ctx, ViewDefinition)
	require.NoError(err)
	assert.Equal(rev, crRev)
	assert.Equal(revRead, cr.GetRevision())

	crDet := cr.GetDetails()
	require.NoError(err)
	assert.Equal(stdDetails, crDet)

	revDel, err := cr.Delete(ctx, false, ViewDefinition)
	require.NoError(err)
	assert.Less(crRev, revDel)
}

func (ts *definitionTestSuite) TestZoneReadDetails() {
	assert := ts.Assert()
	require := ts.Require()

	stdSuffix := "TestZoneReadDetails"

	regionName := ts.regionName(stdSuffix)
	zoneName := ts.zoneName(stdSuffix)

	stdDetails := ts.stdZoneDetails(stdSuffix)

	ctx := context.Background()

	z, err := ts.inventory.NewZone(namespace.DefinitionTable, regionName, zoneName)
	require.NoError(err)

	z.SetDetails(stdDetails)

	rev, err := z.Create(ctx, ViewDefinition)
	require.NoError(err)
	assert.NotEqual(store.RevisionInvalid, rev)

	rev2 := z.GetRevision()
	assert.Equal(rev, rev2)

	// Read the zone back using the direct constructor
	//
	rRead, err := ts.inventory.NewZone(namespace.DefinitionTable, regionName, zoneName)
	require.NoError(err)

	revRead, err := rRead.Read(ctx, ViewDefinition)
	require.NoError(err)
	assert.Equal(rev, revRead)
	assert.Equal(revRead, rRead.GetRevision())

	detRead := rRead.GetDetails()
	require.NoError(err)
	assert.Equal(stdDetails, detRead)

	// Read the zone back using the relative constructor
	//
	root, err := ts.inventory.NewRoot(namespace.DefinitionTable)
	require.NoError(err)

	region, err := root.NewChild(regionName)
	require.NoError(err)

	zone, err := region.NewChild(zoneName)
	require.NoError(err)

	zoneRev, err := zone.Read(ctx, ViewDefinition)
	require.NoError(err)
	assert.Equal(rev, zoneRev)
	assert.Equal(zoneRev, zone.GetRevision())

	zoneDet := zone.GetDetails()
	require.NoError(err)
	assert.Equal(stdDetails, zoneDet)

	revDel, err := zone.Delete(ctx, false, ViewDefinition)
	require.NoError(err)
	assert.Less(zoneRev, revDel)
}

func (ts *definitionTestSuite) TestRackReadDetails() {
	assert := ts.Assert()
	require := ts.Require()

	stdSuffix := "TestRackReadDetails"

	regionName := ts.regionName(stdSuffix)
	zoneName := ts.zoneName(stdSuffix)
	rackName := ts.rackName(stdSuffix)

	stdDetails := ts.stdRackDetails(stdSuffix)

	ctx := context.Background()

	r, err := ts.inventory.NewRack(
		namespace.DefinitionTable,
		regionName,
		zoneName,
		rackName,
	)
	require.NoError(err)

	r.SetDetails(stdDetails)

	rev, err := r.Create(ctx, ViewDefinition)
	require.NoError(err)
	assert.NotEqual(store.RevisionInvalid, rev)

	rev2 := r.GetRevision()
	assert.Equal(rev, rev2)

	// Read the region back using the direct constructor
	//
	rRead, err := ts.inventory.NewRack(
		namespace.DefinitionTable,
		regionName,
		zoneName,
		rackName,
	)
	require.NoError(err)

	revRead, err := rRead.Read(ctx, ViewDefinition)
	require.NoError(err)
	assert.Equal(rev, revRead)
	assert.Equal(revRead, rRead.GetRevision())

	detRead := rRead.GetDetails()
	require.NoError(err)
	assert.Equal(stdDetails, detRead)

	// Read the zone back using the relative constructor
	//
	root, err := ts.inventory.NewRoot( namespace.DefinitionTable)
	require.NoError(err)

	region, err := root.NewChild(regionName)
	require.NoError(err)

	zone, err := region.NewChild(zoneName)
	require.NoError(err)

	rack, err := zone.NewChild(rackName)
	require.NoError(err)

	rackRev, err := rack.Read(ctx, ViewDefinition)
	require.NoError(err)
	assert.Equal(rev, rackRev)
	assert.Equal(rackRev, rack.GetRevision())

	rackDet := rack.GetDetails()
	require.NoError(err)
	assert.Equal(stdDetails, rackDet)

	revDel, err := rack.Delete(ctx, false, ViewDefinition)
	require.NoError(err)
	assert.Less(rackRev, revDel)
}

func (ts *definitionTestSuite) TestPduReadDetails() {
	assert := ts.Assert()
	require := ts.Require()

	stdSuffix := "TestPduReadDetails"

	regionName := ts.regionName(stdSuffix)
	zoneName := ts.zoneName(stdSuffix)
	rackName := ts.rackName(stdSuffix)
	pduID := int64(2)

	stdDetails := ts.stdPduDetails(pduID)
	stdPorts := ts.stdPowerPorts(4)

	ctx := context.Background()

	p, err := ts.inventory.NewPdu(
		namespace.DefinitionTable,
		regionName,
		zoneName,
		rackName,
		pduID,
	)
	require.NoError(err)

	p.SetDetails(stdDetails)
	p.SetPorts(stdPorts)

	rev, err := p.Create(ctx, ViewDefinition)
	require.NoError(err)
	assert.NotEqual(store.RevisionInvalid, rev)

	rev2 := p.GetRevision()
	assert.Equal(rev, rev2)

	// Read the region back using the direct constructor
	//
	p2, err := ts.inventory.NewPdu(
		namespace.DefinitionTable,
		regionName,
		zoneName,
		rackName,
		pduID,
	)
	require.NoError(err)

	p2Rev, err := p2.Read(ctx, ViewDefinition)
	require.NoError(err)
	assert.Equal(rev, p2Rev)
	assert.Equal(p2Rev, p2.GetRevision())

	p2Det := p2.GetDetails()
	require.NoError(err)
	assert.Equal(stdDetails, p2Det)

	// Read the zone back using the relative constructor
	//
	root, err := ts.inventory.NewRoot(namespace.DefinitionTable)
	require.NoError(err)

	region, err := root.NewChild(regionName)
	require.NoError(err)

	zone, err := region.NewChild(zoneName)
	require.NoError(err)

	rack, err := zone.NewChild(rackName)
	require.NoError(err)

	pdu, err := rack.NewPdu(pduID)
	require.NoError(err)

	pduRev, err := pdu.Read(ctx, ViewDefinition)
	require.NoError(err)
	assert.Equal(rev, pduRev)
	assert.Equal(pduRev, pdu.GetRevision())

	pduDet := pdu.GetDetails()
	require.NoError(err)
	assert.Equal(stdDetails, pduDet)

	pduPorts := pdu.GetPorts()
	require.NoError(err)
	assert.Equal(stdPorts, pduPorts)

	revDel, err := pdu.Delete(ctx, false, ViewDefinition)
	require.NoError(err)
	assert.Less(pduRev, revDel)
}

func (ts *definitionTestSuite) TestTorReadDetails() {
	assert := ts.Assert()
	require := ts.Require()

	stdSuffix := "TestTorReadDetails"

	regionName := ts.regionName(stdSuffix)
	zoneName := ts.zoneName(stdSuffix)
	rackName := ts.rackName(stdSuffix)
	torID := int64(1)

	stdDetails := ts.stdTorDetails()
	stdPorts := ts.stdNetworkPorts(4)

	ctx := context.Background()

	t, err := ts.inventory.NewTor(
		namespace.DefinitionTable,
		regionName,
		zoneName,
		rackName,
		torID,
	)
	require.NoError(err)

	t.SetDetails(stdDetails)
	t.SetPorts(stdPorts)

	rev, err := t.Create(ctx, ViewDefinition)
	require.NoError(err)
	assert.NotEqual(store.RevisionInvalid, rev)

	rev2 := t.GetRevision()
	assert.Equal(rev, rev2)

	// Read the region back using the direct constructor
	//
	t2, err := ts.inventory.NewTor(
		namespace.DefinitionTable,
		regionName,
		zoneName,
		rackName,
		torID,
	)
	require.NoError(err)

	t2Rev, err := t2.Read(ctx, ViewDefinition)
	require.NoError(err)
	assert.Equal(rev, t2Rev)
	assert.Equal(t2Rev, t2.GetRevision())

	p2Det := t2.GetDetails()
	require.NoError(err)
	assert.Equal(stdDetails, p2Det)

	// Read the zone back using the relative constructor
	//
	root, err := ts.inventory.NewRoot(namespace.DefinitionTable)
	require.NoError(err)

	region, err := root.NewChild(regionName)
	require.NoError(err)

	zone, err := region.NewChild(zoneName)
	require.NoError(err)

	rack, err := zone.NewChild(rackName)
	require.NoError(err)

	tor, err := rack.NewTor(torID)
	require.NoError(err)

	torRev, err := tor.Read(ctx, ViewDefinition)
	require.NoError(err)
	assert.Equal(rev, torRev)
	assert.Equal(torRev, tor.GetRevision())

	pduDet := tor.GetDetails()
	require.NoError(err)
	assert.Equal(stdDetails, pduDet)

	torPorts := tor.GetPorts()
	require.NoError(err)
	assert.Equal(stdPorts, torPorts)

	revDel, err := tor.Delete(ctx, false, ViewDefinition)
	require.NoError(err)
	assert.Less(torRev, revDel)
}

func (ts *definitionTestSuite) TestBladeReadDetails() {
	assert := ts.Assert()
	require := ts.Require()

	stdSuffix := "TestBladeReadDetails"

	regionName := ts.regionName(stdSuffix)
	zoneName := ts.zoneName(stdSuffix)
	rackName := ts.rackName(stdSuffix)
	bladeID := int64(5)

	stdDetails := ts.stdBladeDetails()
	stdCapacity := ts.stdBladeCapacity()
	stdBootInfo := ts.stdBladeBootInfo()
	stdBootOnPowerOn := true

	ctx := context.Background()

	b, err := ts.inventory.NewBlade(
		namespace.DefinitionTable,
		regionName,
		zoneName,
		rackName,
		bladeID,
	)
	require.NoError(err)

	b.SetDetails(stdDetails)
	b.SetCapacity(stdCapacity)
	b.SetBootInfo(stdBootInfo)
	b.SetBootPowerOn(stdBootOnPowerOn)

	rev, err := b.Create(ctx, ViewDefinition)
	require.NoError(err)
	assert.NotEqual(store.RevisionInvalid, rev)

	rev2 := b.GetRevision()
	assert.Equal(rev, rev2)

	// Read the region back using the direct constructor
	//
	b2, err := ts.inventory.NewBlade(
		namespace.DefinitionTable,
		regionName,
		zoneName,
		rackName,
		bladeID,
	)
	require.NoError(err)

	t2Rev, err := b2.Read(ctx, ViewDefinition)
	require.NoError(err)
	assert.Equal(rev, t2Rev)
	assert.Equal(t2Rev, b2.GetRevision())

	p2Det := b2.GetDetails()
	require.NoError(err)
	assert.Equal(stdDetails, p2Det)

	// Read the zone back using the relative constructor
	//
	root, err := ts.inventory.NewRoot(namespace.DefinitionTable)
	require.NoError(err)

	region, err := root.NewChild(regionName)
	require.NoError(err)

	zone, err := region.NewChild(zoneName)
	require.NoError(err)

	rack, err := zone.NewChild(rackName)
	require.NoError(err)

	blade, err := rack.NewBlade(bladeID)
	require.NoError(err)

	bladeRev, err := blade.Read(ctx, ViewDefinition)
	require.NoError(err)
	assert.Equal(rev, bladeRev)
	assert.Equal(bladeRev, blade.GetRevision())

	bladeDet := blade.GetDetails()
	require.NotNil(bladeDet)
	assert.Equal(stdDetails, bladeDet)

	bladeCapacity := blade.GetCapacity()
	require.NotNil(bladeCapacity)
	assert.Equal(stdCapacity, bladeCapacity)

	bladeBootInfo := blade.GetBootInfo()
	require.NotNil(bladeBootInfo)
	assert.Equal(stdBootInfo, bladeBootInfo)

	bladeBootOnPowerOn := blade.GetBootOnPowerOn()
	assert.Equal(stdBootOnPowerOn, bladeBootOnPowerOn)

	revDel, err := blade.Delete(ctx, false, ViewDefinition)
	require.NoError(err)
	assert.Less(bladeRev, revDel)
}

func (ts *definitionTestSuite) TestRegionUpdateDetails() {

	assert := ts.Assert()
	require := ts.Require()

	stdSuffix := "TestRegionUpdateDetails"

	regionName := ts.regionName(stdSuffix)

	stdDetails := ts.stdRegionDetails(stdSuffix)

	ctx := context.Background()

	r, err := ts.inventory.NewRegion(namespace.DefinitionTable, regionName)
	require.NoError(err)

	r.SetDetails(stdDetails)

	revCreate, err := r.Create(ctx, ViewDefinition)
	require.NoError(err)
	assert.NotEqual(store.RevisionInvalid, revCreate)

	// Read the region back using the relative constructor
	//
	root, err := ts.inventory.NewRoot(namespace.DefinitionTable)
	require.NoError(err)

	cr, err := root.NewChild(regionName)
	require.NoError(err)

	revRead, err := cr.Read(ctx, ViewDefinition)
	require.NoError(err)
	assert.Equal(revCreate, revRead)

	details := cr.GetDetails()
	require.NotNil(details)

	details.State = pb.State_out_of_service
	details.Notes += " (out of service)"

	// Update the region using the direct constructor, conditional on revision
	//
	r.SetDetails(details)

	revUpdate, err := r.Update(ctx, false, ViewDefinition)
	require.NoError(err)
	assert.Less(revRead, revUpdate)

	// Verify update using relative constructor
	//
	revVerify, err := cr.Read(ctx, ViewDefinition)
	require.NoError(err)
	assert.Equal(revUpdate, revVerify)

	detailsVerify := cr.GetDetails()
	require.NotNil(detailsVerify)

	// Compare new details with original + deltas
	//
	assert.Equal(details, detailsVerify)

	revDel, err := cr.Delete(ctx, false, ViewDefinition)
	require.NoError(err)
	assert.Less(revUpdate, revDel)

}

func (ts *definitionTestSuite) TestZoneUpdateDetails() {

	assert := ts.Assert()
	require := ts.Require()

	stdSuffix := "TestZoneUpdateDetails"

	regionName := ts.regionName(stdSuffix)
	zoneName := ts.zoneName(stdSuffix)

	stdDetails := ts.stdZoneDetails(stdSuffix)

	ctx := context.Background()

	z, err := ts.inventory.NewZone(namespace.DefinitionTable, regionName, zoneName)
	require.NoError(err)

	z.SetDetails(stdDetails)

	revCreate, err := z.Create(ctx, ViewDefinition)
	require.NoError(err)
	assert.NotEqual(store.RevisionInvalid, revCreate)

	// Read the zone back using the relative constructor
	//
	root, err := ts.inventory.NewRoot(namespace.DefinitionTable)
	require.NoError(err)

	region, err := root.NewChild(regionName)
	require.NoError(err)

	cz, err := region.NewChild(zoneName)
	require.NoError(err)

	revRead, err := cz.Read(ctx, ViewDefinition)
	require.NoError(err)
	assert.Equal(revCreate, revRead)

	details := cz.GetDetails()
	require.NotNil(details)

	details.State = pb.State_out_of_service
	details.Notes += " (out of service)"

	// Update the record using the direct constructor, conditional on revision
	//
	z.SetDetails(details)

	revUpdate, err := z.Update(ctx, false, ViewDefinition)
	require.NoError(err)
	assert.Less(revRead, revUpdate)

	// Verify update using relative constructor
	//
	revVerify, err := cz.Read(ctx, ViewDefinition)
	require.NoError(err)
	assert.Equal(revUpdate, revVerify)

	detailsVerify := cz.GetDetails()
	require.NotNil(detailsVerify)

	// Compare new details with original + deltas
	//
	assert.Equal(details, detailsVerify)

	revDel, err := cz.Delete(ctx, false, ViewDefinition)
	require.NoError(err)
	assert.Less(revUpdate, revDel)
}

func (ts *definitionTestSuite) TestZoneWatch() {

	assert := ts.Assert()
	require := ts.Require()

	stdSuffix := "TestZoneWatch"

	regionName := ts.regionName(stdSuffix)
	zoneName := ts.zoneName(stdSuffix)
	rackName := ts.rackName(stdSuffix)
	bladeId := ts.bladeID(99)

	stdDetails := ts.stdZoneDetails(stdSuffix)

	zoneAddress  := namespace.NewZone(namespace.DefinitionTable, regionName, zoneName)
	rackAddress  := namespace.NewRack(namespace.DefinitionTable, regionName, zoneName, rackName)
	bladeAddress := namespace.NewBlade(namespace.DefinitionTable, regionName, zoneName, rackName, bladeId)

	ctx := context.Background()

	z, err := ts.inventory.NewZone(namespace.DefinitionTable, regionName, zoneName)
	require.NoError(err)

	z.SetDetails(stdDetails)

	// Set the watch prior to the create so expect to see a create event
	//
	watch, err := z.Watch(ctx)
	require.NoError(err)
	require.NotNil(watch)


	// Create the zone we are interested in monitoring.
	//
	revZoneCreate, err := z.Create(ctx, ViewDefinition)
	require.NoError(err)
	assert.NotEqual(store.RevisionInvalid, revZoneCreate)

	ev := <-watch.Events

	// Verify we got a create for the zone we just created and that
	// all the new and old revisions etc match.
	//
	require.NoError(ev.Err)

	require.Equal(store.WatchEventTypeCreate, ev.Type)
	assert.Equal(revZoneCreate, ev.Revision)
	assert.Equal(revZoneCreate, ev.NewRev)
	assert.Equal(store.RevisionInvalid, ev.OldRev)

	assert.Equal(zoneAddress, ev.Address)

	details := z.GetDetails()
	require.NotNil(details)

	details.State = pb.State_out_of_service
	details.Notes += " (out of service)"

	z.SetDetails(details)

	revZoneUpdate, err := z.Update(ctx, false, ViewDefinition)
	require.NoError(err)

	ev = <-watch.Events

	// Verify we got a create for the zone we just created and that
	// all the new and old revisions etc match.
	//
	require.NoError(ev.Err)

	require.Equal(store.WatchEventTypeUpdate, ev.Type)
	assert.Equal(revZoneUpdate, ev.Revision)
	assert.Equal(revZoneUpdate, ev.NewRev)
	assert.Equal(revZoneCreate, ev.OldRev)

	assert.Equal(zoneAddress, ev.Address)

	// Now create a child within the zone to verify we see an event
	// for the create of the child
	//
	r, err := z.NewChild(rackName)
	require.NoError(err)

	r.SetDetails(ts.stdRackDetails(stdSuffix))
	revRackCreate, err := r.Create(ctx, ViewDefinition)
	require.NoError(err)
	require.Less(revZoneUpdate, revRackCreate)

	ev = <-watch.Events

	// Verify we got a create for the rack we just created and that
	// all the new and old revisions etc match.
	//
	require.NoError(ev.Err)

	require.Equal(store.WatchEventTypeCreate, ev.Type)
	assert.Equal(revRackCreate, ev.Revision)
	assert.Equal(revRackCreate, ev.NewRev)
	assert.Equal(store.RevisionInvalid, ev.OldRev)

	assert.Equal(rackAddress, ev.Address)

	// Now create a blade under the rack
	//
	b, err := r.NewBlade(bladeId)
	require.NoError(err)

	b.SetDetails(ts.stdBladeDetails())
	b.SetCapacity(ts.stdBladeCapacity())
	b.SetBootInfo(ts.stdBladeBootInfo())
	b.SetBootPowerOn(ts.stdBladeBootOnPowerOn())

	revBladeCreate, err := b.Create(ctx, ViewDefinition)
	require.NoError(err)
	require.Less(revZoneUpdate, revBladeCreate)

	ev = <-watch.Events

	// Verify we got a create for the blade we just created and that
	// all the new and old revisions etc match.
	//
	require.NoError(ev.Err)

	require.Equal(store.WatchEventTypeCreate, ev.Type)
	assert.Equal(revBladeCreate, ev.Revision)
	assert.Equal(revBladeCreate, ev.NewRev)
	assert.Equal(store.RevisionInvalid, ev.OldRev)

	assert.Equal(bladeAddress, ev.Address)
}

func (ts *definitionTestSuite) TestZoneWatchSequence() {

	assert := ts.Assert()
	require := ts.Require()

	stdSuffix := "TestZoneWatchSequence"

	regionName := ts.regionName(stdSuffix)
	zoneName := ts.zoneName(stdSuffix)
	rackName := ts.rackName(stdSuffix)
	bladeId := ts.bladeID(99)

	stdDetails := ts.stdZoneDetails(stdSuffix)

	ctx := context.Background()

	z, err := ts.inventory.NewZone(namespace.DefinitionTable, regionName, zoneName)
	require.NoError(err)

	z.SetDetails(stdDetails)

	r, err := z.NewChild(rackName)
	require.NoError(err)

	r.SetDetails(ts.stdRackDetails(stdSuffix))

	b, err := r.NewBlade(bladeId)
	require.NoError(err)

	b.SetDetails(ts.stdBladeDetails())
	b.SetCapacity(ts.stdBladeCapacity())
	b.SetBootInfo(ts.stdBladeBootInfo())
	b.SetBootPowerOn(ts.stdBladeBootOnPowerOn())

	// Set the watch prior to the create so expect to see a create event
	//
	watch, err := z.Watch(ctx)
	require.NoError(err)
	require.NotNil(watch)


	type operation struct {
		revStore  int64
		revNew    int64
		revOld    int64
		eventType store.WatchEventType
		addr      *namespace.Address
	}

	// Keep an ordered set of operations which record the update being performed
	// along with some of the responses for that operation. These then form the
	// set of expectations to compare to the sequence of events that are indicated
	// by the watchpoint.
	//
	ops := make([]operation, 0)

	// Now we issue a sequence of updates and then verify the
	// events arrive in the expected order.
	//
	zoneAddress  := namespace.NewZone(namespace.DefinitionTable, regionName, zoneName)
	rackAddress  := namespace.NewRack(namespace.DefinitionTable, regionName, zoneName, rackName)
	bladeAddress := namespace.NewBlade(namespace.DefinitionTable, regionName, zoneName, rackName, bladeId)

	// Create the zone we are interested in monitoring.
	//
	revZone0, err := z.Create(ctx, ViewDefinition)
	require.NoError(err)

	ops = append(ops, operation{
		revStore:  revZone0,
		revNew:    revZone0,
		revOld:    store.RevisionInvalid,
		addr:      zoneAddress,
		eventType: store.WatchEventTypeCreate})


	details := z.GetDetails()
	require.NotNil(details)

	details.State = pb.State_out_of_service
	details.Notes += " (out of service)"

	z.SetDetails(details)

	revZone1, err := z.Update(ctx, false, ViewDefinition)
	require.NoError(err)

	ops = append(ops, operation{
		revStore:  revZone1,
		revNew:    revZone1,
		revOld:    revZone0,
		addr:      zoneAddress,
		eventType: store.WatchEventTypeUpdate})


	revRack0, err := r.Create(ctx, ViewDefinition)
	require.NoError(err)

	ops = append(ops, operation{
		revStore:  revRack0,
		revNew:    revRack0,
		revOld:    store.RevisionInvalid,
		addr:      rackAddress,
		eventType: store.WatchEventTypeCreate})


	revBlade0, err := b.Create(ctx, ViewDefinition)
	require.NoError(err)

	ops = append(ops, operation{
		revStore:  revBlade0,
		revNew:    revBlade0,
		revOld:    store.RevisionInvalid,
		addr:      bladeAddress,
		eventType: store.WatchEventTypeCreate})


	rackDetails := r.GetDetails()
	rackDetails.Condition = pb.Condition_out_for_repair
	r.SetDetails(rackDetails)

	revRack1, err := r.Update(ctx, true, ViewDefinition)
	require.NoError(err)

	ops = append(ops, operation{
		revStore:  revRack1,
		revNew:    revRack1,
		revOld:    revRack0,
		addr:      rackAddress,
		eventType: store.WatchEventTypeUpdate})


	bladeDetails := b.GetDetails()
	bladeDetails.Condition = pb.Condition_retired

	revBlade1, err := b.Update(ctx, true, ViewDefinition)
	require.NoError(err)

	ops = append(ops, operation{
		revStore:  revBlade1,
		revNew:    revBlade1,
		revOld:    revBlade0,
		addr:      bladeAddress,
		eventType: store.WatchEventTypeUpdate})


	rackDetails.Condition = pb.Condition_operational
	r.SetDetails(rackDetails)

	revRack2, err := r.Update(ctx, true, ViewDefinition)
	require.NoError(err)

	ops = append(ops, operation{
		revStore:  revRack2,
		revNew:    revRack2,
		revOld:    revRack1,
		addr:      rackAddress,
		eventType: store.WatchEventTypeUpdate})


	revBlade2, err := b.Delete(ctx, true, ViewDefinition)
	require.NoError(err)

	ops = append(ops, operation{
		revStore:  revBlade2,
		revNew:    store.RevisionInvalid,
		revOld:    revBlade1,
		addr:      bladeAddress,
		eventType: store.WatchEventTypeDelete})


	revRack3, err := r.Delete(ctx, true, ViewDefinition)
	require.NoError(err)

	ops = append(ops, operation{
		revStore:  revRack3,
		revNew:    store.RevisionInvalid,
		revOld:    revRack2,
		addr:      rackAddress,
		eventType: store.WatchEventTypeDelete})


	revZone2, err := z.Delete(ctx, true, ViewDefinition)
	require.NoError(err)

	ops = append(ops, operation{
		revStore:  revZone2,
		revNew:    store.RevisionInvalid,
		revOld:    revZone1,
		addr:      zoneAddress,
		eventType: store.WatchEventTypeDelete})

	// Verify the captured revisions only increase over the
	// set of operations
	//
	for i, op := range ops {
		if i == 0 {
			assert.Less(store.RevisionInvalid, op.revStore)
		} else {
			assert.Less(ops[i-1].revStore, op.revStore)
		}
	}

	// Verify the sequence of delivered events match expectations and
	// the recorded responses from the actual operations when they
	// were performed.
	//
	for _, op := range ops {
		ev, ok := <-watch.Events

		require.True(ok)
		require.NoError(ev.Err)

		assert.Equal(op.eventType, ev.Type)
		assert.Equal(op.revStore, ev.Revision)
		assert.Equal(op.revNew, ev.NewRev)
		assert.Equal(op.revOld, ev.OldRev)
		assert.Equal(op.addr, ev.Address)
	}
}

func (ts *definitionTestSuite) TestRackUpdateDetails() {

	assert := ts.Assert()
	require := ts.Require()

	stdSuffix := "TestRackUpdateDetails"

	regionName := ts.regionName(stdSuffix)
	zoneName := ts.zoneName(stdSuffix)
	rackName := ts.rackName(stdSuffix)

	stdDetails := ts.stdRackDetails(stdSuffix)

	ctx := context.Background()

	r, err := ts.inventory.NewRack(namespace.DefinitionTable, regionName, zoneName, rackName)
	require.NoError(err)

	r.SetDetails(stdDetails)

	revCreate, err := r.Create(ctx, ViewDefinition)
	require.NoError(err)
	assert.NotEqual(store.RevisionInvalid, revCreate)

	// Read the rack back using the relative constructor
	//
	root, err := ts.inventory.NewRoot(namespace.DefinitionTable)
	require.NoError(err)

	region, err := root.NewChild(regionName)
	require.NoError(err)

	zone, err := region.NewChild(zoneName)
	require.NoError(err)

	cr, err := zone.NewChild(rackName)
	require.NoError(err)

	revRead, err := cr.Read(ctx, ViewDefinition)
	require.NoError(err)
	assert.Equal(revCreate, revRead)

	details := cr.GetDetails()
	require.NotNil(details)

	details.Enabled = false
	details.Condition = pb.Condition_not_in_service
	details.Notes += " (out of service)"

	// Update the record using the direct constructor, conditional on revision
	//
	r.SetDetails(details)

	revUpdate, err := r.Update(ctx, false, ViewDefinition)
	require.NoError(err)
	assert.Less(revRead, revUpdate)

	// Verify update using relative constructor
	//
	revVerify, err := cr.Read(ctx, ViewDefinition)
	require.NoError(err)
	assert.Equal(revUpdate, revVerify)

	detailsVerify := cr.GetDetails()
	require.NotNil(detailsVerify)

	// Compare new details with original + deltas
	//
	assert.Equal(details, detailsVerify)

	revDel, err := cr.Delete(ctx, false, ViewDefinition)
	require.NoError(err)
	assert.Less(revUpdate, revDel)
}

func (ts *definitionTestSuite) TestPduUpdateDetails() {

	assert := ts.Assert()
	require := ts.Require()

	stdSuffix := "TestPduUpdateDetails"

	regionName := ts.regionName(stdSuffix)
	zoneName := ts.zoneName(stdSuffix)
	rackName := ts.rackName(stdSuffix)
	pduID := ts.pduID(1)

	stdDetails := ts.stdPduDetails(pduID)
	stdPorts := ts.stdPowerPorts(4)

	ctx := context.Background()

	p, err := ts.inventory.NewPdu(namespace.DefinitionTable, regionName, zoneName, rackName, pduID)
	require.NoError(err)

	p.SetDetails(stdDetails)
	p.SetPorts(stdPorts)

	revCreate, err := p.Create(ctx, ViewDefinition)
	require.NoError(err)
	assert.NotEqual(store.RevisionInvalid, revCreate)

	// Read the rack back using the relative constructor
	//
	root, err := ts.inventory.NewRoot(namespace.DefinitionTable)
	require.NoError(err)

	region, err := root.NewChild(regionName)
	require.NoError(err)

	zone, err := region.NewChild(zoneName)
	require.NoError(err)

	rack, err := zone.NewChild(rackName)
	require.NoError(err)

	cp, err := rack.NewPdu(pduID)
	require.NoError(err)

	revRead, err := cp.Read(ctx, ViewDefinition)
	require.NoError(err)
	assert.Equal(revCreate, revRead)

	details := cp.GetDetails()
	require.NotNil(details)

	details.Enabled = false
	details.Condition = pb.Condition_not_in_service

	// Update the record using the direct constructor, conditional on revision
	//
	p.SetDetails(details)

	revUpdate, err := p.Update(ctx, false, ViewDefinition)
	require.NoError(err)
	assert.Less(revRead, revUpdate)

	// Verify update using relative constructor
	//
	revVerify, err := cp.Read(ctx, ViewDefinition)
	require.NoError(err)
	assert.Equal(revUpdate, revVerify)

	detailsVerify := cp.GetDetails()
	require.NotNil(detailsVerify)

	portsVerify := cp.GetPorts()
	require.NotNil(portsVerify)

	// Compare new details with original + deltas
	//
	assert.Equal(details, detailsVerify)
	assert.Equal(stdPorts, portsVerify)

	revDel, err := cp.Delete(ctx, false, ViewDefinition)
	require.NoError(err)
	assert.Less(revUpdate, revDel)
}

func (ts *definitionTestSuite) TestTorUpdateDetails() {

	assert := ts.Assert()
	require := ts.Require()

	stdSuffix := "TestTorUpdateDetails"

	regionName := ts.regionName(stdSuffix)
	zoneName := ts.zoneName(stdSuffix)
	rackName := ts.rackName(stdSuffix)
	torID := ts.torID(1)

	stdDetails := ts.stdTorDetails()
	stdPorts := ts.stdNetworkPorts(4)

	ctx := context.Background()

	t, err := ts.inventory.NewTor(namespace.DefinitionTable, regionName, zoneName, rackName, torID)
	require.NoError(err)

	t.SetDetails(stdDetails)
	t.SetPorts(stdPorts)

	revCreate, err := t.Create(ctx, ViewDefinition)
	require.NoError(err)
	assert.NotEqual(store.RevisionInvalid, revCreate)

	// Read the rack back using the relative constructor
	//
	root, err := ts.inventory.NewRoot(namespace.DefinitionTable)
	require.NoError(err)

	region, err := root.NewChild(regionName)
	require.NoError(err)

	zone, err := region.NewChild(zoneName)
	require.NoError(err)

	rack, err := zone.NewChild(rackName)
	require.NoError(err)

	ct, err := rack.NewTor(torID)
	require.NoError(err)

	revRead, err := ct.Read(ctx, ViewDefinition)
	require.NoError(err)
	assert.Equal(revCreate, revRead)

	details := ct.GetDetails()
	require.NotNil(details)

	details.Enabled = false
	details.Condition = pb.Condition_not_in_service

	// Update the record using the direct constructor, conditional on revision
	//
	t.SetDetails(details)

	revUpdate, err := t.Update(ctx, false, ViewDefinition)
	require.NoError(err)
	assert.Less(revRead, revUpdate)

	// Verify update using relative constructor
	//
	revVerify, err := ct.Read(ctx, ViewDefinition)
	require.NoError(err)
	assert.Equal(revUpdate, revVerify)

	detailsVerify := ct.GetDetails()
	require.NotNil(detailsVerify)

	portsVerify := ct.GetPorts()
	require.NotNil(portsVerify)

	// Compare new details with original + deltas
	//
	assert.Equal(details, detailsVerify)
	assert.Equal(stdPorts, portsVerify)

	revDel, err := ct.Delete(ctx, false, ViewDefinition)
	require.NoError(err)
	assert.Less(revUpdate, revDel)
}

func (ts *definitionTestSuite) TestBladeUpdateDetails() {

	assert := ts.Assert()
	require := ts.Require()

	stdSuffix := "TestBladeUpdateDetails"

	regionName := ts.regionName(stdSuffix)
	zoneName := ts.zoneName(stdSuffix)
	rackName := ts.rackName(stdSuffix)
	bladeID := ts.bladeID(1)

	stdDetails := ts.stdBladeDetails()
	stdCapacity := ts.stdBladeCapacity()
	stdBootInfo := ts.stdBladeBootInfo()
	stdBootOnPowerOn := true

	ctx := context.Background()

	b, err := ts.inventory.NewBlade(namespace.DefinitionTable, regionName, zoneName, rackName, bladeID)
	require.NoError(err)

	b.SetDetails(stdDetails)
	b.SetCapacity(stdCapacity)
	b.SetBootInfo(stdBootInfo)
	b.SetBootPowerOn(stdBootOnPowerOn)

	revCreate, err := b.Create(ctx, ViewDefinition)
	require.NoError(err)
	assert.NotEqual(store.RevisionInvalid, revCreate)

	// Read the rack back using the relative constructor
	//
	root, err := ts.inventory.NewRoot(namespace.DefinitionTable)
	require.NoError(err)

	region, err := root.NewChild(regionName)
	require.NoError(err)

	zone, err := region.NewChild(zoneName)
	require.NoError(err)

	rack, err := zone.NewChild(rackName)
	require.NoError(err)

	cb, err := rack.NewBlade(bladeID)
	require.NoError(err)

	revRead, err := cb.Read(ctx, ViewDefinition)
	require.NoError(err)
	assert.Equal(revCreate, revRead)

	details := cb.GetDetails()
	require.NotNil(details)

	details.Enabled = false
	details.Condition = pb.Condition_not_in_service

	// Update the record using the direct constructor, conditional on revision
	//
	b.SetDetails(details)

	revUpdate, err := b.Update(ctx, false, ViewDefinition)
	require.NoError(err)
	assert.Less(revRead, revUpdate)

	// Verify update using relative constructor
	//
	revVerify, err := cb.Read(ctx, ViewDefinition)
	require.NoError(err)
	assert.Equal(revUpdate, revVerify)

	detailsVerify := cb.GetDetails()
	require.NotNil(detailsVerify)

	capacityVerify := cb.GetCapacity()
	require.NotNil(capacityVerify)

	bootInfoVerify := cb.GetBootInfo()
	require.NotNil(bootInfoVerify)

	bootOnPowerOnVerify := cb.GetBootOnPowerOn()

	// Compare new details with original + deltas
	//
	assert.Equal(details, detailsVerify)
	assert.Equal(stdCapacity, capacityVerify)
	assert.Equal(stdBootInfo, bootInfoVerify)
	assert.Equal(stdBootOnPowerOn, bootOnPowerOnVerify)

	revDel, err := cb.Delete(ctx, false, ViewDefinition)
	require.NoError(err)
	assert.Less(revUpdate, revDel)
}

func (ts *definitionTestSuite) TestRootListChildren() {
	assert := ts.Assert()
	require := ts.Require()

	ctx := context.Background()

	root, err := ts.inventory.NewRoot(namespace.DefinitionTableStdTest)
	require.NoError(err)

	_, regions, err := root.ListChildren(ctx)
	require.NoError(err)
	assert.Equal(ts.regionCount, len(regions))
}

func (ts *definitionTestSuite) TestRegionListChildren() {
	assert := ts.Assert()
	require := ts.Require()

	ctx := context.Background()

	root, err := ts.inventory.NewRoot(namespace.DefinitionTableStdTest)
	require.NoError(err)

	_, regions, err := root.ListChildren(ctx)
	require.NoError(err)
	assert.Equal(ts.regionCount, len(regions))

	for _, v := range regions {
		region, err := root.NewChild(v)
		require.NoError(err)

		_, zones, err := region.ListChildren(ctx)
		require.NoError(err)
		assert.Equal(ts.zonesPerRegion, len(zones))
	}
}

func (ts *definitionTestSuite) TestZoneListChildren() {
	assert := ts.Assert()
	require := ts.Require()

	ctx := context.Background()

	root, err := ts.inventory.NewRoot(namespace.DefinitionTableStdTest)
	require.NoError(err)

	_, regions, err := root.ListChildren(ctx)
	require.NoError(err)
	assert.Equal(ts.regionCount, len(regions))

	for _, v := range regions {
		region, err := root.NewChild(v)
		require.NoError(err)

		_, zones, err := region.ListChildren(ctx)
		require.NoError(err)
		assert.Equal(ts.zonesPerRegion, len(zones))

		for _, v := range zones {
			zone, err := region.NewChild(v)
			require.NoError(err)

			_, racks, err := zone.ListChildren(ctx)
			require.NoError(err)
			assert.Equal(ts.racksPerZone, len(racks))
		}
	}
}

func (ts *definitionTestSuite) TestRackListChildren() {
	assert := ts.Assert()
	require := ts.Require()

	ctx := context.Background()

	root, err := ts.inventory.NewRoot(namespace.DefinitionTableStdTest)
	require.NoError(err)

	_, regions, err := root.ListChildren(ctx)
	require.NoError(err)
	assert.Equal(ts.regionCount, len(regions))

	for _, v := range regions {
		region, err := root.NewChild(v)
		require.NoError(err)

		_, zones, err := region.ListChildren(ctx)
		require.NoError(err)
		assert.Equal(ts.zonesPerRegion, len(zones))

		for _, v := range zones {
			zone, err := region.NewChild(v)
			require.NoError(err)

			_, racks, err := zone.ListChildren(ctx)
			require.NoError(err)
			assert.Equal(ts.racksPerZone, len(racks))

			for _, v := range racks {
				rack, err := zone.NewChild(v)
				require.NoError(err)

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

func (ts *definitionTestSuite) TestRootFetchChildren() {
	assert := ts.Assert()
	require := ts.Require()

	ctx := context.Background()

	root, err := ts.inventory.NewRoot(namespace.DefinitionTableStdTest)
	require.NoError(err)

	_, regions, err := root.FetchChildren(ctx)
	require.NoError(err)
	assert.Equal(ts.regionCount, len(*regions))

	for n, v := range *regions {
		ts.verifyStandardInventoryRegionDetails(n, v.GetDetails())
	}
}

func (ts *definitionTestSuite) TestRegionFetchChildren() {
	assert := ts.Assert()
	require := ts.Require()

	ctx := context.Background()

	root, err := ts.inventory.NewRoot(namespace.DefinitionTableStdTest)
	require.NoError(err)

	_, regions, err := root.FetchChildren(ctx)
	require.NoError(err)
	assert.Equal(ts.regionCount, len(*regions))

	for n, v := range *regions {
		ts.verifyStandardInventoryRegionDetails(n, v.GetDetails())

		region, err := root.NewChild(n)
		require.NoError(err)

		_, zones, err := region.FetchChildren(ctx)
		require.NoError(err)
		assert.Equal(ts.zonesPerRegion, len(*zones))

		for n, v := range *zones {
			ts.verifyStandardInventoryZoneDetails(n, v.GetDetails())
		}
	}
}

func (ts *definitionTestSuite) TestZoneFetchChildren() {
	assert := ts.Assert()
	require := ts.Require()

	ctx := context.Background()

	root, err := ts.inventory.NewRoot(namespace.DefinitionTableStdTest)
	require.NoError(err)

	_, regions, err := root.FetchChildren(ctx)
	require.NoError(err)
	assert.Equal(ts.regionCount, len(*regions))

	for n, v := range *regions {
		ts.verifyStandardInventoryRegionDetails(n, v.GetDetails())

		region, err := root.NewChild(n)
		require.NoError(err)

		_, zones, err := region.FetchChildren(ctx)
		require.NoError(err)
		assert.Equal(ts.zonesPerRegion, len(*zones))

		for n, v := range *zones {
			ts.verifyStandardInventoryZoneDetails(n, v.GetDetails())

			zone, err := region.NewChild(n)
			require.NoError(err)

			_, racks, err := zone.FetchChildren(ctx)
			require.NoError(err)
			assert.Equal(ts.zonesPerRegion, len(*zones))

			for n, v := range *racks {
				ts.verifyStandardInventoryRackDetails(n, v.GetDetails())
			}
		}
	}
}

func (ts *definitionTestSuite) TestRackFetchChildren() {
	assert := ts.Assert()
	require := ts.Require()

	ctx := context.Background()

	root, err := ts.inventory.NewRoot(namespace.DefinitionTableStdTest)
	require.NoError(err)

	_, regions, err := root.FetchChildren(ctx)
	require.NoError(err)
	assert.Equal(ts.regionCount, len(*regions))

	for n, v := range *regions {
		ts.verifyStandardInventoryRegionDetails(n, v.GetDetails())

		region, err := root.NewChild(n)
		require.NoError(err)

		_, zones, err := region.FetchChildren(ctx)
		require.NoError(err)
		assert.Equal(ts.zonesPerRegion, len(*zones))

		for n, v := range *zones {
			ts.verifyStandardInventoryZoneDetails(n, v.GetDetails())

			zone, err := region.NewChild(n)
			require.NoError(err)

			_, racks, err := zone.FetchChildren(ctx)
			require.NoError(err)
			assert.Equal(ts.zonesPerRegion, len(*zones))

			for n, v := range *racks {
				ts.verifyStandardInventoryRackDetails(n, v.GetDetails())

				rack, err := zone.NewChild(n)
				require.NoError(err)

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
				assert.Equal(ts.pdusPerRack, len(pdus))

				_, tors, err := rack.FetchTors(ctx)
				require.NoError(err)
				assert.Equal(ts.torsPerRack, len(tors))

				_, blades, err := rack.FetchBlades(ctx)
				require.NoError(err)
				assert.Equal(ts.bladesPerRack, len(blades))

				for n, v := range pdus {
					ts.verifyStandardInventoryPdu(n, v)
				}

				for n, v := range tors {
					ts.verifyStandardInventoryTor(n, &v)
				}

				for n, v := range blades {
					ts.verifyStandardInventoryBlade(n, &v)
				}
			}
		}
	}
}

func (ts *definitionTestSuite) TestReadInventoryFromStore() {
	require := ts.Require()

	_, err := ts.inventory.readInventoryDefinitionFromStore(context.Background())
	require.NoError(err)
}

func (ts *definitionTestSuite) TestReadInventoryDefinitionFromFileBasic() {
	require := ts.Require()

	_, err := ReadInventoryDefinitionFromFile(context.Background(), "./testdata/basic")
	require.NoError(err)
}

func (ts *definitionTestSuite) TestReadInventoryDefinitionFromFileReference() {
	require := ts.Require()

	_, err := ReadInventoryDefinitionFromFile(context.Background(), "./testdata/reference")
	require.NoError(err)
}

func (ts *definitionTestSuite) TestUpdateInventoryDefinitionBasic() {
	require := ts.Require()

	ctx := context.Background()

	err := ts.inventory.UpdateInventoryDefinition(ctx, "./testdata/basic")
	require.NoError(err)

	err = ts.inventory.UpdateInventoryDefinition(ctx, "./testdata/basic")
	require.NoError(err)
}

func (ts *definitionTestSuite) TestUpdateInventoryDefinitionSimple() {
	require := ts.Require()

	ctx := context.Background()

	err := ts.inventory.UpdateInventoryDefinition(ctx, "./testdata/simple")
	require.NoError(err)

	err = ts.inventory.UpdateInventoryDefinition(ctx, "./testdata/simple")
	require.NoError(err)
}

func (ts *definitionTestSuite) TestUpdateInventoryDefinitionChainedNoChange() {
	require := ts.Require()

	regionName := "region1"
	zoneName   := "zone1"

	region, err := ts.inventory.NewRegion(namespace.DefinitionTable, regionName)
	require.NoError(err)
	require.NotNil(region)

	zone, err := region.NewChild(zoneName)
	require.NoError(err)
	require.NotNil(zone)

	ctx := context.Background()

	err = ts.inventory.UpdateInventoryDefinition(ctx, "./testdata/ChainedBase")
	require.NoError(err)

	_, err = region.Read(ctx, ViewDefinition)
	require.NoError(err)

	storeRevision := region.GetRevisionStore()

	// We only have the facility to establish a watchpoint on a zone, so
	// looking for changes is not foolproof.
	//
	watch, err := zone.Watch(ctx)
	require.NoError(err)
	require.NotNil(watch)

	err = ts.inventory.UpdateInventoryDefinition(ctx, "./testdata/ChainedBase")
	require.NoError(err)

	// Since nothing changed, we expect no events.
	//
	select {
	case <-watch.Events:
		require.Fail("Should not get here - expecting no events from channel")

	default:
		// Nothing to do.
	}

	require.NoError(watch.Close(ctx))

	// Also check that the store revision is unchanged.
	//
	_, err = region.Read(ctx, ViewDefinition)
	require.NoError(err)

	require.Equal(storeRevision, region.GetRevisionStore())
}

func (ts *definitionTestSuite) TestUpdateInventoryDefinitionChainedAddBlade() {
	require := ts.Require()

	regionName := "region1"
	zoneName   := "zone1"
	rackName   := "rack1"
	bladeId    := int64(3)

	errBladeNotFound := errors.ErrBladeNotFound{
		Region: regionName,
		Zone:   zoneName,
		Rack:   rackName,
		Blade:  bladeId,
	}

	blade, err := ts.inventory.NewBlade(namespace.DefinitionTable, regionName, zoneName, rackName, bladeId)
	require.NoError(err)
	require.NotNil(blade)

	ctx := context.Background()

	err = ts.inventory.UpdateInventoryDefinition(ctx, "./testdata/ChainedBase")
	require.NoError(err)

	// Check for the absence of the blade that is about to be added
	//
	_, err = blade.Read(ctx, ViewDefinition)
	require.ErrorIs(err, errBladeNotFound)

	err = ts.inventory.UpdateInventoryDefinition(ctx, "./testdata/ChainedAddBlade")
	require.NoError(err)

	// Check that the blade has been properly loaded
	//
	_, err = blade.Read(ctx, ViewDefinition)
	require.NoError(err)


	err = ts.inventory.UpdateInventoryDefinition(ctx, "./testdata/ChainedBase")
	require.NoError(err)

	// And finally, check that the blade has now been removed.
	//
	_, err = blade.Read(ctx, ViewDefinition)
	require.ErrorIs(err, errBladeNotFound)
}

func (ts *definitionTestSuite) TestUpdateInventoryDefinitionChangeBlade() {
	require := ts.Require()

	regionName := "region1"
	zoneName   := "zone1"
	rackName   := "rack1"
	bladeId    := int64(1)

	blade, err := ts.inventory.NewBlade(namespace.DefinitionTable, regionName, zoneName, rackName, bladeId)
	require.NoError(err)
	require.NotNil(blade)

	ctx := context.Background()

	err = ts.inventory.UpdateInventoryDefinition(ctx, "./testdata/ChainedBase")
	require.NoError(err)

	// Check for the original state of the blade that is about to be modified
	//
	_, err = blade.Read(ctx, ViewDefinition)
	require.NoError(err)

	d := blade.GetDetails()

	require.Equal(pb.Condition_operational, d.Condition)

	// Now load the updated inventory
	//
	err = ts.inventory.UpdateInventoryDefinition(ctx, "./testdata/ChainedChangeBlade")
	require.NoError(err)

	// Check for the expected modified state of the blade that should have been updated
	//
	_, err = blade.Read(ctx, ViewDefinition)
	require.NoError(err)

	d = blade.GetDetails()

	require.Equal(pb.Condition_not_in_service, d.Condition)

	// Then change the blade back to its original condition and verify that.
	//
	err = ts.inventory.UpdateInventoryDefinition(ctx, "./testdata/ChainedBase")
	require.NoError(err)

	_, err = blade.Read(ctx, ViewDefinition)
	require.NoError(err)

	d = blade.GetDetails()

	require.Equal(pb.Condition_operational, d.Condition)
}

func (ts *definitionTestSuite) TestUpdateInventoryDefinitionAddRack() {
	require := ts.Require()

	regionName := "region1"
	zoneName   := "zone1"
	rackName   := "rack2"

	errRackNotFound := errors.ErrRackNotFound{
		Region: regionName,
		Zone:   zoneName,
		Rack:   rackName,
	}

	rack, err := ts.inventory.NewRack(namespace.DefinitionTable, regionName, zoneName, rackName)
	require.NoError(err)
	require.NotNil(rack)

	ctx := context.Background()

	err = ts.inventory.UpdateInventoryDefinition(ctx, "./testdata/ChainedBase")
	require.NoError(err)

	// Check for the absence of the blade that is about to be added
	//
	_, err = rack.Read(ctx, ViewDefinition)
	require.ErrorIs(err, errRackNotFound)

	err = ts.inventory.UpdateInventoryDefinition(ctx, "./testdata/ChainedAddRack")
	require.NoError(err)

	// Check that the rack has been properly loaded
	//
	_, err = rack.Read(ctx, ViewDefinition)
	require.NoError(err)

	err = ts.inventory.UpdateInventoryDefinition(ctx, "./testdata/ChainedBase")
	require.NoError(err)

	// And finally, check that the rack has now been removed.
	//
	_, err = rack.Read(ctx, ViewDefinition)
	require.ErrorIs(err, errRackNotFound)
}

func (ts *definitionTestSuite) TestUpdateInventoryDefinitionAddZone() {
	require := ts.Require()

	regionName := "region1"
	zoneName   := "zone2"

	errZoneNotFound := errors.ErrZoneNotFound{
		Region: regionName,
		Zone:   zoneName,
	}

	zone, err := ts.inventory.NewZone(namespace.DefinitionTable, regionName, zoneName)
	require.NoError(err)
	require.NotNil(zone)

	ctx := context.Background()

	err = ts.inventory.UpdateInventoryDefinition(ctx, "./testdata/ChainedBase")
	require.NoError(err)

	// Check for the absence of the zone that is about to be added
	//
	_, err = zone.Read(ctx, ViewDefinition)
	require.ErrorIs(err, errZoneNotFound)

	err = ts.inventory.UpdateInventoryDefinition(ctx, "./testdata/ChainedAddZone")
	require.NoError(err)

	// Check that the zone has been properly loaded
	//
	_, err = zone.Read(ctx, ViewDefinition)
	require.NoError(err)

	err = ts.inventory.UpdateInventoryDefinition(ctx, "./testdata/ChainedBase")
	require.NoError(err)

	// And finally, check that the zone has now been removed.
	//
	_, err = zone.Read(ctx, ViewDefinition)
	require.ErrorIs(err, errZoneNotFound)
}

func (ts *definitionTestSuite) TestUpdateInventoryDefinitionAddRegion() {
	require := ts.Require()

	regionName := "region2"

	errRegionNotFound := errors.ErrRegionNotFound{
		Region: regionName,
	}

	region, err := ts.inventory.NewRegion(namespace.DefinitionTable, regionName)
	require.NoError(err)
	require.NotNil(region)

	ctx := context.Background()

	err = ts.inventory.UpdateInventoryDefinition(ctx, "./testdata/ChainedBase")
	require.NoError(err)

	// Check for the absence of the zone that is about to be added
	//
	_, err = region.Read(ctx, ViewDefinition)
	require.ErrorIs(err, errRegionNotFound)

	err = ts.inventory.UpdateInventoryDefinition(ctx, "./testdata/ChainedAddRegion")
	require.NoError(err)

	// Check that the zone has been properly loaded
	//
	_, err = region.Read(ctx, ViewDefinition)
	require.NoError(err)

	err = ts.inventory.UpdateInventoryDefinition(ctx, "./testdata/ChainedBase")
	require.NoError(err)

	// And finally, check that the zone has now been removed.
	//
	_, err = region.Read(ctx, ViewDefinition)
	require.ErrorIs(err, errRegionNotFound)
}

func (ts *definitionTestSuite) TestDeleteInventoryDefinitionBasic() {
	require := ts.Require()

	regionName := "region1"
	zoneName := "zone1"
	rackName := "rack1"
	torId := int64(0)

	errTorNotFound := errors.ErrTorNotFound{Region: regionName, Zone: zoneName, Rack: rackName, Tor: torId}

	tor, err := ts.inventory.NewTor(namespace.DefinitionTable, regionName, zoneName, rackName, torId)
	require.NoError(err)
	require.NotNil(tor)

	ctx := context.Background()

	// Verify the sample tor is not currently present in the inventory.
	//
	_, err = tor.Read(ctx, ViewDefinition)
	require.ErrorIs(err, errTorNotFound)

	err = ts.inventory.UpdateInventoryDefinition(ctx, "./testdata/basic")
	require.NoError(err)

	// Verify the sample tor is now present in the inventory.
	//
	_, err = tor.Read(ctx, ViewDefinition)
	require.NoError(err)

	err = ts.inventory.DeleteInventoryDefinition(ctx)
	require.NoError(err)

	// Verify the sample tor has been removed from the inventory.
	//
	_, err = tor.Read(ctx, ViewDefinition)
	require.ErrorIs(err, errTorNotFound)
}

func (ts *definitionTestSuite) TestDeleteInventoryDefinitionSimple() {
	require := ts.Require()

	regionName := "region1"
	zoneName := "zone1"
	rackName := "rack2"
	bladeId := int64(2)

	errBladeNotFound := errors.ErrBladeNotFound{Region: regionName, Zone: zoneName, Rack: rackName, Blade: bladeId}

	blade, err := ts.inventory.NewBlade(namespace.DefinitionTable, regionName, zoneName, rackName, bladeId)
	require.NoError(err)
	require.NotNil(blade)

	ctx := context.Background()

	// Verify the sample blade is not currently present in the inventory.
	//
	_, err = blade.Read(ctx, ViewDefinition)
	require.ErrorIs(err, errBladeNotFound)

	err = ts.inventory.UpdateInventoryDefinition(ctx, "./testdata/simple")
	require.NoError(err)

	// Verify the sample blade is now present in the inventory.
	//
	_, err = blade.Read(ctx, ViewDefinition)
	require.NoError(err)

	err = ts.inventory.DeleteInventoryDefinition(ctx)
	require.NoError(err)

	// Verify the sample blade has been removed from the inventory.
	//
	_, err = blade.Read(ctx, ViewDefinition)
	require.ErrorIs(err, errBladeNotFound)
}

func TestDefinitionTestSuite(t *testing.T) {
	suite.Run(t, new(definitionTestSuite))
}
