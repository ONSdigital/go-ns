package identity

import (
	"errors"
	"net/http"

	"github.com/ONSdigital/go-ns/log"
	"context"
)

// Check wraps a HTTP handler. If authentication fails an error code is returned else the HTTP handler is called
func Check(handle func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		log.DebugR(r, "checking for an identity in request context", nil)

		callerIdentity := r.Context().Value(callerIdentityKey)
		logData := log.Data{"caller_identity": callerIdentity}

		// just checking if an identity exists until permissions are being provided.
		if !IsPresent(r.Context()) {
			http.Error(w, "requested resource not found", http.StatusNotFound)
			log.ErrorR(r, errors.New("no identity was found in the context of this request"), logData)
			return
		}

		log.DebugR(r, "identity found in request context, calling downstream handler", logData)

		// The request has been authenticated, now run the clients request
		handle(w, r)
	})
}

// IsPresent determines if an identity is present on the given context.
func IsPresent(ctx context.Context) bool {

	callerIdentity := ctx.Value(callerIdentityKey)
	isPresent := callerIdentity != nil && callerIdentity != ""

	return isPresent
}

// Caller gets the caller identity from the context
func Caller(ctx context.Context) (string) {

	callerIdentity, _ := ctx.Value(callerIdentityKey).(string)
	return callerIdentity
}

// User gets the user identity from the context
func User(ctx context.Context) (string) {

	userIdentity, _ := ctx.Value(userIdentityKey).(string)
	return userIdentity
}
