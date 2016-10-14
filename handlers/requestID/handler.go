package requestID

import (
	"math/rand"
	"net/http"
)

// Handler is a wrapper which adds an X-Request-Id header if one does not yet exist
func Handler(size int) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			requestID := req.Header.Get("X-Request-Id")

			if len(requestID) == 0 {
				b := make([]rune, size)
				for i := range b {
					b[i] = letters[rand.Intn(len(letters))]
				}
				req.Header.Set("X-Request-Id", string(b))
			}

			h.ServeHTTP(w, req)
		})
	}
}
