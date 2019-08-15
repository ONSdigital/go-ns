package common

import (
	"net/http"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGetLangFromSubDomain(t *testing.T) {

	Convey("Given a request when GetLangFromSubDomain is called and subdomain cy. is used", t, func() {
		req, _ := http.NewRequest("GET", "http://cy.localhost:21800/jobs", nil)
		languageToUse := GetLangFromSubDomain(req)

		Convey("Then the language 'cy' is returned in a string", func() {
			So(languageToUse, ShouldEqual, LangCY)
		})
	})
	Convey("Given a request when GetLangFromSubDomain is called and no subdomain is used", t, func() {
		req, _ := http.NewRequest("GET", "http://localhost:21800/jobs", nil)
		languageToUse := GetLangFromSubDomain(req)

		Convey("Then the language 'en' is returned in a string", func() {
			So(languageToUse, ShouldEqual, LangEN)
		})
	})
}

func TestGetLangFromCookieOrDefault(t *testing.T) {

	Convey("Given a cookie with lang set to 'cy' when GetLangFromCookieOrDefault is called", t, func() {
		cookie := http.Cookie{Name: "lang", Value: LangCY}
		languageToUse := GetLangFromCookieOrDefault(&cookie)

		Convey("Then lang returns a string 'cy'", func() {
			So(languageToUse, ShouldEqual, LangCY)
		})
	})
	Convey("Given a cookie with lang set to 'cy' when GetLangFromCookieOrDefault is called", t, func() {
		cookie := http.Cookie{Name: "lang", Value: LangEN}
		languageToUse := GetLangFromCookieOrDefault(&cookie)

		Convey("Then lang returns a string 'en'", func() {
			So(languageToUse, ShouldEqual, LangEN)
		})
	})
}

func TestSetLocaleCode(t *testing.T) {

	Convey("Given a request on cy domain", t, func() {
		req, _ := http.NewRequest("GET", "http://cy.localhost:21800/jobs", nil)
		Convey(" And given a cookie containing 'en' lang", func() {
			cookie := http.Cookie{Name: "lang", Value: LangEN}
			req.AddCookie(&cookie)
			req = SetLocaleCode(req)
			Convey("Then lang returns a string 'en'", func() {
				So(req.Header.Get("LocaleCode"), ShouldEqual, LangEN)
			})

		})
		Convey("and given a cookie containing 'cy' lang", func() {
			cookie := http.Cookie{Name: "lang", Value: LangCY}
			req.AddCookie(&cookie)
			req = SetLocaleCode(req)
			Convey("Then lang returns a string 'cy'", func() {
				So(req.Header.Get("LocaleCode"), ShouldEqual, LangCY)
			})
		})
	})

	Convey("Given a request on no subdomain", t, func() {
		req, _ := http.NewRequest("GET", "http://localhost:21800/jobs", nil)
		Convey(" And given a cookie containing 'en' lang", func() {
			cookie := http.Cookie{Name: "lang", Value: LangEN}
			req.AddCookie(&cookie)
			req = SetLocaleCode(req)
			Convey("Then lang returns a string 'en'", func() {
				So(req.Header.Get("LocaleCode"), ShouldEqual, LangEN)
			})

		})
		Convey("and given a cookie containing 'cy' lang", func() {
			cookie := http.Cookie{Name: "lang", Value: LangCY}
			req.AddCookie(&cookie)
			req = SetLocaleCode(req)
			Convey("Then lang returns a string 'cy'", func() {
				So(req.Header.Get("LocaleCode"), ShouldEqual, LangCY)
			})
		})
	})

	Convey("Given a request on cy subdomain", t, func() {
		req, _ := http.NewRequest("GET", "http://cy.localhost:21800/jobs", nil)
		Convey(" And no cookie set", func() {
			req = SetLocaleCode(req)
			Convey("Then lang returns a string 'cy'", func() {
				So(req.Header.Get("LocaleCode"), ShouldEqual, LangCY)
			})
		})
	})

	Convey("Given a request on no subdomain", t, func() {
		req, _ := http.NewRequest("GET", "http://localhost:21800/jobs", nil)
		Convey(" And no cookie set", func() {
			req = SetLocaleCode(req)
			Convey("Then lang returns a string 'en'", func() {
				So(req.Header.Get("LocaleCode"), ShouldEqual, LangEN)
			})
		})
	})

	Convey("Given a request on no subdomain", t, func() {
		req, _ := http.NewRequest("GET", "http://localhost:21800/jobs", nil)
		Convey("And cookie set to invalid/unused localeCode cookie set", func() {
			cookie := http.Cookie{Name: "lang", Value: "foo"}
			req.AddCookie(&cookie)
			req = SetLocaleCode(req)
			Convey("Then lang returns a string 'en'", func() {
				So(req.Header.Get("LocaleCode"), ShouldEqual, LangEN)
			})
		})
	})

}
