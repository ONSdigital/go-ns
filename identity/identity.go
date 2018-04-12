package identity

import (
	"errors"
	"net/http"

	clientsidentity "github.com/ONSdigital/go-ns/clients/identity"
	"github.com/ONSdigital/go-ns/log"
)

// Handler controls the authenticating of a request
func Handler(doAuth bool, zebedeeURL string) func(http.Handler) http.Handler {
	authClient := clientsidentity.NewAPIClient(nil, zebedeeURL)
	return HandlerForHTTPClient(doAuth, authClient)
}

// HandlerForHTTPClient allows a handler to be created that uses the given HTTP client
func HandlerForHTTPClient(doAuth bool, cli clientsidentity.IdentityClienter) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

			logData := log.Data{"auth_enabled": doAuth}
			log.DebugR(req, "identity middleware called", logData)

			if doAuth {
				ctx, statusCode, err := cli.CheckRequest(req)
				if err != nil || statusCode != http.StatusOK {
					logData["auth_status_code"] = statusCode
					if err == nil {
						err = errors.New("Got bad status from CheckRequest")
					}
					log.ErrorR(req, err, logData)
					if statusCode == 0 {
						statusCode = http.StatusInternalServerError
					}
					w.WriteHeader(statusCode)
					return
				}
				req = req.WithContext(ctx)
			}

			log.DebugR(req, "identity middleware finished, calling downstream handler", logData)

			h.ServeHTTP(w, req)
		})
	}
}
