package common

import (
	"context"
	"net/http"
	"strings"
)

// LocaleCode the locale code to use to determine the language to use
const LocaleCode = ContextKey("LocaleCode")

// SetLocaleCode sets the Locale code used to set the language
func SetLocaleCode(req *http.Request) context.Context {
	lang := ExtractLangFromSubDomain(req)

	if c, err := req.Cookie("lang"); err == nil && len(c.Value) > 0 {
		lang = ExtractLangFromCookie(c)
	}

	return context.WithValue(req.Context(), LocaleCode, lang)
}

// ExtractLangFromSubDomain returns a language based on subdomain
func ExtractLangFromSubDomain(req *http.Request) string {
	lang := "en"
	if strings.Split(req.Host, ".")[0] == "cy" {
		lang = "cy"
	} else {
		lang = "en"
	}
	return lang
}

// ExtractLangFromCookie returns a language based on the lang cookie
func ExtractLangFromCookie(c *http.Cookie) string {
	lang := c.Value
	return lang
}
