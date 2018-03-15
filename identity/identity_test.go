package identity

import (
	"context"
	"github.com/ONSdigital/go-ns/identity/identitytest"
	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"net/http/httptest"
	"testing"
	"bytes"
	"io/ioutil"
	"encoding/json"
)

var url = "/whatever"
var florenceToken = "123"
var authToken = "345"
var serviceIdentifier = "api1"
var userIdentifier = "fred@ons.gov.uk"
var expectedZebedeeUrl = "http://localhost:8082/identity"

func TestHandler_NoAuth(t *testing.T) {

	Convey("Given an instance of identity handler with doAuth set to false", t, func() {

		doAuth := false
		req := httptest.NewRequest("GET", url, nil)
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

			identityHandler.ServeHTTP(responseRecorder, req)

			Convey("Then the downstream HTTP handler is called", func() {
				So(handlerCalled, ShouldBeTrue)
			})
		})
	})
}

func TestHandler_NoHeaders(t *testing.T) {

	Convey("Given a http request with no headers", t, func() {

		doAuth := true
		req := httptest.NewRequest("GET", url, nil)
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

			identityHandler.ServeHTTP(responseRecorder, req)

			Convey("Then the downstream HTTP handler is called", func() {
				So(handlerCalled, ShouldBeTrue)
			})
		})
	})
}

func TestHandler_IdentityServiceError(t *testing.T) {

	Convey("Given a request with a florence token, and mock client that returns an error", t, func() {

		doAuth := true
		req := httptest.NewRequest("GET", url, nil)
		req.Header = map[string][]string{
			florenceHeaderKey: {florenceToken},
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

			identityHandler.ServeHTTP(responseRecorder, req)

			Convey("Then the identity service is called as expected", func() {
				So(len(httpClient.DoCalls()), ShouldEqual, 1)
				So(httpClient.DoCalls()[0].Req.URL.String(), ShouldEqual, expectedZebedeeUrl)
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
		req := httptest.NewRequest("GET", url, nil)
		req.Header = map[string][]string{
			florenceHeaderKey: {florenceToken},
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

			identityHandler.ServeHTTP(responseRecorder, req)

			Convey("Then the identity service is called as expected", func() {
				So(len(httpClient.DoCalls()), ShouldEqual, 1)
				So(httpClient.DoCalls()[0].Req.URL.String(), ShouldEqual, expectedZebedeeUrl)
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

func TestHandler_florenceToken(t *testing.T) {

	Convey("Given a request with a florence token, and mock client that returns 200", t, func() {

		doAuth := true
		req := httptest.NewRequest("GET", url, nil)
		req.Header = map[string][]string{
			florenceHeaderKey: {florenceToken},
		}
		responseRecorder := httptest.NewRecorder()

		httpClient := &identitytest.HttpClientMock{
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {

				response := &identityResponse{
					Identifier: userIdentifier,
				}

				body, _ := json.Marshal(response)
				reader := bytes.NewBuffer(body)
				readCloser := ioutil.NopCloser(reader)

				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       readCloser,
				}, nil
			},
		}

		handlerCalled := false
		var handlerReq *http.Request
		httpHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			handlerReq = req
			handlerCalled = true
		})

		identityHandler := handler(doAuth, httpClient)(httpHandler)

		Convey("When ServeHTTP is called", func() {

			identityHandler.ServeHTTP(responseRecorder, req)

			Convey("Then the identity service is called as expected", func() {
				So(len(httpClient.DoCalls()), ShouldEqual, 1)
				zebedeeReq := httpClient.DoCalls()[0].Req
				So(zebedeeReq.URL.String(), ShouldEqual, expectedZebedeeUrl)
				So(zebedeeReq.Header[florenceHeaderKey][0], ShouldEqual, florenceToken)
			})

			Convey("Then the downstream HTTP handler is called", func() {
				So(handlerCalled, ShouldBeTrue)
			})

			Convey("Then the downstream HTTP handler request has the expected context values", func() {
				So(handlerReq.Context().Value(callerIdentityKey), ShouldEqual, userIdentifier)
				So(handlerReq.Context().Value(userIdentityKey), ShouldEqual, userIdentifier)
			})
		})
	})
}

func TestHandler_InvalidIdentityResponse(t *testing.T) {

	Convey("Given a request with a florence token, and mock client that returns invalid response JSON", t, func() {

		doAuth := true
		req := httptest.NewRequest("GET", url, nil)
		req.Header = map[string][]string{
			florenceHeaderKey: {florenceToken},
		}
		responseRecorder := httptest.NewRecorder()

		httpClient := &identitytest.HttpClientMock{
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {


				reader := bytes.NewBufferString("{ invalid JSON")
				readCloser := ioutil.NopCloser(reader)

				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       readCloser,
				}, nil
			},
		}

		handlerCalled := false
		var handlerReq *http.Request
		httpHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			handlerReq = req
			handlerCalled = true
		})

		identityHandler := handler(doAuth, httpClient)(httpHandler)

		Convey("When ServeHTTP is called", func() {

			identityHandler.ServeHTTP(responseRecorder, req)

			Convey("Then the identity service is called as expected", func() {
				So(len(httpClient.DoCalls()), ShouldEqual, 1)
				zebedeeReq := httpClient.DoCalls()[0].Req
				So(zebedeeReq.URL.String(), ShouldEqual, expectedZebedeeUrl)
				So(zebedeeReq.Header[florenceHeaderKey][0], ShouldEqual, florenceToken)
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

func TestHandler_authToken(t *testing.T) {

	Convey("Given a request with an auth token, and mock client that returns 200", t, func() {

		doAuth := true
		req := httptest.NewRequest("GET", url, nil)
		req.Header = map[string][]string{
			authHeaderKey: {authToken},
			userIdentityKey : {userIdentifier},
		}
		responseRecorder := httptest.NewRecorder()

		httpClient := &identitytest.HttpClientMock{
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {

				response := &identityResponse{
					Identifier: serviceIdentifier,
				}

				body, _ := json.Marshal(response)

				reader := bytes.NewBuffer(body)
				readCloser := ioutil.NopCloser(reader)

				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       readCloser,
				}, nil
			},
		}

		handlerCalled := false
		var handlerReq *http.Request
		httpHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			handlerReq = req
			handlerCalled = true
		})

		identityHandler := handler(doAuth, httpClient)(httpHandler)

		Convey("When ServeHTTP is called", func() {

			identityHandler.ServeHTTP(responseRecorder, req)

			Convey("Then the identity service is called as expected", func() {
				So(len(httpClient.DoCalls()), ShouldEqual, 1)
				zebedeeReq := httpClient.DoCalls()[0].Req
				So(zebedeeReq.URL.String(), ShouldEqual, expectedZebedeeUrl)
				So(zebedeeReq.Header[authHeaderKey][0], ShouldEqual, authToken)

			})

			Convey("Then the downstream HTTP handler is called", func() {
				So(handlerCalled, ShouldBeTrue)
			})

			Convey("Then the downstream HTTP handler request has the expected context values", func() {
				So(handlerReq.Context().Value(callerIdentityKey), ShouldEqual, serviceIdentifier)
				So(handlerReq.Context().Value(userIdentityKey), ShouldEqual, userIdentifier)
			})
		})
	})
}

func TestHandler_bothTokens(t *testing.T) {

	Convey("Given a request with both a florence token and service token", t, func() {

		doAuth := true
		req := httptest.NewRequest("GET", url, nil)
		req.Header = map[string][]string{
			florenceHeaderKey: {florenceToken},
			authHeaderKey: {authToken},
		}
		responseRecorder := httptest.NewRecorder()

		httpClient := &identitytest.HttpClientMock{
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {

				response := &identityResponse{
					Identifier: userIdentifier,
				}

				body, _ := json.Marshal(response)
				reader := bytes.NewBuffer(body)
				readCloser := ioutil.NopCloser(reader)

				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       readCloser,
				}, nil
			},
		}

		handlerCalled := false
		var handlerReq *http.Request
		httpHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			handlerReq = req
			handlerCalled = true
		})

		identityHandler := handler(doAuth, httpClient)(httpHandler)

		Convey("When ServeHTTP is called", func() {

			identityHandler.ServeHTTP(responseRecorder, req)

			Convey("Then the identity service is called as expected", func() {
				So(len(httpClient.DoCalls()), ShouldEqual, 1)
				zebedeeReq := httpClient.DoCalls()[0].Req
				So(zebedeeReq.URL.String(), ShouldEqual, expectedZebedeeUrl)
				So(zebedeeReq.Header[florenceHeaderKey][0], ShouldEqual, florenceToken)
			})

			Convey("Then the downstream HTTP handler is called", func() {
				So(handlerCalled, ShouldBeTrue)
			})

			Convey("Then the downstream HTTP handler request has the expected context values", func() {
				So(handlerReq.Context().Value(userIdentityKey), ShouldEqual, userIdentifier)
				So(handlerReq.Context().Value(callerIdentityKey), ShouldEqual, userIdentifier)
			})
		})
	})
}
