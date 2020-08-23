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


	r,ok:= response.Racks["rack1"]
	require.True(t, ok)
	assert.Equal(t, 2, len(r.Blades))
	
	b,ok:=r.Blades[1]
	require.True(t, ok)
	assert.Equal(t, int64(16), b.Cores )
	


		
}


