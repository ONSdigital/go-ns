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

// Handler controls the authenticating of a request
func Handler(doAuth bool) func(http.Handler) http.Handler {
	return handler(doAuth, rchttp.DefaultClient)
}

type identityResponse struct {
	Identifier string `json:"identifier"`
}

func handler(doAuth bool, cli HttpClient) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

			ctx := req.Context()

			if doAuth {
				florenceToken := req.Header.Get(florenceHeaderKey)
				authToken := req.Header.Get(authHeaderKey)

				isUserReq := len(florenceToken) > 0
				isServiceReq := len(authToken) > 0

				if isUserReq || isServiceReq {

					// set a default zebedee value if it isn't set
					if len(zebedeeURL) == 0 {
						zebedeeURL = "http://localhost:8082"
					}

					zebReq, err := http.NewRequest("GET", zebedeeURL+"/identity", nil)
					if err != nil {
						log.ErrorR(req, err, nil)
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
						log.ErrorR(req, err, nil)
						w.WriteHeader(http.StatusInternalServerError)
						return
					}

					// Check to see if the user is authorised
					if resp.StatusCode != http.StatusOK {
						log.DebugR(req, "unexpected status code returned from zebedee identity endpoint",
							log.Data{"status": resp.StatusCode})
						w.WriteHeader(resp.StatusCode)
						return
					}

					identityResp, err := unmarshalResponse(resp)
					if err != nil {
						log.ErrorR(req, err, nil)
						w.WriteHeader(http.StatusInternalServerError)
						return
					}

					userIdentity := identityResp.Identifier
					if !isUserReq {
						userIdentity = req.Header.Get(userIdentityKey)
					}

					ctx = context.WithValue(ctx, userIdentityKey, userIdentity)
					ctx = context.WithValue(ctx, callerIdentityKey, identityResp.Identifier)
				}
			}

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
