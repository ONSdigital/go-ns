package common

import (
	"net/http"
	"strings"
)

// LangEN is a reference to the english short-form localeCode "en"
const LangEN = "en"

// LangCY is a reference to the english short-form localeCode "cy"
const LangCY = "cy"

// DefaultLang is the desired langauge to default to if requested language is unavailable
const DefaultLang = LangEN

// SetLocaleCode sets the Locale code used to set the language
func SetLocaleCode(req *http.Request) *http.Request {
	localeCode := GetLangFromSubDomain(req)

	// Language is overridden by cookie 'lang' here if present.
	if c, err := req.Cookie("lang"); err == nil && len(c.Value) > 0 {
		localeCode = GetLangFromCookieOrDefault(c)
	}
	req.Header.Set("LocaleCode", localeCode)

	return req
}

// GetLangFromSubDomain returns a language based on subdomain
func GetLangFromSubDomain(req *http.Request) string {
	args := strings.Split(req.Host, ".")
	if len(args) == 0 {
		// Defaulting to "en" (LangEN) if no arguments
		return LangEN
	}
	if args[0] == LangCY {
		return LangCY
	}
	return LangEN
}

// GetLangFromCookieOrDefault returns a language based on the lang cookie or if not valid defaults it
func GetLangFromCookieOrDefault(c *http.Cookie) string {
	if c.Value == LangCY || c.Value == LangEN {
		return c.Value
	}
	return DefaultLang
}
