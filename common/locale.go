package common

import (
	"net/http"
	"strings"
)

const (
	LangEN = "en"
	LangCY = "cy"

	LocaleCookieKey = "lang"
	LocaleHeaderKey = "LocaleCode"
)

// SetLocaleCode sets the Locale code used to set the language
func SetLocaleCode(req *http.Request) *http.Request {
	localeCode := ExtractLangFromSubDomain(req)

	// Language is overridden by cookie 'lang' here if present.
	if c, err := req.Cookie(LocaleCookieKey); err == nil && len(c.Value) > 0 {
		localeCode = ExtractLangFromCookie(c)
	}
	req.Header.Set(LocaleHeaderKey, localeCode)

	return req
}

// ExtractLangFromSubDomain returns a language based on subdomain
func ExtractLangFromSubDomain(req *http.Request) string {
	args := strings.Split(req.Host, ".")
	if len(args) == 0 {
		// Defaulting to "en" (LangEN) if no arguments
		return LangEN
	}
	if strings.Split(req.Host, ".")[0] == LangCY {
		return LangCY
	}
	return LangEN
}

// ExtractLangFromCookie returns a language based on the lang cookie
func ExtractLangFromCookie(c *http.Cookie) string {
	if c.Value == LangCY || c.Value == LangEN {
		return c.Value
	}
	return LangEN

}
