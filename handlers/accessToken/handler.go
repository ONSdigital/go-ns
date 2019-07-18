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
		accessToken := req.Header.Get(common.AccessTokenHeaderKey)
		if accessToken != "" {
			req = addUserAccessTokenToRequestContext(accessToken, req)
		}

		h.ServeHTTP(w, req)
	})
}

// CheckCookieValueAndForwardWithRequestContext is a wrapper which adds a accessToken from the cookie to context if one does not yet exist
func CheckCookieValueAndForwardWithRequestContext(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		accessTokenCookie, err := req.Cookie(common.AccessTokenCookieKey)
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

// addUserAccessTokenToRequestContext add the user florence access token to the request context. TODO TECHNICAL DEBT:
//
// There is inconsistency around which content key is used to store/retrieve the token. As a temp fix we are adding the
// same value with both keys. To fix this properly we need to pick 1 context key and update all uses to use the same one.
func addUserAccessTokenToRequestContext(userAccessToken string, req *http.Request) *http.Request {
	req = req.WithContext(context.WithValue(req.Context(), common.AccessTokenHeaderKey, userAccessToken))
	return req.WithContext(context.WithValue(req.Context(), common.FlorenceIdentityKey, userAccessToken))
}
