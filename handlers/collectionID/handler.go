package collectionID

import (
	"context"
	"net/http"
	"errors"

	"github.com/ONSdigital/go-ns/common"
	"github.com/ONSdigital/go-ns/log"
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
			collectionID := collectionIDCookie.String()
			req = req.WithContext(context.WithValue(req.Context(), common.CollectionIDHeaderKey, collectionID))
		} else {
			if err != http.ErrNoCookie {
				log.ErrorCtx(req.Context(), errors.New("unexpected error while extracting collection ID from cookie"), nil)
			}
		}

		h.ServeHTTP(w, req)
	})
}
