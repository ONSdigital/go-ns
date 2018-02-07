package identity

import (
	"net/http"
	"os"

	"github.com/ONSdigital/go-ns/rchttp"

	"github.com/ONSdigital/go-ns/log"
)

var zebedeeURL = os.Getenv("ZEBEDEE_URL")

// Handler controls the authenticating of a request
func Handler(doAuth bool) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if doAuth && len(req.Header.Get("X-Florence-Token")) > 0 && len(zebedeeURL) > 0 {
				cli := rchttp.DefaultClient

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
