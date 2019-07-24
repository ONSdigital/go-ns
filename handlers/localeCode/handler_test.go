package localeCode

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ONSdigital/go-ns/common"
	. "github.com/smartystreets/goconvey/convey"
)

const testLocale = "cy"

type mockHandler struct {
	invocations int
	ctx         context.Context
}

func (m *mockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.invocations += 1
	m.ctx = r.Context()
}

func TestCheckHeaderValueAndForwardWithRequestContext(t *testing.T) {
	Convey("given the request with a locale header ", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:8080", nil)
		r.Header.Set(common.LocaleHeaderKey, testLocale)
		w := httptest.NewRecorder()

		mockHandler := &mockHandler{
			invocations: 0,
		}

		target := CheckHeaderValueAndForwardWithRequestContext(mockHandler)

		Convey("when the handler is called", func() {
			target.ServeHTTP(w, r)

			Convey("then the wrapped handle is called 1 time", func() {
				So(mockHandler.invocations, ShouldEqual, 1)
			})

			Convey("and the request context contains a value for key localeCode", func() {
				localeCode, ok := mockHandler.ctx.Value(common.LocaleHeaderKey).(string)
				fmt.Println("LOCALE CODE >>>>", localeCode)
				So(ok, ShouldBeTrue)
				So(localeCode, ShouldEqual, testLocale)
			})
		})
	})
}

func TestCheckCookieValueAndForwardWithRequestContext(t *testing.T) {
	Convey("given the request contain a cookie for locale code ", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:8080", nil)
		r.AddCookie(&http.Cookie{Name: common.LocaleCookieKey, Value: testLocale})

		w := httptest.NewRecorder()

		mockHandler := &mockHandler{
			invocations: 0,
		}

		target := CheckCookieValueAndForwardWithRequestContext(mockHandler)

		Convey("when the handler is called", func() {
			target.ServeHTTP(w, r)

			Convey("then the wrapped handle is called 1 time", func() {
				So(mockHandler.invocations, ShouldEqual, 1)
			})

			Convey("and the request context contains a value for key localeCode", func() {
				localeCode, ok := mockHandler.ctx.Value(common.LocaleHeaderKey).(string)
				fmt.Println("LOCALE CODE >>>>", localeCode)
				So(ok, ShouldBeTrue)
				So(localeCode, ShouldEqual, testLocale)
			})
		})
	})
}
