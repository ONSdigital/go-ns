package accessToken

import (
	"context"
	"errors"
	"net/http"

	"github.com/ONSdigital/go-ns/common"
	"github.com/ONSdigital/go-ns/log"
)

// CheckHeader is a wrapper which adds a accessToken from the request header to context if one does not yet exist
func CheckHeader(h http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

		accessToken := req.Header.Get(common.AccessTokenHeaderKey)

		if accessToken != "" {
			req = req.WithContext(context.WithValue(req.Context(), common.AccessTokenHeaderKey, accessToken))
		}

		h.ServeHTTP(w, req)
	})
}

// CheckCookie is a wrapper which adds a accessToken from the cookie to context if one does not yet exist
func CheckCookie(h http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

		accessTokenCookie, err := req.Cookie(common.AccessTokenCookieKey)
		if err == nil {
			accessToken := accessTokenCookie.Value
			req = req.WithContext(context.WithValue(req.Context(), common.AccessTokenHeaderKey, accessToken))
		} else {
			if err != http.ErrNoCookie {
				log.ErrorCtx(req.Context(), errors.New("unexpected error while extracting collection ID from cookie"), nil)
			}
		}

		h.ServeHTTP(w, req)
	})
}
