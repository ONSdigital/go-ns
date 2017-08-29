package healthcheck

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/ONSdigital/go-ns/log"
)

// Client represents the methods required to healthcheck a client
type Client interface {
	Healthcheck() (string, error)
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
		if name, err := client.Healthcheck(); err != nil {
			healthState[name] = err
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
