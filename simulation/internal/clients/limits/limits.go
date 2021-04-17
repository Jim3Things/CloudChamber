// This package is used to provide a small set of limit values
// which are not obtained from the configuration, at least at present.

package limits

import (
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
)

const (
	// MaxBladeID is the highest blade number accepted as valid. This is
	// an arbitrary choice intended to help prevent configuration issues.
	//
	MaxBladeID = int64(10 * 1000 * 1000)

	// MaxPduID defines the larget number of Pdus that can be configured within
	// a single rack.
	//
	MaxPduID = int64(2)

	// MaxTorID defines the larget number of Tors that can be configured within
	// a single rack.
	//
	MaxTorID = int64(2)
)

var (
	// BackoffSettings contains the grpc reconnection parameters - how long to
	// wait to retry, how much to extend the retry interval by, and other
	// related values.
	BackoffSettings = grpc.ConnectParams{
		Backoff: backoff.Config{
			BaseDelay:  100 * time.Millisecond,
			Multiplier: 1.6,
			Jitter:     0.2,
			MaxDelay:   5 * time.Second,
		},
		MinConnectTimeout: 0,
	}
)
