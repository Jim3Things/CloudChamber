// Test to parse the Inventory.Yaml file
package config

import (
	"os"
	"testing"

	"github.com/spf13/viper"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Jim3Things/CloudChamber/internal/tracing/exporters"
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
	exporters.Init(utf)
}

// first inventory definition test

func TestReadInventoryDefinition (t *testing.T) {

	_ = utf.Open(t)
	defer utf.Close()

	response, err := ReadInventoryDefinition(".")
	require.Nil(t, err)

	require.Equal(t, 2, len(response.Racks))

	r,ok:= response.Racks["rack1"]
	require.True(t, ok)
	assert.Equal(t, 2, len(r.Blades))

	b,ok:=r.Blades[1]
	require.True(t, ok)
	assert.Equal(t, int64(16), b.Cores)
	assert.Equal(t, int64(16834), b.MemoryInMb)
	assert.Equal(t, int64(240), b.DiskInGb)
	assert.Equal(t, int64(2048), b.NetworkBandwidthInMbps)
	assert.Equal(t, "X64", b.Arch)

	s,ok:=response.Racks["rack2"]
	require.True(t, ok)
	assert.Equal(t, 2, len(s.Blades))

	c,ok:=r.Blades[2]
	require.True(t, ok)
	assert.Equal(t, int64(8), c.Cores)
	assert.Equal(t, int64(16834), c.MemoryInMb)
	assert.Equal(t, int64(120), c.DiskInGb)
	assert.Equal(t, int64(2048), c.NetworkBandwidthInMbps)
}

func TestReadInventoryBogusPath (t *testing.T){
	_ = utf.Open(t)
	defer utf.Close()
	viper.Reset()

	response, err := ReadInventoryDefinition("C://Users//Waheguru")
	require.NotNil(t, err)
	assert.NotEqual(t, "%v", response)
}

//TestInventoryUniqueRack test to check that zone always contain unique rack numbers
func TestInventoryUniquRack (t *testing.T) {

	_ = utf.Open(t)
	defer utf.Close()
	viper.Reset()

	_, err := ReadInventoryDefinition(".//BadYaml")
	require.NotNil(t, err)
	assert.Equal (t, "Duplicate rack \"rack1\" detected", err.Error() )
}

func TestInventoryUniqueBlade (t *testing.T){

	_ = utf.Open(t)
	defer utf.Close()
	viper.Reset()

	_, err := ReadInventoryDefinition(".//BadYamlBlade")
	require.NotNil(t, err)
	assert.Equal (t, "Duplicate Blade 1 in Rack \"rack1\" detected", err.Error() )
}

func TestInventoryValidateBlade (t *testing.T){

	_ = utf.Open(t)
	defer utf.Close()
	viper.Reset()

	_, err := ReadInventoryDefinition(".//BadYamlValidate")
	require.NotNil(t, err)
	assert.Equal (t,  "In rack \"rack1\": the field \"Blades[2].Cores\" must be greater than or equal to 1.  It is 0, which is invalid", err.Error() )
}
