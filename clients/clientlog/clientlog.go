package clientlog

import (
	"context"
	"fmt"

	"github.com/ONSdigital/go-ns/log"
)

// Do should be used by clients to log a request to a given service
// before it is made. If no log.Data is given then the request type
// is assumed to be GET
func Do(ctx context.Context, action, service, uri string, data ...log.Data) {
	d := log.Data{
		"action": action,
		"uri":    uri,
	}

	if len(data) == 0 {
		d["method"] = "GET"
	} else {
		for _, dat := range data {
			for k, v := range dat {
				d[k] = v
			}
		}
	}

	log.InfoCtx(ctx, fmt.Sprintf("Making request to service: %s", service), d)
}
