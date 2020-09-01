package common

import (
	"github.com/Jim3Things/CloudChamber/internal/clients/timestamp"
)

// Tick provides the current simulated time Tick, or '-1' if the simulated time
// cannot be retrieved (e.g. during startup)
func Tick() int64 {
	now, err := clients.Now()
	if err != nil {
		return -1
	}

	return now.Ticks
}