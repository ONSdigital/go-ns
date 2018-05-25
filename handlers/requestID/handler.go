package requestID

import (
	"context"
	"math/rand"
	"net/http"
	"time"
)

var requestIDRandom = rand.New(rand.NewSource(time.Now().UnixNano()))

type contextKey string

// ContextKey represents the constant key name name
const ContextKey = contextKey("request-id")

// Handler is a wrapper which adds an X-Request-Id header if one does not yet exist
func Handler(size int) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {

		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			requestID := req.Header.Get("X-Request-Id")

			if len(requestID) == 0 {
				requestID = NewRequestID(size)
				req.Header.Set("X-Request-Id", requestID)
			}

			ctx := context.WithValue(req.Context(), ContextKey, requestID)
			h.ServeHTTP(w, req.WithContext(ctx))
		})
	}
}

// NewRequestID generates a random string of requested length
func NewRequestID(size int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, size)
	for i := range b {
		b[i] = letters[requestIDRandom.Intn(len(letters))]
	}
	return string(b)
}

// Get retrieves the value of the context key (request-id)
func Get(ctx context.Context) string {
	id, _ := ctx.Value(ContextKey).(string)
	return id
}

// Set creates/overwrites the value of the context key (request-id)
func Set(ctx context.Context, requestID string) context.Context {
	ctx = context.WithValue(ctx, ContextKey, requestID)
	return ctx
}
