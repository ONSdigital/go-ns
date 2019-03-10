package healthcheck

import (
	"sync"
	"time"

	"github.com/ONSdigital/log.go/log"
)

// Ticker calls the health check monitor function at configured intervals.
type Ticker struct {
	timeTicker *time.Ticker
	asyncCheck chan bool
	closing    chan bool
	closed     chan bool
}

// NewTicker returns a new ticker that checks the given clients at intervals of the given duration
func NewTicker(duration, recoveryDuration time.Duration, clients ...Client) *Ticker {
	return NewTickerWithAlerts(duration, recoveryDuration, nil, nil, clients...)
}

// NewTickerWithAlerts returns a new ticker that checks the given clients at intervals
// - intervals vary in length (when healthy, use `duration`; at startup, or unhealthy state, use `recoveryDuration`)
// - sends alerts of any change in health via the given channel (if non-nil) - with the new state (`true`:newly-healthy, `false`:newly-unhealthy)
// - healthchecks can be requested using `requestCheckChan`
//   - sending `false` on this channel indicates that a failure has prompted this check (fail fast)
//   - sending `true` indicates that a success has happened (unlikely to indicate 100% health)
// - healthchecks will only be run at a maximum of `recoveryDuration` frequency
func NewTickerWithAlerts(
	duration,
	recoveryDuration time.Duration,
	stateChangeChan,
	requestCheckChan chan bool,
	clients ...Client,
) *Ticker {

	ticker := Ticker{
		timeTicker: time.NewTicker(duration),
		closing:    make(chan bool),
		closed:     make(chan bool),
	}

	asyncCheck := make(chan bool)

	// mutexChecking locks the below vars, and also locks MonitorExternal
	var mutexChecking sync.Mutex
	var lastCheckStarted time.Time
	var lastCheckCompleted time.Time

	// mutexCurrentHealthOK locks the below var
	var mutexCurrentHealthOK sync.Mutex
	var currentHealthOK bool

	// create recovery ticks which initiate a healthcheck only on startup and when unhealthy
	go func() {
		for range time.Tick(recoveryDuration) {
			mutexCurrentHealthOK.Lock()
			if !currentHealthOK {
				asyncCheck <- true
			}
			mutexCurrentHealthOK.Unlock()
		}
	}()

	// main goroutine to run MonitorExternal at intervals, and send any health state changes
	go func() {
		defer close(ticker.closed)
		for {
			// block until closing, or a check is due (either ticker chan) or requested
			select {
			case <-ticker.closing:
				return
			case <-asyncCheck:
			case <-ticker.timeTicker.C:
			case claimedNewHealthOK := <-requestCheckChan:
				mutexCurrentHealthOK.Lock()
				logData := log.Data{"prev_health": currentHealthOK, "new_health": claimedNewHealthOK}
				if currentHealthOK != claimedNewHealthOK {
					if !claimedNewHealthOK {
						// fail fast if claimedNewHealthOK is bad
						log.Event(nil, "health state change - fail via request", logData)
						stateChangeChan <- claimedNewHealthOK
						currentHealthOK = claimedNewHealthOK
					} else {
						// recover slowly: only change state after successful MonitorExternal()
						log.Event(nil, "health state change - success claimed - will check", logData)
					}
				} else {
					log.Event(nil, "healthcheck requested - no change in health", logData)
				}
				mutexCurrentHealthOK.Unlock()
			}

			now := time.Now()
			mutexChecking.Lock()

			// too soon for healthcheck?
			if now.Before(lastCheckStarted.Add(recoveryDuration)) || now.Before(lastCheckCompleted.Add(recoveryDuration)) {
				mutexChecking.Unlock()
				log.Event(nil, "too soon for healthcheck", log.Data{"last_check_start": lastCheckStarted, "last_check_ended": lastCheckCompleted})

			} else {

				lastCheckStarted = now

				// run check in the background,  mutexChecking.Lock() applies
				go func() {

					log.Event(nil, "conducting service healthcheck", nil)
					MonitorExternal(clients...)

					lastCheckCompleted = time.Now()
					mutexChecking.Unlock()

					// if the new state has changed, change currentHealthOK, and if the channel for state changes exists, signal the channel
					newState, _, _ := GetState()
					newHealthOK := len(newState) == 0
					mutexCurrentHealthOK.Lock()
					defer mutexCurrentHealthOK.Unlock()
					if currentHealthOK != newHealthOK {
						log.Event(nil, "health state change", log.Data{"prev_health": currentHealthOK, "new_health": newHealthOK})
						currentHealthOK = newHealthOK
						if stateChangeChan != nil {
							stateChangeChan <- newHealthOK
						}
					}
				}()
			}
		}
	}()

	return &ticker
}

// Close the ticker to exit its internal goroutine
func (ticker *Ticker) Close() {
	ticker.timeTicker.Stop()
	close(ticker.closing)
	<-ticker.closed
}
