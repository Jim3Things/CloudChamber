// Test to parse the Inventory.Yaml file
package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// first inventory definition test

func ReadInventoryDefinitiontest (t *testing.T) {

	response := ReadInventoryDefinition("C:\\Users\\Waheguru")

	 assert.Equal(t, "C:\\User\\Waheguru", response)
	 
}

