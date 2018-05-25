package requestID

import (
	"context"
	"net/http"

	"github.com/ONSdigital/go-ns/common"
)

type contextKey string

// ContextKey represents the constant key name
const ContextKey = contextKey("request-id")

// Handler is a wrapper which adds an X-Request-Id header if one does not yet exist
func Handler(size int) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {

		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			requestID := req.Header.Get(common.RequestHeaderKey)

			if len(requestID) == 0 {
				requestID = common.NewRequestID(size)
				req.Header.Set(common.RequestHeaderKey, requestID)
			}

			ctx := context.WithValue(req.Context(), ContextKey, requestID)
			h.ServeHTTP(w, req.WithContext(ctx))
		})
	}
}

// Get retrieves the value of the context key (request-id)
func Get(ctx context.Context) string {
	id, _ := ctx.Value(ContextKey).(string)
	return id
}

// Set creates/overwrites the value of the context key (request-id)
func Set(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, ContextKey, requestID)
}
