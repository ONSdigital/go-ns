package common

import (
	"net/http"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestExtractLangFromSubDomain(t *testing.T) {

	Convey("Given a request when ExtractLangFromSubDomain is called and subdomain cy. is used", t, func() {
		req, _ := http.NewRequest("GET", "http://cy.localhost:21800/jobs", nil)
		languageToUse := ExtractLangFromSubDomain(req)

		Convey("Then the language 'cy' is returned in a string", func() {
			So(languageToUse, ShouldEqual, "cy")
		})
	})
	Convey("Given a request when ExtractLangFromSubDomain is called and no subdomain is used", t, func() {
		req, _ := http.NewRequest("GET", "http://localhost:21800/jobs", nil)
		languageToUse := ExtractLangFromSubDomain(req)

		Convey("Then the language 'en' is returned in a string", func() {
			So(languageToUse, ShouldEqual, "en")
		})
	})
}

func TestExtractLangFromCookie(t *testing.T) {

	Convey("Given a cookie with lang set to 'cy' when ExtractLangFromCookie is called", t, func() {
		cookie := http.Cookie{Name: "lang", Value: "cy"}
		languageToUse := ExtractLangFromCookie(&cookie)

		Convey("Then lang returns a string 'cy'", func() {
			So(languageToUse, ShouldEqual, "cy")
		})
	})
	Convey("Given a cookie with lang set to 'cy' when ExtractLangFromCookie is called", t, func() {
		cookie := http.Cookie{Name: "lang", Value: "en"}
		languageToUse := ExtractLangFromCookie(&cookie)

		Convey("Then lang returns a string 'en'", func() {
			So(languageToUse, ShouldEqual, "en")
		})
	})
}

func TestSetLocaleCode(t *testing.T) {

	Convey("Given a request on cy domain", t, func() {
		req, _ := http.NewRequest("GET", "http://cy.localhost:21800/jobs", nil)
		Convey(" And given a cookie containing 'en' lang", t, func() {
			cookie := http.Cookie{Name: "lang", Value: "en"}
			req.AddCookie(&cookie)
			req = SetLocaleCode(req)
			Convey("Then lang returns a string 'en'", func() {
				So(req.Header.Get("LocaleCode"), ShouldEqual, "en")
			})

		})
		Convey("and given a cookie containing 'cy' lang", t, func() {
			cookie := http.Cookie{Name: "lang", Value: "cy"}
			req.AddCookie(&cookie)
			req = SetLocaleCode(req)
			Convey("Then lang returns a string 'cy'", func() {
				So(req.Header.Get("LocaleCode"), ShouldEqual, "cy")
			})
		})
	})

	Convey("Given a request on no subdomain", t, func() {
		req, _ := http.NewRequest("GET", "http://localhost:21800/jobs", nil)
		Convey(" And given a cookie containing 'en' lang", t, func() {
			cookie := http.Cookie{Name: "lang", Value: "en"}
			req.AddCookie(&cookie)
			req = SetLocaleCode(req)
			Convey("Then lang returns a string 'en'", func() {
				So(req.Header.Get("LocaleCode"), ShouldEqual, "en")
			})

		})
		Convey("and given a cookie containing 'cy' lang", t, func() {
			cookie := http.Cookie{Name: "lang", Value: "cy"}
			req.AddCookie(&cookie)
			req = SetLocaleCode(req)
			Convey("Then lang returns a string 'cy'", func() {
				So(req.Header.Get("LocaleCode"), ShouldEqual, "cy")
			})
		})
	})

	Convey("Given a request on cy subdomain", t, func() {
		req, _ := http.NewRequest("GET", "http://cy.localhost:21800/jobs", nil)
		Convey(" And no cookie set", t, func() {
			req = SetLocaleCode(req)
			Convey("Then lang returns a string 'cy'", func() {
				So(req.Header.Get("LocaleCode"), ShouldEqual, "cy")
			})
		})
	})

	Convey("Given a request on no subdomain", t, func() {
		req, _ := http.NewRequest("GET", "http://localhost:21800/jobs", nil)
		Convey(" And no cookie set", t, func() {
			req = SetLocaleCode(req)
			Convey("Then lang returns a string 'en'", func() {
				So(req.Header.Get("LocaleCode"), ShouldEqual, "en")
			})
		})
	})

	Convey("Given a request on no subdomain", t, func() {
		req, _ := http.NewRequest("GET", "http://localhost:21800/jobs", nil)
		Convey("And cookie set to invalid/unused localeCode cookie set", t, func() {
			cookie := http.Cookie{Name: "lang", Value: "foo"}
			req.AddCookie(&cookie)
			req = SetLocaleCode(req)
			Convey("Then lang returns a string 'en'", func() {
				So(req.Header.Get("LocaleCode"), ShouldEqual, "en")
			})
		})
	})

}
