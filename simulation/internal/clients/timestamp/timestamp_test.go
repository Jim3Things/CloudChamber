package timestamp

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/Jim3Things/CloudChamber/simulation/internal/common"
	pb "github.com/Jim3Things/CloudChamber/simulation/pkg/protos/services"
)

type timestampTestSuite struct {
	stepperTestClientSuite
}

func (ts *timestampTestSuite) verifyNow(ctx context.Context, tick int64) {
	require := ts.Require()

	status, err := Status(ctx)
	require.NoError(err)
	require.EqualValues(tick, status.Now)
}

func (ts *timestampTestSuite) TestStatus() {
	require := ts.Require()

	ctx := context.Background()

	ts.verifyNow(ctx, 0)
	ts.verifyNow(ctx, 0)

	require.NoError(Advance(ctx))

	ts.verifyNow(ctx, 1)
	ts.verifyNow(ctx, 1)
}

func (ts *timestampTestSuite) TestTimestamp_After() {
	require := ts.Require()
	assert := ts.Assert()

	ctx := context.Background()

	ts.verifyNow(ctx, 0)

	ch := make(chan bool)

	go func(deadline int64, res chan<- bool) {
		data := <-After(ctx, deadline)

		require.NoError(data.Err)
		assert.GreaterOrEqual(deadline, data.Status.Now)
		res <- true
	}(3, ch)

	require.NoError(Advance(ctx))
	require.NoError(Advance(ctx))
	assert.True(common.DoNotCompleteWithin(ch, 2*time.Second))

	require.NoError(Advance(ctx))
	assert.True(common.CompleteWithin(ch, 2*time.Second))
}

func (ts *timestampTestSuite) TestForcedError() {
	require := ts.Require()
	assert := ts.Assert()

	ctx := context.Background()

	acl := tsc.(*activeClient)
	testErr := errors.New("bogus")

	client, err := acl.dial()
	require.NoError(err)
	require.NotNil(client)

	// Test serial cleanup & reuse
	require.Equal(testErr, acl.cleanup(client, testErr))

	ts.verifyNow(ctx, 0)

	// Test overlapping cleanup and reuse
	client2, err := acl.dial()
	require.NoError(err)
	require.NotNil(client2)
	require.NotEqual(client, client2)

	require.Equal(testErr, acl.cleanup(client, testErr))

	cts, err := client2.GetStatus(ctx, &pb.GetStatusRequest{})
	require.NoError(err)

	_, err = Status(ctx)
	require.NoError(err)
	assert.Equal(int64(0), cts.Now)
}

func TestTimestampTestSuite(t *testing.T) {
	suite.Run(t, new(timestampTestSuite))
}
