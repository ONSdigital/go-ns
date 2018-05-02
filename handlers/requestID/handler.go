package requestID

import (
	"context"
	"math/rand"
	"net/http"
	"time"
)

var requestIDRandom = rand.New(rand.NewSource(time.Now().UnixNano()))

type contextKey string

const ContextKey = contextKey("request-id")

// Handler is a wrapper which adds an X-Request-Id header if one does not yet exist
func Handler(size int) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			requestID := req.Header.Get("X-Request-Id")

			if len(requestID) == 0 {
				b := make([]rune, size)
				for i := range b {
					b[i] = letters[requestIDRandom.Intn(len(letters))]
				}
				requestID = string(b)
				req.Header.Set("X-Request-Id", requestID)
			}

			ctx := context.WithValue(req.Context(), ContextKey, requestID)
			h.ServeHTTP(w, req.WithContext(ctx))
		})
	}
}

func Get(ctx context.Context) string {
	id, _ := ctx.Value(ContextKey).(string)
	return id
}
