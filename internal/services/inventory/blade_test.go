package inventory

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	tsc "github.com/Jim3Things/CloudChamber/internal/clients/timestamp"
	"github.com/Jim3Things/CloudChamber/internal/common"
	"github.com/Jim3Things/CloudChamber/internal/sm"
	"github.com/Jim3Things/CloudChamber/internal/tracing"
)

type BladeTestSuite struct {
	testSuiteCore
}

func (ts *BladeTestSuite) TestPowerOn() {
	require := ts.Require()
	assert := ts.Assert()

	rackDef := ts.createDummyRack(2)

	r := newRack(context.Background(), ts.rackName(), rackDef)
	require.NotNil(r)

	ctx := ts.advance(context.Background())

	ctx, span := tracing.StartSpan(
		ctx,
		tracing.WithName("test powering off a PDU"),
		tracing.WithContextValue(tsc.EnsureTickInContext))

	rsp := make(chan *sm.Response)

	msg := newSetPower(ctx, newTargetBlade(ts.rackName(), 0), common.TickFromContext(ctx), true, rsp)

	span.End()

	ts.execute(msg, r.blades[0].Receive)

	res := ts.completeWithin(rsp, time.Duration(1)*time.Second)
	require.NotNil(res)
	require.NoError(res.Err)
	assert.Equal(common.TickFromContext(ctx), res.At)
	assert.Nil(res.Msg)

	assert.Equal("booting", r.blades[0].sm.Current.Name())
}

func TestBladeSuite(t *testing.T) {
	suite.Run(t, new(BladeTestSuite))
}
