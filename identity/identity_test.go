package identity

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/ONSdigital/dp-api-clients-go/headers"

	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"io"

	clientsidentity "github.com/ONSdigital/dp-api-clients-go/identity"
	nethttp "github.com/ONSdigital/dp-net/http"
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

		req := httptest.NewRequest("GET", url, bytes.NewBufferString("some body content"))
		responseRecorder := httptest.NewRecorder()

		httpClient := &nethttp.ClienterMock{}
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

			Convey("Then the request body has been drained", func() {
				_, err := req.Body.Read(make([]byte, 1))
				So(err, ShouldEqual, io.EOF)
			})
		})
	})
}

func TestHandler_IdentityServiceError(t *testing.T) {

	Convey("Given a request with a florence token, and mock client that returns an error", t, func() {

		req := httptest.NewRequest("GET", url, bytes.NewBufferString("some body content"))
		req.Header = map[string][]string{
			nethttp.FlorenceHeaderKey: {florenceToken},
		}
		responseRecorder := httptest.NewRecorder()

		httpClient := &nethttp.ClienterMock{
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

			Convey("Then the request body has been drained", func() {
				_, err := req.Body.Read(make([]byte, 1))
				So(err, ShouldEqual, io.EOF)
			})
		})
	})
}

func TestHandler_IdentityServiceErrorResponseCode(t *testing.T) {

	Convey("Given a request with a florence token, and mock client that returns a non-200 response", t, func() {

		req := httptest.NewRequest("GET", url, bytes.NewBufferString("some body content"))
		req.Header = map[string][]string{
			nethttp.FlorenceHeaderKey: {florenceToken},
		}
		responseRecorder := httptest.NewRecorder()

		httpClient := &nethttp.ClienterMock{
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

			Convey("Then the request body has been drained", func() {
				_, err := req.Body.Read(make([]byte, 1))
				So(err, ShouldEqual, io.EOF)
			})
		})
	})
}

func TestHandler_florenceToken(t *testing.T) {

	Convey("Given a request with a florence token, and mock client that returns 200", t, func() {

		req := httptest.NewRequest("GET", url, bytes.NewBufferString("some body content"))
		req.Header = map[string][]string{
			nethttp.FlorenceHeaderKey: {florenceToken},
		}
		responseRecorder := httptest.NewRecorder()

		httpClient := &nethttp.ClienterMock{
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {

				response := &nethttp.IdentityResponse{Identifier: userIdentifier}

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
				So(zebedeeReq.Header[nethttp.FlorenceHeaderKey][0], ShouldEqual, florenceToken)
			})

			Convey("Then the downstream HTTP handler is called", func() {
				So(handlerCalled, ShouldBeTrue)
			})

			Convey("Then the downstream HTTP handler request has the expected context values", func() {
				So(handlerReq.Context().Value(nethttp.CallerIdentityKey), ShouldEqual, userIdentifier)
				So(handlerReq.Context().Value(nethttp.UserIdentityKey), ShouldEqual, userIdentifier)
			})

			Convey("Then the request body has not been drained", func() {
				_, err := req.Body.Read(make([]byte, 1))
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestHandler_InvalidIdentityResponse(t *testing.T) {

	Convey("Given a request with a florence token, and mock client that returns invalid response JSON", t, func() {

		req := httptest.NewRequest("GET", url, bytes.NewBufferString("some body content"))
		req.Header = map[string][]string{
			nethttp.FlorenceHeaderKey: {florenceToken},
		}
		responseRecorder := httptest.NewRecorder()

		httpClient := &nethttp.ClienterMock{
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
				So(zebedeeReq.Header[nethttp.FlorenceHeaderKey][0], ShouldEqual, florenceToken)
			})

			Convey("Then the response code is set as expected", func() {
				So(responseRecorder.Code, ShouldEqual, http.StatusInternalServerError)
			})

			Convey("Then the downstream HTTP handler is not called", func() {
				So(handlerCalled, ShouldBeFalse)
			})

			Convey("Then the request body has been drained", func() {
				_, err := req.Body.Read(make([]byte, 1))
				So(err, ShouldEqual, io.EOF)
			})
		})
	})
}

func TestHandler_authToken(t *testing.T) {

	Convey("Given a request with an auth token, and mock client that returns 200", t, func() {

		req := httptest.NewRequest("GET", url, bytes.NewBufferString("some body content"))
		req.Header = map[string][]string{
			nethttp.AuthHeaderKey: {upstreamAuthToken},
			nethttp.UserHeaderKey: {userIdentifier},
		}
		responseRecorder := httptest.NewRecorder()

		httpClient := &nethttp.ClienterMock{
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {

				response := &nethttp.IdentityResponse{Identifier: serviceIdentifier}

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
				So(len(zebedeeReq.Header[nethttp.UserHeaderKey]), ShouldEqual, 0)
				So(len(zebedeeReq.Header[nethttp.AuthHeaderKey]), ShouldEqual, 1)
				So(zebedeeReq.Header[nethttp.AuthHeaderKey][0], ShouldEqual, "Bearer "+upstreamAuthToken)

			})

			Convey("Then the downstream HTTP handler is called", func() {
				So(handlerCalled, ShouldBeTrue)
			})

			Convey("Then the downstream HTTP handler request has the expected context values", func() {
				So(nethttp.Caller(handlerReq.Context()), ShouldEqual, serviceIdentifier)
				So(nethttp.User(handlerReq.Context()), ShouldEqual, userIdentifier)
			})

			Convey("Then the request body has not been drained", func() {
				_, err := req.Body.Read(make([]byte, 1))
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestHandler_bothTokens(t *testing.T) {

	Convey("Given a request with both a florence token and service token", t, func() {

		req := httptest.NewRequest("GET", url, bytes.NewBufferString("some body content"))
		req.Header = map[string][]string{
			nethttp.FlorenceHeaderKey:    {florenceToken},
			nethttp.DeprecatedAuthHeader: {upstreamAuthToken},
		}
		responseRecorder := httptest.NewRecorder()

		httpClient := &nethttp.ClienterMock{
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {

				response := &nethttp.IdentityResponse{Identifier: userIdentifier}

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
				So(zebedeeReq.Header[nethttp.FlorenceHeaderKey][0], ShouldEqual, florenceToken)
			})

			Convey("Then the downstream HTTP handler is called", func() {
				So(handlerCalled, ShouldBeTrue)
			})

			Convey("Then the downstream HTTP handler request has the expected context values", func() {
				So(handlerReq.Context().Value(nethttp.UserIdentityKey), ShouldEqual, userIdentifier)
				So(handlerReq.Context().Value(nethttp.CallerIdentityKey), ShouldEqual, userIdentifier)
			})

			Convey("Then the request body has not been drained", func() {
				_, err := req.Body.Read(make([]byte, 1))
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestHandler_GetTokenError(t *testing.T) {

	Convey("Given getting the user auth token from the request returns an error", t, func() {

		httpClient := &nethttp.ClienterMock{}
		idClient := clientsidentity.NewAPIClient(httpClient, zebedeeURL)

		handlerCalled := false
		wrappedHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			handlerCalled = true
		})

		getUserTokenCalls := 0
		getUserTokenFunc := func(ctx context.Context, r *http.Request) (string, error) {
			getUserTokenCalls++
			return "", errors.New("bork bork bork")
		}

		getServiceTokenCalls := 0
		getServiceTokenFunc := func(ctx context.Context, r *http.Request) (string, error) {
			getServiceTokenCalls++
			return "", nil
		}

		identityHandler := handlerForHTTPClient(idClient, getUserTokenFunc, getServiceTokenFunc)(wrappedHandler)

		Convey("when a request is received", func() {
			req := httptest.NewRequest("GET", url, bytes.NewBufferString("some body content"))
			resp := httptest.NewRecorder()

			identityHandler.ServeHTTP(resp, req)

			Convey("then a status 500 internal server error response is returned", func() {
				So(resp.Code, ShouldEqual, http.StatusInternalServerError)
			})

			Convey("and get getUserTokenFunc is called 1 time", func() {
				So(getUserTokenCalls, ShouldEqual, 1)
			})

			Convey("and the request is not processed any further", func() {
				So(getServiceTokenCalls, ShouldEqual, 0)
				So(httpClient.DoCalls(), ShouldHaveLength, 0)
				So(handlerCalled, ShouldBeFalse)
			})
		})
	})

	Convey("Given getting the service auth token from the request returns an error", t, func() {

		httpClient := &nethttp.ClienterMock{}
		idClient := clientsidentity.NewAPIClient(httpClient, zebedeeURL)

		handlerCalled := false
		wrappedHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			handlerCalled = true
		})

		getUserTokenCalls := 0
		getUserTokenFunc := func(ctx context.Context, r *http.Request) (string, error) {
			getUserTokenCalls++
			return "1234", nil
		}

		getServiceTokenCalls := 0
		getServiceTokenFunc := func(ctx context.Context, r *http.Request) (string, error) {
			getServiceTokenCalls++
			return "", errors.New("aaaaaaaallllll righty then")
		}

		identityHandler := handlerForHTTPClient(idClient, getUserTokenFunc, getServiceTokenFunc)(wrappedHandler)

		Convey("when a request is received", func() {
			req := httptest.NewRequest("GET", url, bytes.NewBufferString("some body content"))
			resp := httptest.NewRecorder()

			identityHandler.ServeHTTP(resp, req)

			Convey("then a status 500 internal server error response is returned", func() {
				So(resp.Code, ShouldEqual, http.StatusInternalServerError)
			})

			Convey("and get getUserTokenFunc is called 1 time", func() {
				So(getUserTokenCalls, ShouldEqual, 1)
			})

			Convey("and get getServiceTokenFunc is called 1 time", func() {
				So(getServiceTokenCalls, ShouldEqual, 1)
			})

			Convey("and the request is not processed any further", func() {
				So(httpClient.DoCalls(), ShouldHaveLength, 0)
				So(handlerCalled, ShouldBeFalse)
			})
		})
	})

}

func TestGetFlorenceToken(t *testing.T) {
	expectedToken := "666"

	Convey("should return florence token from request header", t, func() {
		req := httptest.NewRequest("GET", "http://localhost:8080", nil)
		req.Header.Set(nethttp.FlorenceHeaderKey, expectedToken)

		actual, err := getFlorenceToken(nil, req)

		So(actual, ShouldEqual, expectedToken)
		So(err, ShouldBeNil)
	})

	Convey("should return access token from request cookie", t, func() {
		req := httptest.NewRequest("GET", "http://localhost:8080", nil)
		req.AddCookie(&http.Cookie{Name: nethttp.FlorenceCookieKey, Value: expectedToken})

		actual, err := getFlorenceToken(nil, req)

		So(actual, ShouldEqual, expectedToken)
		So(err, ShouldBeNil)
	})

	Convey("should return empty token if no header or cookie is set", t, func() {
		req := httptest.NewRequest("GET", "http://localhost:8080", nil)

		actual, err := getFlorenceToken(nil, req)

		So(actual, ShouldBeEmpty)
		So(err, ShouldBeNil)
	})

	Convey("should return empty token and error if get header returns an error that is not ErrHeaderNotFound", t, func() {
		actual, err := getFlorenceToken(nil, nil)

		So(actual, ShouldBeEmpty)
		So(err, ShouldResemble, headers.ErrRequestNil)
	})
}

func TestGetFlorenceTokenFromCookie(t *testing.T) {
	expectedToken := "666"

	Convey("should return florence token from request cookie", t, func() {
		req := httptest.NewRequest("GET", "http://localhost:8080", nil)
		req.AddCookie(&http.Cookie{Name: nethttp.FlorenceCookieKey, Value: expectedToken})

		actual, err := getFlorenceTokenFromCookie(nil, req)

		So(actual, ShouldEqual, expectedToken)
		So(err, ShouldBeNil)
	})

	Convey("should return empty token if token cookie not found", t, func() {
		req := httptest.NewRequest("GET", "http://localhost:8080", nil)

		actual, err := getFlorenceTokenFromCookie(nil, req)

		So(actual, ShouldBeEmpty)
		So(err, ShouldBeNil)
	})
}

func TestGetServiceAuthToken(t *testing.T) {
	Convey("should return error if not equal to headers.ErrHeaderNotFound", t, func() {
		token, err := getServiceAuthToken(nil, nil)

		So(token, ShouldBeEmpty)
		So(err, ShouldResemble, headers.ErrRequestNil)
	})

	Convey("should return empty token if error equals headers.ErrHeaderNotFound", t, func() {
		req := httptest.NewRequest("GET", "http://localhost:8080", nil)

		token, err := getServiceAuthToken(nil, req)

		So(token, ShouldBeEmpty)
		So(err, ShouldBeNil)
	})

	Convey("should return token if header found", t, func() {
		req := httptest.NewRequest("GET", "http://localhost:8080", nil)
		headers.SetServiceAuthToken(req, "666")

		token, err := getServiceAuthToken(nil, req)

		So(token, ShouldEqual, "666")
		So(err, ShouldBeNil)
	})
}

func getTokenFunc(token string, err error) getTokenFromReqFunc {
	return func(ctx context.Context, r *http.Request) (string, error) {
		return token, err
	}
}
