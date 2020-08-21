package common

import (
	"context"
	"errors"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Jim3Things/CloudChamber/pkg/protos/log"
)

func TestSimple(t *testing.T) {
	queue := NewDeferrable(10)

	assert.Equal(t, 10, queue.GetLimit())
	assert.Equal(t, 0, queue.GetCount())
}

func TestEmptyFlush(t *testing.T) {
	queue := NewDeferrable(10)

	err := queue.Flush(context.Background(), func(ctx context.Context, entry *log.Entry) error {
		assert.Fail(t, "was called to flush with an empty queue")

		return nil
	})

	assert.Nil(t, err)
}

func TestSimpleDefer(t *testing.T) {
	var entries = []*log.Entry{
		{
			Name:       "1",
			SpanID:     "0000",
			ParentID:   "0000",
			Status:     "ok",
			StackTrace: "xxx",
			Event:      nil,
		},
		{
			Name:       "2",
			SpanID:     "0000",
			ParentID:   "0000",
			Status:     "ok",
			StackTrace: "xxx",
			Event:      nil,
		}}

	queue := NewDeferrable(10)

	err := queue.Defer(entries[0])

	assert.Nil(t, err)
	assert.Equal(t, 1, queue.GetCount())

	err = queue.Defer(entries[1])

	assert.Nil(t, err)
	assert.Equal(t, 2, queue.GetCount())

	countHit := 0

	err = queue.Flush(context.Background(), func(ctx context.Context, item *log.Entry) error {
		assert.Same(t, entries[countHit], item)
		countHit++

		return nil
	})

	assert.Nil(t, err)
	assert.Equal(t, 2, countHit)
}

func TestDeferLimitExceeded(t *testing.T) {
	queue := NewDeferrable(10)

	index := 0

	for i := 0; i < 12; i++ {
		if err := queue.Defer(&log.Entry{
			Name:       strconv.Itoa(index),
			SpanID:     "0000",
			ParentID:   "0000",
			Status:     "ok",
			StackTrace: "",
			Event:      nil,
		}); err != nil {
			break
		}

		index++
	}

	assert.Equal(t, 10, index)
	assert.Equal(t, 10, queue.GetCount())

	index = 0
	err := queue.Flush(context.Background(), func(ctx context.Context, item *log.Entry) error {
		index++
		return nil
	})

	assert.Nil(t, err)
	assert.Equal(t, 10, index)
	assert.Equal(t, 0, queue.GetCount())
}

func TestFlushError(t *testing.T) {
	var entries = []*log.Entry{
		{
			Name:       "1",
			SpanID:     "0000",
			ParentID:   "0000",
			Status:     "ok",
			StackTrace: "xxx",
			Event:      nil,
		},
		{
			Name:       "2",
			SpanID:     "0000",
			ParentID:   "0000",
			Status:     "ok",
			StackTrace: "xxx",
			Event:      nil,
		},
		{
			Name:       "3",
			SpanID:     "0000",
			ParentID:   "0000",
			Status:     "ok",
			StackTrace: "xxx",
			Event:      nil,
		},
		{
			Name:       "4",
			SpanID:     "0000",
			ParentID:   "0000",
			Status:     "ok",
			StackTrace: "xxx",
			Event:      nil,
		},
		{
			Name:       "5",
			SpanID:     "0000",
			ParentID:   "0000",
			Status:     "ok",
			StackTrace: "xxx",
			Event:      nil,
		},
	}

	queue := NewDeferrable(10)

	for _, item := range entries {
		err := queue.Defer(item)
		assert.Nil(t, err)
	}

	failError := errors.New("forced failure")
	err := queue.Flush(context.Background(), func(ctx context.Context, item *log.Entry) error {
		if item.Name == "3" {
			return failError
		}

		return nil
	})

	assert.Equal(t, failError, err)
	assert.Equal(t, 3, queue.GetCount())

	// Now, try again.  This should start with the failed flush entry (third in the array)
	index := 2

	err = queue.Flush(context.Background(), func(ctx context.Context, item *log.Entry) error {
		assert.Same(t, entries[index], item)
		index++

		return nil
	})

	assert.Nil(t, err)
	assert.Equal(t, 0, queue.GetCount())
}
