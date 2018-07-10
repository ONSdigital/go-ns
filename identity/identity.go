package identity

import (
	"net/http"

	clientsidentity "github.com/ONSdigital/go-ns/clients/identity"
	"github.com/ONSdigital/go-ns/log"
)

// Handler controls the authenticating of a request
func Handler(zebedeeURL string) func(http.Handler) http.Handler {
	authClient := clientsidentity.NewAPIClient(nil, zebedeeURL)
	return HandlerForHTTPClient(authClient)
}

// HandlerForHTTPClient allows a handler to be created that uses the given HTTP client
func HandlerForHTTPClient(cli clientsidentity.Clienter) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

			log.DebugR(req, "identity middleware called", nil)

			ctx, statusCode, err := cli.CheckRequest(req)
			if err != nil {
				logData := log.Data{"auth_status_code": statusCode}
				log.ErrorR(req, err, logData)

				w.WriteHeader(statusCode)
				return
			}
			req = req.WithContext(ctx)

			log.DebugR(req, "identity middleware finished, calling downstream handler", nil)

			h.ServeHTTP(w, req)
		})
	}
}
