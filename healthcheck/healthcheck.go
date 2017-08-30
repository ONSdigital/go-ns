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
	mutex.Unlock()
}

// Do is responsible for returning the healthcheck status to the user
func Do(w http.ResponseWriter, req *http.Request) {
	mutex.RLock()
	defer mutex.RUnlock()

	if len(healthState) > 0 {
		for k, v := range healthState {
			err := fmt.Errorf("unsuccessful healthcheck for %s: %v", k, v)
			log.ErrorR(req, err, nil)
		}
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}
