package accessToken

import (
	"context"
	"github.com/ONSdigital/go-ns/common"
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"net/http/httptest"
	"testing"
)

const testToken = "666"

type mockHandler struct {
	invocations int
	ctx         context.Context
}

func (m *mockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.invocations += 1
	m.ctx = r.Context()
}

func TestCheckHeaderValueAndForwardWithRequestContext(t *testing.T) {
	Convey("given the request with a florence access token header ", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:8080", nil)
		r.Header.Set(common.AccessTokenHeaderKey, testToken)
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

			Convey("and the request context contains a value for key X-Florence-Token", func() {
				headerVal, ok := mockHandler.ctx.Value(common.AccessTokenHeaderKey).(string)
				So(ok, ShouldBeTrue)
				So(headerVal, ShouldEqual, testToken)
			})

			Convey("and the request context contains a value for key florence-id", func() {
				xFlorenceToken, ok := mockHandler.ctx.Value(common.FlorenceIdentityKey).(string)
				So(ok, ShouldBeTrue)
				So(xFlorenceToken, ShouldEqual, testToken)
			})
		})
	})

	Convey("given the request does not contain a florence access token header ", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:8080", nil)
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

			Convey("and the request context does not contain a value for key X-Florence-Token", func() {
				headerVal := mockHandler.ctx.Value(common.AccessTokenHeaderKey)
				So(headerVal, ShouldBeNil)
			})

			Convey("and the request context contains a value for key florence-id", func() {
				xFlorenceToken := mockHandler.ctx.Value(common.FlorenceIdentityKey)
				So(xFlorenceToken, ShouldBeNil)
			})
		})
	})
}

func TestCheckCookieValueAndForwardWithRequestContext(t *testing.T) {
	Convey("given the request contain a cookie for a florence access token header ", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:8080", nil)
		r.AddCookie(&http.Cookie{Name: common.AccessTokenCookieKey, Value: testToken})

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

			Convey("and the request context contains a value for key X-Florence-Token", func() {
				headerVal, ok := mockHandler.ctx.Value(common.AccessTokenHeaderKey).(string)
				So(ok, ShouldBeTrue)
				So(headerVal, ShouldEqual, testToken)
			})

			Convey("and the request context contains a value for key florence-id", func() {
				xFlorenceToken, ok := mockHandler.ctx.Value(common.FlorenceIdentityKey).(string)
				So(ok, ShouldBeTrue)
				So(xFlorenceToken, ShouldEqual, testToken)
			})
		})
	})

	Convey("given the request does not contain a cookie for a florence access token header ", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:8080", nil)
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

			Convey("and the request context does not contain value for key X-Florence-Token", func() {
				headerVal := mockHandler.ctx.Value(common.AccessTokenHeaderKey)
				So(headerVal, ShouldBeNil)
			})

			Convey("and the request context does not contain value for key florence-id", func() {
				xFlorenceToken := mockHandler.ctx.Value(common.FlorenceIdentityKey)
				So(xFlorenceToken, ShouldBeNil)
			})
		})
	})
}
