package identity

import (
	"context"
	"github.com/ONSdigital/dp-api-clients-go/headers"
	"net/http"

	clientsidentity "github.com/ONSdigital/dp-api-clients-go/identity"
	"github.com/ONSdigital/go-ns/common"
	"github.com/ONSdigital/go-ns/request"
	"github.com/ONSdigital/log.go/log"
)

type getTokenFromReqFunc func(ctx context.Context, r *http.Request) (string, error)

// Handler controls the authenticating of a request
func Handler(zebedeeURL string) func(http.Handler) http.Handler {
	authClient := clientsidentity.NewAPIClient(nil, zebedeeURL)
	return HandlerForHTTPClient(authClient)
}

// HandlerForHTTPClient allows a handler to be created that uses the given HTTP client
func HandlerForHTTPClient(cli *clientsidentity.Client) func(http.Handler) http.Handler {
	// maintain the public interface to ensure backwards compatible and allow the get X token functions to be passed into the handler func.
	return handlerForHTTPClient(cli, getFlorenceToken, getServiceAuthToken)
}

func handlerForHTTPClient(cli *clientsidentity.Client, getFlorenceToken, getServiceToken getTokenFromReqFunc) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ctx := req.Context()
			log.Event(ctx, "executing identity check middleware")

			florenceToken, err := getFlorenceToken(ctx, req)
			if err != nil {
				handleFailedRequest(ctx, w, req, http.StatusInternalServerError, "error getting florence access token from request", err, nil)
				return
			}

			serviceAuthToken, err := getServiceToken(ctx, req)
			if err != nil {
				handleFailedRequest(ctx, w, req, http.StatusInternalServerError, "error getting service access token from request", err, nil)
				return
			}

			ctx, statusCode, authFailure, err := cli.CheckRequest(req, florenceToken, serviceAuthToken)
			logData := log.Data{"auth_status_code": statusCode}

			if err != nil {
				handleFailedRequest(ctx, w, req, statusCode, "identity client check request returned an error", err, logData)
				return
			}

			if authFailure != nil {
				handleFailedRequest(ctx, w, req, statusCode, "identity client check request returned an auth error", authFailure, logData)
				log.Event(ctx, "identity client check request returned an auth error", log.Error(authFailure), logData)
				return
			}

			log.Event(ctx, "identity client check request completed successfully invoking downstream http handler")

			req = req.WithContext(ctx)
			h.ServeHTTP(w, req)
		})
	}
}

// handleFailedRequest adhering to the DRY principle - clean up for failed identity requests, log the error, drain the request body and write the status code.
func handleFailedRequest(ctx context.Context, w http.ResponseWriter, r *http.Request, status int, event string, err error, data log.Data) {
	log.Event(ctx, event, log.Error(err), data)
	request.DrainBody(r)
	w.WriteHeader(status)
}

func getFlorenceToken(ctx context.Context, req *http.Request) (string, error) {
	var florenceToken string

	token, err := headers.GetUserAuthToken(req)
	if err == nil {
		florenceToken = token
	} else if headers.IsErrNotFound(err) {
		log.Event(ctx, "florence access token header not found attempting to find access token cookie")
		florenceToken, err = getFlorenceTokenFromCookie(ctx, req)
	}

	return florenceToken, err
}

func getFlorenceTokenFromCookie(ctx context.Context, req *http.Request) (string, error) {
	var florenceToken string
	var err error

	c, err := req.Cookie(common.FlorenceCookieKey)
	if err == nil {
		florenceToken = c.Value
	} else if err == http.ErrNoCookie {
		err = nil // we don't consider this scenario an error so we set err to nil and return an empty token
		log.Event(ctx, "florence access token cookie not found in request")
	}

	return florenceToken, err
}

func getServiceAuthToken(ctx context.Context, req *http.Request) (string, error) {
	var authToken string

	token, err := headers.GetServiceAuthToken(req)
	if err == nil {
		authToken = token
	} else if headers.IsErrNotFound(err) {
		err = nil // we don't consider this scenario an error so we set err to nil and return an empty token
		log.Event(ctx, "service auth token request header is not found")
	}

	return authToken, err
}
