// Test to parse the Inventory.Yaml file
package inventory

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/spf13/viper"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Jim3Things/CloudChamber/internal/tracing/exporters"
	pb "github.com/Jim3Things/CloudChamber/pkg/protos/inventory"
)

var (
	utf *exporters.Exporter
)

// Common test startup method.  This is the _only_ Test* function in this
// file.
func TestMain(m *testing.M) {
	commonSetup()

	os.Exit(m.Run())
}

// Establish the test environment, including starting a test frontend service
// over a faked http connection.
func commonSetup() {
	utf = exporters.NewExporter(exporters.NewUTForwarder())
	exporters.ConnectToProvider(utf)
}

// first inventory definition test

func TestReadInventoryDefinition(t *testing.T) {

	_ = utf.Open(t)
	defer utf.Close()

	response, err := ReadInventoryDefinition(context.Background(), "./testdata/Basic")
	require.Nil(t, err)

	require.Equal(t, 2, len(response.Racks))

	r, ok := response.Racks["rack1"]
	require.True(t, ok)
	assert.Equal(t, 2, len(r.Blades))

	b, ok := r.Blades[1]
	require.True(t, ok)
	assert.Equal(t, int64(16), b.Cores)
	assert.Equal(t, int64(16834), b.MemoryInMb)
	assert.Equal(t, int64(240), b.DiskInGb)
	assert.Equal(t, int64(2048), b.NetworkBandwidthInMbps)
	assert.Equal(t, "X64", b.Arch)

	s, ok := response.Racks["rack2"]
	require.True(t, ok)
	assert.Equal(t, 2, len(s.Blades))

	c, ok := r.Blades[2]
	require.True(t, ok)
	assert.Equal(t, int64(8), c.Cores)
	assert.Equal(t, int64(16834), c.MemoryInMb)
	assert.Equal(t, int64(120), c.DiskInGb)
	assert.Equal(t, int64(2048), c.NetworkBandwidthInMbps)
}

func TestReadInventoryBogusPath(t *testing.T) {
	_ = utf.Open(t)
	defer utf.Close()
	viper.Reset()

	response, err := ReadInventoryDefinition(context.Background(), "./missing/path")
	require.NotNil(t, err)
	assert.NotEqual(t, "%v", response)
}

// TestInventoryUniqueRack test to check that zone always contain unique rack numbers
func TestInventoryUniqueRack(t *testing.T) {

	_ = utf.Open(t)
	defer utf.Close()
	viper.Reset()

	_, err := ReadInventoryDefinition(context.Background(), "./testdata/BadYaml")
	require.NotNil(t, err)
	assert.Equal(t, "Duplicate rack \"rack1\" detected", err.Error())
}

func TestInventoryUniqueBlade(t *testing.T) {

	_ = utf.Open(t)
	defer utf.Close()
	viper.Reset()

	_, err := ReadInventoryDefinition(context.Background(), "./testdata/BadYamlBlade")
	require.NotNil(t, err)
	assert.Equal(t, "Duplicate Blade 1 in Rack \"rack1\" detected", err.Error())
}

func TestInventoryValidateBlade(t *testing.T) {

	_ = utf.Open(t)
	defer utf.Close()
	viper.Reset()

	_, err := ReadInventoryDefinition(context.Background(), "./testdata/BadYamlValidate")
	require.NotNil(t, err)
	assert.Equal(t,
		"In rack \"rack1\": the field \"Blades[2].Cores\" must be greater than or equal to 1.  It is 0, which is invalid",
		err.Error())
}

func TestReadInventoryDefinitionFromFile(t *testing.T) {

	_ = utf.Open(t)
	defer utf.Close()
	viper.Reset()

	zonemap, err := ReadInventoryDefinitionFromFile(context.Background(), "./testdata/Basic")
	require.NoError(t, err)

	// There should only be a single zone.
	//
	require.Equal(t, 1, len(zonemap.Zones))

	zone, ok := zonemap.Zones[DefaultZone]
	require.True(t, ok)

	assert.True(t, zone.Details.Enabled)
	assert.Equal(t, pb.State_in_service, zone.Details.State)
	assert.Equal(t, "DC-PNW-0", zone.Details.Location)
	assert.Equal(t, "Base zone", zone.Details.Notes)

	require.Equal(t, 2, len(zone.Racks))

	for i := 1; i <= 2; i++ {

		name := fmt.Sprintf("rack%d", i)
		
		r, ok := zone.Racks[name]
		require.True(t, ok)

		assert.True(t, r.Details.Enabled)
		assert.Equal(t, pb.Condition_operational, r.Details.Condition)
		assert.Equal(t, "DC-PNW-0-" + name, r.Details.Location)
		assert.Equal(t, "RackName: " + name, r.Details.Notes)

		assert.Equal(t, 1, len(r.Pdus))
		assert.Equal(t, 1, len(r.Tors))
		assert.Equal(t, 2, len(r.Blades))

		// There should be a single PDU at index 0
		//
		p0, ok := r.Pdus[0]
		require.True(t, ok)

		// The PDU should have a wired port for each of the two expected blades.
		//
		assert.Equal(t, 2, len(p0.Ports))

		p0b1, ok := p0.Ports[1]
		require.True(t, ok)

		assert.True(t, p0b1.Wired)
		assert.Equal(t, pb.Hardware_blade, p0b1.Item.Type)
		assert.Equal(t, int64(1), p0b1.Item.Id)
		assert.Equal(t, int64(0), p0b1.Item.Port)

		p0b2, ok := p0.Ports[2]
		require.True(t, ok)

		assert.True(t, p0b2.Wired)
		assert.Equal(t, pb.Hardware_blade, p0b2.Item.Type)
		assert.Equal(t, int64(2), p0b2.Item.Id)
		assert.Equal(t, int64(0), p0b2.Item.Port)

		// There should be a single TOR at index 0
		//
		t0, ok := r.Tors[0]
		require.True(t, ok)

		// The TOR should have a wired port for each of the two expected blades.
		//
		assert.Equal(t, 2, len(t0.Ports))

		t0b1, ok := t0.Ports[1]
		require.True(t, ok)

		assert.True(t, t0b1.Wired)
		assert.Equal(t, pb.Hardware_blade, t0b1.Item.Type)
		assert.Equal(t, int64(1), t0b1.Item.Id)
		assert.Equal(t, int64(0), t0b1.Item.Port)

		t0b2, ok := p0.Ports[2]
		require.True(t, ok)

		assert.True(t, t0b2.Wired)
		assert.Equal(t, pb.Hardware_blade, t0b2.Item.Type)
		assert.Equal(t, int64(2), t0b2.Item.Id)
		assert.Equal(t, int64(0), t0b2.Item.Port)

		// There should be exactly two blades at indices 1 and 2.
		//
		b1, ok := r.Blades[1]
		require.True(t, ok)

		assert.True(t, b1.Details.Enabled)
		assert.Equal(t, pb.Condition_operational, b1.Details.Condition)

		assert.Equal(t, int64(16),    b1.Capacity.Cores)
		assert.Equal(t, int64(16834), b1.Capacity.MemoryInMb)
		assert.Equal(t, int64(240),   b1.Capacity.DiskInGb)
		assert.Equal(t, int64(2048),  b1.Capacity.NetworkBandwidthInMbps)
		assert.Equal(t, "X64",        b1.Capacity.Arch)

		b2, ok := r.Blades[2]
		require.True(t, ok)

		assert.True(t, b2.Details.Enabled)
		assert.Equal(t, pb.Condition_operational, b2.Details.Condition)

		assert.Equal(t, int64(8),     b2.Capacity.Cores)
		assert.Equal(t, int64(16834), b2.Capacity.MemoryInMb)
		assert.Equal(t, int64(120),   b2.Capacity.DiskInGb)
		assert.Equal(t, int64(2048),  b2.Capacity.NetworkBandwidthInMbps)
		assert.Equal(t, "X64",        b2.Capacity.Arch)
	}
}

func TestReadInventoryDefinitionFromFileBogusPath(t *testing.T) {
	_ = utf.Open(t)
	defer utf.Close()
	viper.Reset()

	response, err := ReadInventoryDefinitionFromFile(context.Background(), "./missing/path")
	require.Error(t, err)
	assert.NotEqual(t, "%v", response)
}

// TestInventoryUniqueRack test to check that zone always contain unique rack numbers
//
func TestIReadInventoryDefinitionFromFileUniqueRack(t *testing.T) {

	_ = utf.Open(t)
	defer utf.Close()
	viper.Reset()

	_, err := ReadInventoryDefinitionFromFile(context.Background(), "./testdata/BadYaml")
	require.Error(t, err)
	assert.Equal(t, "Duplicate rack \"rack1\" detected", err.Error())
}

func TestReadInventoryDefinitionFromFileUniqueBlade(t *testing.T) {

	_ = utf.Open(t)
	defer utf.Close()
	viper.Reset()

	_, err := ReadInventoryDefinitionFromFile(context.Background(), "./testdata/BadYamlBlade")
	require.Error(t, err)
	assert.Equal(t, "Duplicate Blade 1 in Rack \"rack1\" detected", err.Error())
}

func TestReadInventoryDefinitionFromFileValidateBlade(t *testing.T) {

	_ = utf.Open(t)
	defer utf.Close()
	viper.Reset()

	_, err := ReadInventoryDefinitionFromFile(context.Background(), "./testdata/BadYamlValidate")
	require.Error(t, err)
	assert.Equal(t,
		"In rack \"rack1\": the field \"Blades[2].Cores\" must be greater than or equal to 1.  It is 0, which is invalid",
		err.Error())
}
