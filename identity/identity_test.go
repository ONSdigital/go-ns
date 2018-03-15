package identity

import (
	"context"
	"github.com/ONSdigital/go-ns/identity/identitytest"
	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"net/http/httptest"
	"testing"
)

var url = "/whatever"

func TestHandler_NoAuth(t *testing.T) {

	Convey("Given an instance of identity handler with doAuth set to false", t, func() {

		doAuth := false
		request := httptest.NewRequest("GET", url, nil)
		responseRecorder := httptest.NewRecorder()

		httpClient := &identitytest.HttpClientMock{
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return nil, nil
			},
		}

		handlerCalled := false
		httpHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			handlerCalled = true
		})

		identityHandler := handler(doAuth, httpClient)(httpHandler)

		Convey("When ServeHTTP is called", func() {

			identityHandler.ServeHTTP(responseRecorder, request)

			Convey("Then the downstream HTTP handler is called", func() {
				So(handlerCalled, ShouldBeTrue)
			})
		})
	})
}

func TestHandler_NoHeaders(t *testing.T) {

	Convey("Given a http request with no headers", t, func() {

		doAuth := true
		request := httptest.NewRequest("GET", url, nil)
		responseRecorder := httptest.NewRecorder()

		httpClient := &identitytest.HttpClientMock{
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return nil, nil
			},
		}

		handlerCalled := false
		httpHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			handlerCalled = true
		})

		identityHandler := handler(doAuth, httpClient)(httpHandler)

		Convey("When ServeHTTP is called", func() {

			identityHandler.ServeHTTP(responseRecorder, request)

			Convey("Then the downstream HTTP handler is called", func() {
				So(handlerCalled, ShouldBeTrue)
			})
		})
	})
}

func TestHandler_florenceToken(t *testing.T) {

	Convey("Given a request with a florence token, and mock client that returns 200", t, func() {

		doAuth := true
		request := httptest.NewRequest("GET", url, nil)
		request.Header = map[string][]string{
			"X-Florence-Token": {"123"},
		}
		responseRecorder := httptest.NewRecorder()

		httpClient := &identitytest.HttpClientMock{
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
				}, nil
			},
		}

		handlerCalled := false
		httpHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			handlerCalled = true
		})

		identityHandler := handler(doAuth, httpClient)(httpHandler)

		Convey("When ServeHTTP is called", func() {

			identityHandler.ServeHTTP(responseRecorder, request)

			Convey("Then the identity service is called as expected", func() {
				So(len(httpClient.DoCalls()), ShouldEqual, 1)
				So(httpClient.DoCalls()[0].Req.URL.String(), ShouldEqual, "http://localhost:8082/permission")
			})

			Convey("Then the downstream HTTP handler is called", func() {
				So(handlerCalled, ShouldBeTrue)
			})
		})
	})
}

func TestHandler_IdentityServiceError(t *testing.T) {

	Convey("Given a request with a florence token, and mock client that returns an error", t, func() {

		doAuth := true
		request := httptest.NewRequest("GET", url, nil)
		request.Header = map[string][]string{
			"X-Florence-Token": {"123"},
		}
		responseRecorder := httptest.NewRecorder()

		httpClient := &identitytest.HttpClientMock{
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return nil, errors.New("broken")
			},
		}

		handlerCalled := false
		httpHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			handlerCalled = true
		})

		identityHandler := handler(doAuth, httpClient)(httpHandler)

		Convey("When ServeHTTP is called", func() {

			identityHandler.ServeHTTP(responseRecorder, request)

			Convey("Then the identity service is called as expected", func() {
				So(len(httpClient.DoCalls()), ShouldEqual, 1)
				So(httpClient.DoCalls()[0].Req.URL.String(), ShouldEqual, "http://localhost:8082/permission")
			})

			Convey("Then the response code is set as expected", func() {
				So(responseRecorder.Code, ShouldEqual, http.StatusInternalServerError)
			})

			Convey("Then the downstream HTTP handler is not called", func() {
				So(handlerCalled, ShouldBeFalse)
			})
		})
	})
}

func TestHandler_IdentityServiceErrorResponseCode(t *testing.T) {

	Convey("Given a request with a florence token, and mock client that returns a non 200 response", t, func() {

		doAuth := true
		request := httptest.NewRequest("GET", url, nil)
		request.Header = map[string][]string{
			"X-Florence-Token": {"123"},
		}
		responseRecorder := httptest.NewRecorder()

		httpClient := &identitytest.HttpClientMock{
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusNotFound,
				}, nil
			},
		}

		handlerCalled := false
		httpHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			handlerCalled = true
		})

		identityHandler := handler(doAuth, httpClient)(httpHandler)

		Convey("When ServeHTTP is called", func() {

			identityHandler.ServeHTTP(responseRecorder, request)

			Convey("Then the identity service is called as expected", func() {
				So(len(httpClient.DoCalls()), ShouldEqual, 1)
				So(httpClient.DoCalls()[0].Req.URL.String(), ShouldEqual, "http://localhost:8082/permission")
			})

			Convey("Then the response code is the same as returned from the identity service", func() {
				So(responseRecorder.Code, ShouldEqual, http.StatusNotFound)
			})

			Convey("Then the downstream HTTP handler is not called", func() {
				So(handlerCalled, ShouldBeFalse)
			})
		})
	})
}
