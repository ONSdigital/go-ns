package localeCode

import (
	"context"
	"net/http"

	"github.com/ONSdigital/dp-api-clients-go/headers"
	"github.com/ONSdigital/go-ns/common"
	"github.com/ONSdigital/log.go/log"
)

// CheckHeaderValueAndForwardWithRequestContext is a wrapper which adds a localeCode from the request header to context if one does not yet exist
func CheckHeaderValueAndForwardWithRequestContext(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

		localeCode, err := headers.GetLocaleCode(req)
		if err == nil && localeCode != "" {
			req = req.WithContext(context.WithValue(req.Context(), common.LocaleContextKey, localeCode))
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
			req = req.WithContext(context.WithValue(req.Context(), common.LocaleContextKey, localeCode))
		} else {
			if err != http.ErrNoCookie {
				log.Event(req.Context(), "unexpected error while extracting language from cookie", log.Error(err))
			}
		}

		h.ServeHTTP(w, req)
	})
}
