package identity

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	clientsidentity "github.com/ONSdigital/go-ns/clients/identity"
	"github.com/ONSdigital/go-ns/common"
	"github.com/ONSdigital/go-ns/common/commontest"
	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	url                = "/whatever"
	florenceToken      = "123"
	upstreamAuthToken  = "YourClaimToBeWhoYouAre"
	serviceIdentifier  = "api1"
	userIdentifier     = "fred@ons.gov.uk"
	zebedeeURL         = "http://localhost:8082"
	expectedZebedeeURL = zebedeeURL + "/identity"
)

func TestHandler_NoHeaders(t *testing.T) {

	Convey("Given a http request with no headers", t, func() {

		req := httptest.NewRequest("GET", url, nil)
		responseRecorder := httptest.NewRecorder()

		httpClient := &commontest.RCHTTPClienterMock{}
		idClient := clientsidentity.NewAPIClient(httpClient, zebedeeURL)

		handlerCalled := false
		httpHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			handlerCalled = true
		})

		identityHandler := HandlerForHTTPClient(idClient)(httpHandler)

		Convey("When ServeHTTP is called", func() {

			identityHandler.ServeHTTP(responseRecorder, req)

			Convey("Then the downstream HTTP handler is called", func() {
				So(handlerCalled, ShouldBeFalse)
			})

			Convey("Then the http response should have a 401 status", func() {
				So(responseRecorder.Result().StatusCode, ShouldEqual, http.StatusUnauthorized)
			})
		})
	})
}

func TestHandler_IdentityServiceError(t *testing.T) {

	Convey("Given a request with a florence token, and mock client that returns an error", t, func() {

		req := httptest.NewRequest("GET", url, nil)
		req.Header = map[string][]string{
			common.FlorenceHeaderKey: {florenceToken},
		}
		responseRecorder := httptest.NewRecorder()

		httpClient := &commontest.RCHTTPClienterMock{
			SetAuthTokenFunc: func(string) {},
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return nil, errors.New("broken")
			},
		}
		idClient := clientsidentity.NewAPIClient(httpClient, zebedeeURL)

		handlerCalled := false
		httpHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			handlerCalled = true
		})

		identityHandler := HandlerForHTTPClient(idClient)(httpHandler)

		Convey("When ServeHTTP is called", func() {

			identityHandler.ServeHTTP(responseRecorder, req)

			Convey("Then the identity service is called as expected", func() {
				So(len(httpClient.DoCalls()), ShouldEqual, 1)
				So(httpClient.DoCalls()[0].Req.URL.String(), ShouldEqual, expectedZebedeeURL)
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

	Convey("Given a request with a florence token, and mock client that returns a non-200 response", t, func() {

		req := httptest.NewRequest("GET", url, nil)
		req.Header = map[string][]string{
			common.FlorenceHeaderKey: {florenceToken},
		}
		responseRecorder := httptest.NewRecorder()

		httpClient := &commontest.RCHTTPClienterMock{
			SetAuthTokenFunc: func(string) {},
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusNotFound,
				}, nil
			},
		}
		idClient := clientsidentity.NewAPIClient(httpClient, zebedeeURL)

		handlerCalled := false
		httpHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			handlerCalled = true
		})

		identityHandler := HandlerForHTTPClient(idClient)(httpHandler)

		Convey("When ServeHTTP is called", func() {

			identityHandler.ServeHTTP(responseRecorder, req)

			Convey("Then the identity service is called as expected", func() {
				So(len(httpClient.DoCalls()), ShouldEqual, 1)
				So(httpClient.DoCalls()[0].Req.URL.String(), ShouldEqual, expectedZebedeeURL)
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

		req := httptest.NewRequest("GET", url, nil)
		req.Header = map[string][]string{
			common.FlorenceHeaderKey: {florenceToken},
		}
		responseRecorder := httptest.NewRecorder()

		httpClient := &commontest.RCHTTPClienterMock{
			SetAuthTokenFunc: func(string) {},
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {

				response := &common.IdentityResponse{Identifier: userIdentifier}

				body, _ := json.Marshal(response)
				readCloser := ioutil.NopCloser(bytes.NewBuffer(body))

				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       readCloser,
				}, nil
			},
		}
		idClient := clientsidentity.NewAPIClient(httpClient, zebedeeURL)

		handlerCalled := false
		var handlerReq *http.Request
		httpHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			handlerReq = req
			handlerCalled = true
		})

		identityHandler := HandlerForHTTPClient(idClient)(httpHandler)

		Convey("When ServeHTTP is called", func() {

			identityHandler.ServeHTTP(responseRecorder, req)

			Convey("Then the identity service is called as expected", func() {
				So(len(httpClient.DoCalls()), ShouldEqual, 1)
				zebedeeReq := httpClient.DoCalls()[0].Req
				So(zebedeeReq.URL.String(), ShouldEqual, expectedZebedeeURL)
				So(zebedeeReq.Header[common.FlorenceHeaderKey][0], ShouldEqual, florenceToken)
			})

			Convey("Then the downstream HTTP handler is called", func() {
				So(handlerCalled, ShouldBeTrue)
			})

			Convey("Then the downstream HTTP handler request has the expected context values", func() {
				So(handlerReq.Context().Value(common.CallerIdentityKey), ShouldEqual, userIdentifier)
				So(handlerReq.Context().Value(common.UserIdentityKey), ShouldEqual, userIdentifier)
			})
		})
	})
}

func TestHandler_InvalidIdentityResponse(t *testing.T) {

	Convey("Given a request with a florence token, and mock client that returns invalid response JSON", t, func() {

		req := httptest.NewRequest("GET", url, nil)
		req.Header = map[string][]string{
			common.FlorenceHeaderKey: {florenceToken},
		}
		responseRecorder := httptest.NewRecorder()

		httpClient := &commontest.RCHTTPClienterMock{
			SetAuthTokenFunc: func(string) {},
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {

				readCloser := ioutil.NopCloser(bytes.NewBufferString("{ invalid JSON"))

				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       readCloser,
				}, nil
			},
		}
		idClient := clientsidentity.NewAPIClient(httpClient, zebedeeURL)

		handlerCalled := false
		httpHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			handlerCalled = true
		})

		identityHandler := HandlerForHTTPClient(idClient)(httpHandler)

		Convey("When ServeHTTP is called", func() {

			identityHandler.ServeHTTP(responseRecorder, req)

			Convey("Then the identity service is called as expected", func() {
				So(len(httpClient.DoCalls()), ShouldEqual, 1)
				zebedeeReq := httpClient.DoCalls()[0].Req
				So(zebedeeReq.URL.String(), ShouldEqual, expectedZebedeeURL)
				So(zebedeeReq.Header[common.FlorenceHeaderKey][0], ShouldEqual, florenceToken)
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

		req := httptest.NewRequest("GET", url, nil)
		req.Header = map[string][]string{
			common.AuthHeaderKey: {upstreamAuthToken},
			common.UserHeaderKey: {userIdentifier},
		}
		responseRecorder := httptest.NewRecorder()

		httpClient := &commontest.RCHTTPClienterMock{
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {

				response := &common.IdentityResponse{Identifier: serviceIdentifier}

				body, _ := json.Marshal(response)
				readCloser := ioutil.NopCloser(bytes.NewBuffer(body))

				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       readCloser,
				}, nil
			},
		}
		idClient := clientsidentity.NewAPIClient(httpClient, zebedeeURL)

		handlerCalled := false
		var handlerReq *http.Request
		httpHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			handlerReq = req
			handlerCalled = true
		})

		identityHandler := HandlerForHTTPClient(idClient)(httpHandler)

		Convey("When ServeHTTP is called", func() {

			identityHandler.ServeHTTP(responseRecorder, req)

			Convey("Then the identity service is not called", func() {
				So(len(httpClient.DoCalls()), ShouldEqual, 1)
				zebedeeReq := httpClient.DoCalls()[0].Req
				So(zebedeeReq.URL.String(), ShouldEqual, expectedZebedeeURL)
				So(len(zebedeeReq.Header[common.UserHeaderKey]), ShouldEqual, 0)
				So(len(zebedeeReq.Header[common.AuthHeaderKey]), ShouldEqual, 1)
				So(zebedeeReq.Header[common.AuthHeaderKey][0], ShouldEqual, upstreamAuthToken)

			})

			Convey("Then the downstream HTTP handler is called", func() {
				So(handlerCalled, ShouldBeTrue)
			})

			Convey("Then the downstream HTTP handler request has the expected context values", func() {
				So(common.Caller(handlerReq.Context()), ShouldEqual, serviceIdentifier)
				So(common.User(handlerReq.Context()), ShouldEqual, userIdentifier)
			})
		})
	})
}

func TestHandler_bothTokens(t *testing.T) {

	Convey("Given a request with both a florence token and service token", t, func() {

		req := httptest.NewRequest("GET", url, nil)
		req.Header = map[string][]string{
			common.FlorenceHeaderKey:    {florenceToken},
			common.DeprecatedAuthHeader: {upstreamAuthToken},
		}
		responseRecorder := httptest.NewRecorder()

		httpClient := &commontest.RCHTTPClienterMock{
			SetAuthTokenFunc: func(string) {},
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {

				response := &common.IdentityResponse{Identifier: userIdentifier}

				body, _ := json.Marshal(response)
				readCloser := ioutil.NopCloser(bytes.NewBuffer(body))

				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       readCloser,
				}, nil
			},
		}
		idClient := clientsidentity.NewAPIClient(httpClient, zebedeeURL)

		handlerCalled := false
		var handlerReq *http.Request
		httpHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			handlerReq = req
			handlerCalled = true
		})

		identityHandler := HandlerForHTTPClient(idClient)(httpHandler)

		Convey("When ServeHTTP is called", func() {

			identityHandler.ServeHTTP(responseRecorder, req)

			Convey("Then the identity service is called as expected", func() {
				So(len(httpClient.DoCalls()), ShouldEqual, 1)
				zebedeeReq := httpClient.DoCalls()[0].Req
				So(zebedeeReq.URL.String(), ShouldEqual, expectedZebedeeURL)
				So(zebedeeReq.Header[common.FlorenceHeaderKey][0], ShouldEqual, florenceToken)
			})

			Convey("Then the downstream HTTP handler is called", func() {
				So(handlerCalled, ShouldBeTrue)
			})

			Convey("Then the downstream HTTP handler request has the expected context values", func() {
				So(handlerReq.Context().Value(common.UserIdentityKey), ShouldEqual, userIdentifier)
				So(handlerReq.Context().Value(common.CallerIdentityKey), ShouldEqual, userIdentifier)
			})
		})
	})
}
