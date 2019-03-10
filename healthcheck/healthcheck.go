package healthcheck

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/ONSdigital/log.go/log"
)

// Client represents the methods required to healthcheck a client
type Client interface {
	Healthcheck() (string, error)
}

type HealthState map[string]error

var (
	healthState       = make(HealthState)
	healthLastChecked time.Time
	healthLastSuccess time.Time
	mutex             = &sync.RWMutex{}
)

type healthResponse struct {
	Status      string         `json:"status"`
	Errors      *[]healthError `json:"errors,omitempty"`
	LastSuccess time.Time      `json:"last_success,omitempty"`
	LastChecked time.Time      `json:"last_checked,omitempty"`
}
type healthError struct {
	Namespace    string `json:"namespace"`
	ErrorMessage string `json:"error"`
}

// MonitorExternal concurrently monitors external service health and if they are unhealthy,
// records the status in a map
func MonitorExternal(clients ...Client) {
	hs := make(HealthState)

	type externalError struct {
		name string
		err  error
	}
	errs := make(chan externalError)
	done := make(chan bool)

	go func() {
		for extErr := range errs {
			hs[extErr.name] = extErr.err
		}
		close(done)
	}()

	var wg sync.WaitGroup
	wg.Add(len(clients))
	for _, client := range clients {
		go func(client Client) {
			if name, err := client.Healthcheck(); err != nil {
				log.Event(nil, "unsuccessful healthcheck", log.Error(err), log.Data{"external_service": name})
				errs <- externalError{name, err}
			}
			wg.Done()
		}(client)
	}

	wg.Wait()
	close(errs)
	<-done

	mutex.Lock()
	healthState = hs
	healthLastChecked = time.Now()
	if len(hs) == 0 {
		healthLastSuccess = healthLastChecked
	}
	mutex.Unlock()
}

// Do is responsible for returning the healthcheck status to the user
func Do(w http.ResponseWriter, req *http.Request) {
	state, lastTry, lastSuccess := GetState()

	w.Header().Set("Content-Type", "application/json")

	var healthStateInfo healthResponse
	if len(state) > 0 {
		w.WriteHeader(http.StatusInternalServerError)
		healthStateInfo = healthResponse{Status: "error", Errors: &[]healthError{}}
		// add errors to healthStateInfo and set its Status to "error"
		for stateKey := range state {
			*(healthStateInfo.Errors) = append(*(healthStateInfo.Errors), healthError{Namespace: stateKey, ErrorMessage: state[stateKey].Error()})
		}
	} else if lastTry.IsZero() {
		w.WriteHeader(http.StatusTooManyRequests)
		return
	} else {
		w.WriteHeader(http.StatusOK)
		healthStateInfo.Status = "OK"
	}
	healthStateInfo.LastChecked = lastTry
	healthStateInfo.LastSuccess = lastSuccess

	healthStateJSON, err := json.Marshal(healthStateInfo)
	if err != nil {
		log.Event(req.Context(), "error marshaling json", log.Error(err), log.Data{"struct": healthStateInfo})
		return
	}
	if _, err = w.Write(healthStateJSON); err != nil {
		log.Event(req.Context(), "error writing json body", log.Error(err), log.Data{"json": string(healthStateJSON)})
	}
}

// GetState returns current map of errors and times of last check, last success
func GetState() (HealthState, time.Time, time.Time) {
	mutex.RLock()
	defer mutex.RUnlock()

	hs := make(HealthState)
	for key := range healthState {
		hs[key] = healthState[key]
	}

	return hs, healthLastChecked, healthLastSuccess
}
