package identity

import (
	"context"
	"net/http"

	clientsidentity "github.com/ONSdigital/dp-api-clients-go/identity"
	"github.com/ONSdigital/go-ns/common"
	"github.com/ONSdigital/go-ns/request"
	"github.com/ONSdigital/log.go/log"
)

// Handler controls the authenticating of a request
func Handler(zebedeeURL string) func(http.Handler) http.Handler {
	authClient := clientsidentity.NewAPIClient(nil, zebedeeURL)
	return HandlerForHTTPClient(authClient)
}

// HandlerForHTTPClient allows a handler to be created that uses the given HTTP client
func HandlerForHTTPClient(cli *clientsidentity.Client) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ctx := req.Context()
			log.Event(ctx, "executing identity check middleware")

			florenceToken := getFlorenceTokenFromRequest(ctx, req)
			serviceAuthToken := req.Header.Get(common.AuthHeaderKey)

			ctx, statusCode, authFailure, err := cli.CheckRequest(req, florenceToken, serviceAuthToken)
			logData := log.Data{"auth_status_code": statusCode}

			if err != nil {
				log.Event(ctx, "identity client check request returned an error", log.Error(err), logData)
				request.DrainBody(req)
				w.WriteHeader(statusCode)
				return
			}

			if authFailure != nil {
				log.Event(ctx, "identity client check request returned an auth error", log.Error(authFailure), logData)
				request.DrainBody(req)
				w.WriteHeader(statusCode)
				return
			}

			log.Event(ctx, "identity client check request completed successfully invoking downstream http handler")

			req = req.WithContext(ctx)
			h.ServeHTTP(w, req)
		})
	}
}

func getFlorenceTokenFromRequest(ctx context.Context, req *http.Request) string {
	var florenceToken string

	florenceToken = req.Header.Get(common.FlorenceHeaderKey)
	if len(florenceToken) >= 1 {
		log.Event(ctx, "florence access token header found")
		return florenceToken
	}

	log.Event(ctx, "florence access token header not found attempting to find access token cookie")
	c, err := req.Cookie(common.FlorenceCookieKey)
	if err != nil {
		log.Event(ctx, "error getting florence access token cookie from request", log.Error(err))
		return florenceToken
	}

	log.Event(ctx, "found request access token cookie")
	florenceToken = c.Value

	return florenceToken
}
