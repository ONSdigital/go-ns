package healthcheck

import (
	"github.com/ONSdigital/go-ns/log"
	"time"
)

// Ticker calls the health check monitor function at configured intervals.
type Ticker struct {
	timeTicker *time.Ticker
}

// NewTicker returns a new ticker that checks the given clients at intervals of the given duration.
func NewTicker(duration time.Duration, clients ...Client) *Ticker {

	timeTicker := time.NewTicker(duration)

	go func() {

		for range timeTicker.C {

			log.Debug("conducting service healthcheck", nil)
			MonitorExternal(clients...)
		}
	}()

	return &Ticker{
		timeTicker: timeTicker,
	}
}

// Close the ticker to exit its internal goroutine
func (ticker *Ticker) Close() {
	ticker.timeTicker.Stop()
}
