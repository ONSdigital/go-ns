package healthcheck

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/ONSdigital/go-ns/log"
)

// Client represents the methods required to healthcheck a client
type Client interface {
	Healthcheck() (string, error)
}

var (
	healthState       = make(map[string]error)
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
	hs := make(map[string]error)

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
				log.ErrorC("unsuccessful healthcheck", err, log.Data{"external_service": name})
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
	if len(hs) == 0 {
		healthLastSuccess = time.Now()
	}
	healthLastChecked = time.Now()
	mutex.Unlock()
}

// Do is responsible for returning the healthcheck status to the user
func Do(w http.ResponseWriter, req *http.Request) {
	mutex.RLock()
	defer mutex.RUnlock()

	w.Header().Set("Content-Type", "application/json")

	var healthStateInfo healthResponse
	if len(healthState) > 0 {
		w.WriteHeader(http.StatusInternalServerError)
		healthStateInfo = healthResponse{Status: "error", Errors: &[]healthError{}}
		// add errors to healthStateInfo and set its Status to "error"
		for stateKey := range healthState {
			*(healthStateInfo.Errors) = append(*(healthStateInfo.Errors), healthError{Namespace: stateKey, ErrorMessage: healthState[stateKey].Error()})
		}
	} else if healthLastChecked.IsZero() {
		w.WriteHeader(http.StatusTooManyRequests)
		return
	} else {
		w.WriteHeader(http.StatusOK)
		healthStateInfo.Status = "OK"
	}
	healthStateInfo.LastChecked = healthLastChecked
	healthStateInfo.LastSuccess = healthLastSuccess

	healthStateJSON, err := json.Marshal(healthStateInfo)
	if err != nil {
		log.ErrorC("marshal json", err, log.Data{"struct": healthStateInfo})
		return
	}
	if _, err = w.Write(healthStateJSON); err != nil {
		log.ErrorC("writing json body", err, log.Data{"json": string(healthStateJSON)})
	}
}
