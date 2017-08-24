package healthcheck

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/ONSdigital/go-ns/log"
)

// Client represents the methods required to healthcheck a client
type Client interface {
	Healthcheck() error
	Name() string
}

var (
	healthState = make(map[string]error)
	mutex       = &sync.Mutex{}
)

// MonitorExternal monitors external service health and if they are unhealthy, records the
// status in a map
func MonitorExternal(clients ...Client) {
	mutex.Lock()
	healthState = make(map[string]error)

	for _, client := range clients {
		if err := client.Healthcheck(); err != nil {
			healthState[client.Name()] = err
		}
	}
	mutex.Unlock()
}

// Do is responsible for returning the healthcheck status to the user
func Do(w http.ResponseWriter, req *http.Request) {
	mutex.Lock()

	if len(healthState) > 0 {
		for k, v := range healthState {
			err := fmt.Errorf("unsuccessful healthcheck for %s: %v", k, v)
			log.ErrorR(req, err, nil)
			w.WriteHeader(http.StatusInternalServerError)
		}
	} else {
		w.WriteHeader(http.StatusOK)
	}

	mutex.Unlock()
}
