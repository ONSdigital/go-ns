package identity

import (
	"net/http"
	"os"

	"github.com/ONSdigital/go-ns/rchttp"

	"context"
	"github.com/ONSdigital/go-ns/log"
	"encoding/json"
	"io/ioutil"
)

//go:generate moq -out identitytest/http_client.go -pkg identitytest . HttpClient

var zebedeeURL = os.Getenv("ZEBEDEE_URL")
const florenceHeaderKey = "X-Florence-Token"
const authHeaderKey = "Authorization"
const userIdentityKey = "User-Identity"
const callerIdentityKey = "Caller-Identity"

type HttpClient interface {
	Do(ctx context.Context, req *http.Request) (*http.Response, error)
}

type identityResponse struct {
	Identifier string `json:"identifier"`
}

// Handler controls the authenticating of a request
func Handler(doAuth bool) func(http.Handler) http.Handler {
	return HandlerForHttpClient(doAuth, rchttp.DefaultClient)
}

// HandlerForHttpClient allows a handler to be created that uses the given HTTP client
func HandlerForHttpClient(doAuth bool, cli HttpClient) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

			logData := log.Data{}
			log.DebugR(req, "identity middleware called", logData)

			ctx := req.Context()

			if doAuth {

				florenceToken := req.Header.Get(florenceHeaderKey)
				authToken := req.Header.Get(authHeaderKey)

				isUserReq := len(florenceToken) > 0
				isServiceReq := len(authToken) > 0

				logData["is_user_request"] = isUserReq
				logData["is_service_request"] = isServiceReq

				log.DebugR(req, "authentication enabled, checking for expected tokens", logData)

				if isUserReq || isServiceReq {

					// set a default zebedee value if it isn't set
					if len(zebedeeURL) == 0 {
						zebedeeURL = "http://localhost:8082"
					}

					url := zebedeeURL + "/identity"

					logData["url"] = url
					log.DebugR(req, "calling zebedee to authenticate", logData)

					zebReq, err := http.NewRequest("GET", url, nil)
					if err != nil {
						log.ErrorR(req, err, logData)
						w.WriteHeader(http.StatusInternalServerError)
						return
					}

					if isUserReq {
						zebReq.Header.Set(florenceHeaderKey, florenceToken)
					} else {
						zebReq.Header.Set(authHeaderKey, authToken)
					}

					resp, err := cli.Do(ctx, zebReq)
					if err != nil {
						log.ErrorR(req, err, logData)
						w.WriteHeader(http.StatusInternalServerError)
						return
					}

					// Check to see if the user is authorised
					if resp.StatusCode != http.StatusOK {
						logData["status"] = resp.StatusCode
						log.DebugR(req, "unexpected status code returned from zebedee identity endpoint", logData)
						w.WriteHeader(resp.StatusCode)
						return
					}

					identityResp, err := unmarshalResponse(resp)
					if err != nil {
						log.ErrorR(req, err, logData)
						w.WriteHeader(http.StatusInternalServerError)
						return
					}

					userIdentity := identityResp.Identifier
					if !isUserReq {
						userIdentity = req.Header.Get(userIdentityKey)
					}

					logData["user_identity"] = userIdentity
					logData["caller_identity"] = identityResp.Identifier
					log.DebugR(req, "user identity retrieved, setting context values", logData)

					ctx = context.WithValue(ctx, userIdentityKey, userIdentity)
					ctx = context.WithValue(ctx, callerIdentityKey, identityResp.Identifier)
				}
			} else {
				log.DebugR(req, "skipping authentication against zebedee, auth is not enabled", nil)
			}

			log.DebugR(req, "identity middleware finished, calling downstream handler", logData)

			h.ServeHTTP(w, req.WithContext(ctx))
		})
	}
}

func unmarshalResponse(resp *http.Response) (identityResp *identityResponse, err error) {

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return identityResp, err
	}
	defer resp.Body.Close()

	return identityResp, json.Unmarshal(b, &identityResp)
}
