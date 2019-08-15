package localeCode

import (
	"context"
	"errors"
	"net/http"

	"github.com/ONSdigital/go-ns/common"
	"github.com/ONSdigital/go-ns/log"
)

// CheckHeaderValueAndForwardWithRequestContext is a wrapper which adds a localeCode from the request header to context if one does not yet exist
func CheckHeaderValueAndForwardWithRequestContext(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

		localeCode := req.Header.Get(common.LocaleHeaderKey)

		if localeCode != "" {
			req = req.WithContext(context.WithValue(req.Context(), common.LocaleHeaderKey, localeCode))
		}

		h.ServeHTTP(w, req)
	})
}

// CheckCookieValueAndForwardWithRequestContext is a wrapper which adds a localeCode from the cookie to context if one does not yet exist
func CheckCookieValueAndForwardWithRequestContext(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

		localeCodeCookie, err := req.Cookie(common.LocaleCookieKey)
		if err == nil {
			localeCode := localeCodeCookie.Value
			req = req.WithContext(context.WithValue(req.Context(), common.LocaleHeaderKey, localeCode))
		} else {
			if err != http.ErrNoCookie {
				log.ErrorCtx(req.Context(), errors.New("unexpected error while extracting language from cookie"), nil)
			}
		}

		h.ServeHTTP(w, req)
	})
}
