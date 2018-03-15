package identity

import (
	"net/http"
	"os"

	"github.com/ONSdigital/go-ns/rchttp"

	"context"
	"github.com/ONSdigital/go-ns/log"
)

//go:generate moq -out identitytest/http_client.go -pkg identitytest . HttpClient

var zebedeeURL = os.Getenv("ZEBEDEE_URL")

type HttpClient interface {
	Do(ctx context.Context, req *http.Request) (*http.Response, error)
}

func Handler(doAuth bool) func(http.Handler) http.Handler {
	return handler(doAuth, rchttp.DefaultClient)
}

// Handler controls the authenticating of a request
func handler(doAuth bool, cli HttpClient) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if doAuth && len(req.Header.Get("X-Florence-Token")) > 0 {

				// set a default zebedee value if it isn't set
				if len(zebedeeURL) == 0 {
					zebedeeURL = "http://localhost:8082"
				}

				zebReq, err := http.NewRequest("GET", zebedeeURL+"/permission", nil)
				if err != nil {
					log.ErrorR(req, err, nil)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				zebReq.Header.Set("X-Florence-Token", req.Header.Get("X-Florence-Token"))

				resp, err := cli.Do(req.Context(), zebReq)
				if err != nil {
					log.ErrorR(req, err, nil)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				// Check to see if the florence user is authorised
				if resp.StatusCode != http.StatusOK {
					w.WriteHeader(resp.StatusCode)
					return
				}

			}

			h.ServeHTTP(w, req)
		})
	}
}
