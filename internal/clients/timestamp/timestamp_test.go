package timestamp

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/Jim3Things/CloudChamber/internal/common"
	ct "github.com/Jim3Things/CloudChamber/pkg/protos/common"
)

type timestampTestSuite struct {
	stepperTestClientSuite
}

func (ts *timestampTestSuite) TestNow() {
	require := ts.Require()
	assert := ts.Assert()

	ctx := context.Background()

	now, err := Now(ctx)
	require.Nil(err)
	assert.Equal(int64(0), now.Ticks)

	now, err = Now(ctx)
	require.Nil(err)
	assert.Equal(int64(0), now.Ticks)

	assert.Nil(Advance(ctx))
	now, err = Now(ctx)
	require.Nil(err)
	assert.Equal(int64(1), now.Ticks)

	now, err = Now(ctx)
	require.Nil(err)
	assert.Equal(int64(1), now.Ticks)
}

func (ts *timestampTestSuite) TestTimestamp_After() {
	require := ts.Require()
	assert := ts.Assert()

	ctx := context.Background()

	now, err := Now(ctx)
	require.Nil(err)
	assert.Equal(int64(0), now.Ticks)

	ch := make(chan bool)

	go func(deadline int64, res chan<- bool) {
		data := <-After(ctx, &ct.Timestamp{Ticks: deadline})

		require.Nil(data.Err)
		assert.GreaterOrEqual(deadline, data.Time.Ticks)
		res <- true
	}(3, ch)

	require.Nil(Advance(ctx))
	assert.True(common.DoNotCompleteWithin(ch, time.Duration(2)*time.Second))

	require.Nil(Advance(ctx))
	assert.True(common.DoNotCompleteWithin(ch, time.Duration(2)*time.Second))

	require.Nil(Advance(ctx))
	assert.True(common.CompleteWithin(ch, time.Duration(2)*time.Second))
}


func TestTimestampTestSuite(t *testing.T) {
	suite.Run(t, new(timestampTestSuite))
}
