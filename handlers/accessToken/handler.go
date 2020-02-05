package accessToken

import (
	"context"
	"net/http"

	"github.com/ONSdigital/dp-api-clients-go/headers"
	"github.com/ONSdigital/go-ns/common"
	"github.com/ONSdigital/log.go/log"
)

// CheckHeaderValueAndForwardWithRequestContext is a wrapper which adds a accessToken from the request header to context if one does not yet exist
func CheckHeaderValueAndForwardWithRequestContext(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		accessToken, err := headers.GetUserAuthToken(req)
		if err == nil {
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
				log.Event(req.Context(), "unexpected error while extracting user Florence access token from cookie", log.Error(err))
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
