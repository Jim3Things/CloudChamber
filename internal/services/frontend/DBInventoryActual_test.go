package main

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestInitDBInventoryActual(t *testing.T){

	response, err := InitDBInventoryActual(inven *DBInventory)
	require.NotNil(t, err)
	r, ok := actual.Zone.Racks["rack1"]
	require.True(t, ok)
	assert.Equal(t, 8, len(r.Blades))

}