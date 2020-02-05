package requestID

import (
	"net/http"

	"github.com/ONSdigital/dp-api-clients-go/headers"
	"github.com/ONSdigital/go-ns/common"
)

// Handler is a wrapper which adds an X-Request-Id header if one does not yet exist
func Handler(size int) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {

		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			requestID, err := headers.GetRequestID(req)
			if err != nil || len(requestID) == 0 {
				requestID = common.NewRequestID(size)
				headers.SetRequestID(req, requestID)
			}

			h.ServeHTTP(w, req.WithContext(common.WithRequestId(req.Context(), requestID)))
		})
	}
}
