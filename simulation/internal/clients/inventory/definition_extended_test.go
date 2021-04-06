package inventory

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
)

type definitionExtendedTestSuite struct {
	testSuiteCore

	inventory      *Inventory
}

func (ts *definitionExtendedTestSuite) SetupSuite() {
	ts.testSuiteCore.SetupSuite()

	ts.inventory = NewInventory(ts.cfg, ts.store)
}

func (ts *definitionExtendedTestSuite) SetupTest() {
	ts.testSuiteCore.SetupTest()
}

func (ts *definitionExtendedTestSuite) TearDownTest() {
	ts.testSuiteCore.TearDownTest()
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
