package identity

import (
	"errors"
	"net/http"

	"github.com/ONSdigital/go-ns/log"
)

// Check wraps a HTTP handler. If authentication fails an error code is returned else the HTTP handler is called
func Check(handle func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		log.DebugR(r, "checking for an identity in request context", nil)

		callerIdentity := r.Context().Value(callerIdentityKey)
		logData := log.Data{ "caller_identity": callerIdentity }

		// just checking if an identity exists until permissions are being provided.
		if callerIdentity == nil || callerIdentity == ""{
			http.Error(w, "requested resource not found", http.StatusNotFound)
			log.ErrorR(r, errors.New("no identity was found in the context of this request"), logData)
			return
		}

		log.DebugR(r, "identity found in request context, calling downstream handler", logData)

		// The request has been authenticated, now run the clients request
		handle(w, r)
	})
}
