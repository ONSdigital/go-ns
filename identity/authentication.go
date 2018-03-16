package identity

import (
	"errors"
	"net/http"

	"github.com/ONSdigital/go-ns/log"
)

// Check wraps a HTTP handler. If authentication fails an error code is returned else the HTTP handler is called
func Check(handle func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		callerIdentity := r.Context().Value(callerIdentityKey)

		// just checking if an identity exists until permissions are being provided.
		if callerIdentity == nil || callerIdentity == ""{
			http.Error(w, "requested resource not found", http.StatusNotFound)
			log.Error(errors.New("client missing auth token in header"), nil)
			return
		}

		// The request has been authenticated, now run the clients request
		handle(w, r)
	})
}
