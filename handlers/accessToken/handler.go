package accessToken

import (
	"context"
	"errors"
	"net/http"

	"github.com/ONSdigital/go-ns/common"
	"github.com/ONSdigital/go-ns/log"
)

// CheckHeaderValueAndForwardWithRequestContext is a wrapper which adds a accessToken from the request header to context if one does not yet exist
func CheckHeaderValueAndForwardWithRequestContext(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		accessToken := req.Header.Get(common.FlorenceHeaderKey)
		if accessToken != "" {
			req = addUserAccessTokenToRequestContext(accessToken, req)
		}

		h.ServeHTTP(w, req)
	})
}

// CheckCookieValueAndForwardWithRequestContext is a wrapper which adds a accessToken from the cookie to context if one does not yet exist
func CheckCookieValueAndForwardWithRequestContext(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		accessTokenCookie, err := req.Cookie(common.FlorenceCookieKey)
		if err != nil {
			if err != http.ErrNoCookie {
				log.ErrorCtx(req.Context(), errors.New("unexpected error while extracting user Florence access token from cookie"), nil)
			}
		} else {
			req = addUserAccessTokenToRequestContext(accessTokenCookie.Value, req)
		}

		h.ServeHTTP(w, req)
	})
}

// addUserAccessTokenToRequestContext add the user florence access token to the request context.
func addUserAccessTokenToRequestContext(userAccessToken string, req *http.Request) *http.Request {
	return req.WithContext(context.WithValue(req.Context(), common.FlorenceIdentityKey, userAccessToken))
}
