package clientlog

import (
	"fmt"

	"github.com/ONSdigital/go-ns/log"
)

// Do should be used by clients to log a request to a given service
// before it is made
func Do(action, service, uri, method string) {
	log.Trace(fmt.Sprintf("Making request to service: %s", service), log.Data{
		"action": action,
		"method": method,
		"uri":    uri,
	})
}
