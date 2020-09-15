package common

import (
	"context"
)

type tickKeyType struct {}
var tickKey = tickKeyType{}

// ContextWithTick returns a new context with the current simulated time added.
func ContextWithTick(ctx context.Context, tick int64) context.Context {
	return context.WithValue(ctx, tickKey, tick)
}

// TickFromContext extracts the simulated time from the supplied context.
func TickFromContext(ctx context.Context) int64 {
	if tick, ok := ctx.Value(tickKey).(int64); ok {
		return tick
	}

	return -1
}
