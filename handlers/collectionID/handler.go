package collectionID

import (
	"context"
	"net/http"

	"github.com/ONSdigital/go-ns/common"
	"github.com/ONSdigital/log.go/v2/log"
)

// CheckHeader is a wrapper which adds a CollectionID from the request header to context if one does not yet exist
func CheckHeader(h http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

		collectionID := req.Header.Get(common.CollectionIDHeaderKey)

		if collectionID != "" {
			req = req.WithContext(context.WithValue(req.Context(), common.CollectionIDHeaderKey, collectionID))
		}

		h.ServeHTTP(w, req)
	})
}

// CheckCookie is a wrapper which adds a CollectionID from the cookie to context if one does not yet exist
func CheckCookie(h http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

		collectionIDCookie, err := req.Cookie(common.CollectionIDCookieKey)
		if err == nil {
			collectionID := collectionIDCookie.Value
			req = req.WithContext(context.WithValue(req.Context(), common.CollectionIDHeaderKey, collectionID))
		} else {
			if err != http.ErrNoCookie {
				log.Error(req.Context(), "unexpected error while extracting collection ID from cookie", err)
			}
		}

		h.ServeHTTP(w, req)
	})
}
