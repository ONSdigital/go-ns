package common

import (
	"net/http"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestExtractLangFromSubDomain(t *testing.T) {

	Convey("Given a request when ExtractLangFromSubDomain is called and subdomain cy. is used", func() {
		req, _ := http.NewRequest("GET", "http://cy.localhost:21800/jobs", nil)
		lang := "cy"
		languageToUse := ExtractLangFromSubDomain(req)

		Convey("Then the language 'cy' is returned in a string", func() {
			So(languageToUse, ShouldEqual, lang)
		})
	})
	Convey("Given a request when ExtractLangFromSubDomain is called and no subdomain is used", func() {
		req, _ := http.NewRequest("GET", "http://localhost:21800/jobs", nil)

		lang := "en"
		languageToUse := ExtractLangFromSubDomain(req)

		Convey("Then the language 'en' is returned in a string", func() {
			So(languageToUse, ShouldEqual, lang)
		})
	})
}

func TestExtractLangFromCookie(t *testing.T) {

	Convey("Given a cookie with lang set to 'cy' when ExtractLangFromCookie is called", func() {
		cookie := http.Cookie{Name: "lang", Value: "cy"}
		lang := "cy"
		languageToUse := ExtractLangFromCookie(&cookie)

		Convey("Then lang returns a string 'cy'", func() {
			So(languageToUse, ShouldEqual, lang)
		})
	})
	Convey("Given a cookie with lang set to 'cy' when ExtractLangFromCookie is called", func() {
		cookie := http.Cookie{Name: "lang", Value: "en"}
		lang := "en"
		languageToUse := ExtractLangFromCookie(&cookie)

		Convey("Then lang returns a string 'en'", func() {
			So(languageToUse, ShouldEqual, lang)
		})
	})
}

func TestSetLocaleCode(t *testing.T) {

	Convey("Given a request on cy domain", t, func() {
		req, _ := http.NewRequest("GET", "http://cy.localhost:21800/jobs", nil)
		Convey(" And given a cookie containing 'en' lang", t, func() {
			cookie := http.Cookie{Name: "lang", Value: "en"}
			req.AddCookie(&cookie)
			ctx := SetLocaleCode(req)
			Convey("Then lang returns a string 'en'", func() {

				So(ctx.Value(ContextKey("LocaleCode")).(string), ShouldEqual, "en")
			})

		})
		Convey("and given a cookie containing 'cy' lang", t, func() {
			cookie := http.Cookie{Name: "lang", Value: "cy"}
			req.AddCookie(&cookie)
			ctx := SetLocaleCode(req)
			Convey("Then lang returns a string 'cy'", func() {
				So(ctx.Value(ContextKey("LocaleCode")).(string), ShouldEqual, "cy")
			})
		})
	})

	Convey("Given a request on no subdomain", t, func() {
		req, _ := http.NewRequest("GET", "http://localhost:21800/jobs", nil)
		Convey(" And given a cookie containing 'en' lang", t, func() {
			cookie := http.Cookie{Name: "lang", Value: "en"}
			req.AddCookie(&cookie)
			ctx := SetLocaleCode(req)
			Convey("Then lang returns a string 'en'", func() {
				So(ctx.Value(ContextKey("LocaleCode")).(string), ShouldEqual, "en")
			})

		})
		Convey("and given a cookie containing 'cy' lang", t, func() {
			cookie := http.Cookie{Name: "lang", Value: "cy"}
			req.AddCookie(&cookie)
			ctx := SetLocaleCode(req)
			Convey("Then lang returns a string 'cy'", func() {
				So(ctx.Value(ContextKey("LocaleCode")).(string), ShouldEqual, "cy")
			})
		})
	})

	Convey("Given a request on cy subdomain", t, func() {
		req, _ := http.NewRequest("GET", "http://cy.localhost:21800/jobs", nil)
		Convey(" And no cookie set", t, func() {
			ctx := SetLocaleCode(req)
			Convey("Then lang returns a string 'cy'", func() {
				So(ctx.Value(ContextKey("LocaleCode")).(string), ShouldEqual, "cy")
			})
		})
	})

	Convey("Given a request on no subdomain", t, func() {
		req, _ := http.NewRequest("GET", "http://localhost:21800/jobs", nil)
		Convey(" And no cookie set", t, func() {
			ctx := SetLocaleCode(req)
			Convey("Then lang returns a string 'en'", func() {
				So(ctx.Value(ContextKey("LocaleCode")).(string), ShouldEqual, "en")
			})
		})
	})

}
