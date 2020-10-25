package inventory

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Jim3Things/CloudChamber/internal/common"
	"github.com/Jim3Things/CloudChamber/internal/tracing"
)

type RackTestSuite struct {
	testSuiteCore
}

func (ts *RackTestSuite) TestCreateRack() {
	require := ts.Require()
	assert := ts.Assert()

	ctx := common.ContextWithTick(context.Background(), 1)

	rackDef := ts.createDummyRack(2)
	ctx, span := tracing.StartSpan(ctx, tracing.WithName("test rack creation"))
	r := newRack(ctx, ts.rackName(), rackDef)
	span.End()

	require.NotNil(r)
	assert.Equal(len(rackDef.Blades), len(r.blades))
}

func (ts *RackTestSuite) TestStartStopRack() {
	require := ts.Require()
	assert := ts.Assert()

	ctx := common.ContextWithTick(context.Background(), 1)

	rackDef := ts.createDummyRack(2)

	ctx, span := tracing.StartSpan(
		ctx,
		tracing.WithName("test rack start and stop"))

	r := newRack(ctx, ts.rackName(), rackDef)
	require.NotNil(r)
	assert.Equal(len(rackDef.Blades), len(r.blades))
	assert.Equal(rackAwaitingStartState, r.sm.CurrentIndex)

	err := r.start(ctx)
	assert.NoError(err)

	assert.Equal(rackWorkingState, r.sm.CurrentIndex)

	r.stop(ctx)
	assert.Equal(rackTerminalState, r.sm.CurrentIndex)

	span.End()
}

func TestRackTestSuite(t *testing.T) {
	suite.Run(t, new(RackTestSuite))
}
