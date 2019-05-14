package collectionID

import (
	"net/http"
	"context"
	"github.com/ONSdigital/go-ns/common"
)

// Handler is a wrapper which adds a CollectionID header if one does not yet exist
func Handler(h http.Handler) http.Handler {

		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			collectionID := req.Header.Get(common.CollectionIDHeaderKey)
			h.ServeHTTP(w, req.WithContext(context.WithValue(req.Context(), common.CollectionIDHeaderKey, collectionID)))
		})
	}
