// Test to parse the Inventory.Yaml file
package config

import (
	"os"
	"testing"

	"github.com/Jim3Things/CloudChamber/internal/tracing/exporters"
	"github.com/Jim3Things/CloudChamber/internal/tracing/exporters/unit_test"
	"github.com/Jim3Things/CloudChamber/internal/tracing/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	setup.Init(exporters.UnitTest)
}

// first inventory definition test

func TestReadInventoryDefinition (t *testing.T) {

	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	response, err := ReadInventoryDefinition(".")
	require.Nil(t, err)

	require.Equal(t, 2, len(response.Racks))

// Racks is a map of a maps (A struct or array type). for eg: Rack1 is an map of blade 1 and blade2. blade 1 itself is a map.
	
	r,ok:= response.Racks["rack1"] // reading array value of  map Racks that was captured in response into a new variable r. 
	require.True(t, ok) // Continue only if the boolean value is True. i.e rack1 was found.
	assert.Equal(t, 2, len(r.Blades)) //Confirming if the length of array item r.Blades is 2. i.e there are two blades in rack1.
	
	b,ok:=r.Blades[1] //capturing key blade 1 into the variable b. The second value is a boolean which indicates if the key was successfully found  or not.
	require.True(t, ok)//continue only if the boolean value is True. i.e Blade 1 was found. 
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
	unit_test.SetTesting(t)
	defer unit_test.SetTesting(nil)

	response, err := ReadInventoryDefinition("C://Users//Waheguru") //Points to the user directory that do not have the YAML File
	require.NotNil(t, err) //Checks that the error is not empty. IF it is than it should throw back an error
	assert.NotEqual(t, "%v", response)//This is just so that I can use variable response. No actual impact on the test.
}