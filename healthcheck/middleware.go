package healthcheck

import (
	"net/http"
)

// HandlerForHTTPClient allows a handler to be created that uses the given HTTP client
func NewMiddleware(healthcheckHandler func(http.ResponseWriter, *http.Request)) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

			if req.Method == "GET" && req.URL.Path == "/healthcheck" {
				healthcheckHandler(w, req)
				return
			}

			h.ServeHTTP(w, req)
		})
	}
}
