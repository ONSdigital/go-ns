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
	mutex       = &sync.RWMutex{}
)

// MonitorExternal concurrently monitors external service health and if they are unhealthy,
// records the status in a map
func MonitorExternal(clients ...Client) {
	hs := make(map[string]error)

	var wg sync.WaitGroup

	wg.Add(len(clients))
	for _, client := range clients {
		go func(client Client) {
			if name, err := client.Healthcheck(); err != nil {
				mutex.Lock()
				hs[name] = err
				mutex.Unlock()
			}
			wg.Done()
		}(client)
	}

	wg.Wait()

	mutex.RLock()
	healthState = hs
	mutex.RUnlock()
}

// Do is responsible for returning the healthcheck status to the user
func Do(w http.ResponseWriter, req *http.Request) {
	mutex.Lock()

	if len(healthState) > 0 {
		for k, v := range healthState {
			err := fmt.Errorf("unsuccessful healthcheck for %s: %v", k, v)
			log.ErrorR(req, err, nil)
		}
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	mutex.Unlock()
}
