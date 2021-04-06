package inventory

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
)

type definitionExtendedTestSuite struct {
	testSuiteCore

	inventory      *Inventory

	// regionCount    int
	// zonesPerRegion int
	// racksPerZone   int
	// pdusPerRack    int
	// torsPerRack    int
	// bladesPerRack  int

	// portsPerPdu int
	// portsPerTor int
}

func (ts *definitionExtendedTestSuite) SetupSuite() {
	// require := ts.Require()

	// ctx := context.Background()

	ts.testSuiteCore.SetupSuite()

	ts.inventory = NewInventory(ts.cfg, ts.store)

	// // These values are relatively arbitrary. The only criteria is that different
	// // constants were chosen to help separate different multiples of different
	// // object types where possible and not to have values which are too large
	// // to avoid lots of IO when setting up the test suite.
	// //
	// ts.regionCount = 2
	// ts.zonesPerRegion = 3
	// ts.racksPerZone = 4
	// ts.pdusPerRack = 1
	// ts.torsPerRack = 1
	// ts.bladesPerRack = 5

	// ts.portsPerPdu = ts.torsPerRack + ts.bladesPerRack
	// ts.portsPerTor = ts.pdusPerRack + ts.torsPerRack + ts.bladesPerRack

	// require.NoError(ts.utf.Open(ts.T()))
	// require.NoError(ts.store.Connect())

	// err := ts.createStandardInventory(ctx)
	// require.NoError(err, "failed to create standard inventory")

	// ts.store.Disconnect()
	// ts.utf.Close()
}

func (ts *definitionExtendedTestSuite) SetupTest() {
	require := ts.Require()

	require.NoError(ts.utf.Open(ts.T()))

	require.NoError(ts.store.Connect())
}

func (ts *definitionExtendedTestSuite) TearDownTest() {
	ts.store.Disconnect()
	ts.utf.Close()
}

func (ts *definitionExtendedTestSuite) TestReadInventoryDefinitionFromFileExtended() {
	require := ts.Require()

	_, err := ReadInventoryDefinitionFromFile(context.Background(), "./testdata/extended")
	require.NoError(err)
}

func (ts *definitionExtendedTestSuite) TestUpdateInventoryDefinitionExtended() {
	require := ts.Require()

	ctx := context.Background()

	err := ts.inventory.UpdateInventoryDefinition(ctx, "./testdata/extended")
	require.NoError(err)

	err = ts.inventory.UpdateInventoryDefinition(ctx, "./testdata/extended")
	require.NoError(err)
}

func (ts *definitionExtendedTestSuite) TestWriteInventoryDefinitionExtended() {
	require := ts.Require()

	ctx := context.Background()

	root, err := ts.inventory.readInventoryDefinitionFromStore(ctx)
	require.NoError(err)

	err = ts.inventory.deleteInventoryDefinitionFromStore(ctx, root)
	require.NoError(err)

	root, err = ReadInventoryDefinitionFromFile(ctx, "./testdata/extended")
	require.NoError(err)
	require.NotNil(root)

	err = ts.inventory.writeInventoryDefinitionToStore(ctx, root)
	require.NoError(err)

	rootReload, err := ts.inventory.readInventoryDefinitionFromStore(ctx)
	require.NoError(err)
	require.NotNil(rootReload)

	err = ts.inventory.deleteInventoryDefinitionFromStore(ctx, rootReload)
	require.NoError(err)
}

func (ts *definitionExtendedTestSuite) TestDeleteInventoryDefinitionExtended() {
	require := ts.Require()

	ctx := context.Background()

	err := ts.inventory.UpdateInventoryDefinition(ctx, "./testdata/extended")
	require.NoError(err)

	err = ts.inventory.DeleteInventoryDefinition(ctx)
	require.NoError(err)
}


func TestDefinitionExtendedTestSuite(t *testing.T) {
	suite.Run(t, new(definitionExtendedTestSuite))
}
