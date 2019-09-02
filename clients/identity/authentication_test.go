package identity

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ONSdigital/go-ns/common"
	"github.com/ONSdigital/go-ns/common/commontest"
	"github.com/ONSdigital/go-ns/log"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/ONSdigital/dp-rchttp"
)

const (
	url                = "/whatever"
	florenceToken      = "roundabout"
	callerAuthToken    = "YourClaimToBeWhoYouAre"
	callerIdentifier   = "externalCaller"
	userIdentifier     = "fred@ons.gov.uk"
	zebedeeURL         = "http://localhost:8082"
	expectedZebedeeURL = zebedeeURL + "/identity"
)

func TestHandler_NoAuth(t *testing.T) {

	Convey("Given a request with no auth headers", t, func() {

		req := httptest.NewRequest("GET", url, nil)
		httpClient := &rchttp.ClienterMock{}
		idClient := NewAPIClient(httpClient, zebedeeURL)

		Convey("When CheckRequest is called", func() {

			ctx, status, authFailure, err := idClient.CheckRequest(req, "", "")

			Convey("Then the downstream HTTP handler should not be called", func() {
				So(len(httpClient.DoCalls()), ShouldEqual, 0)
				So(err, ShouldBeNil)
				So(authFailure.Error(), ShouldContainSubstring, "no headers set on request: "+errUnableToIdentifyRequest.Error())
				So(ctx, ShouldNotBeNil)
				So(common.IsUserPresent(ctx), ShouldBeFalse)
				So(common.IsCallerPresent(ctx), ShouldBeFalse)
			})

			Convey("Then the returned code should be 401", func() {
				So(status, ShouldEqual, http.StatusUnauthorized)
			})
		})
	})
}

func TestHandler_IdentityServiceError(t *testing.T) {

	Convey("Given a request with a florence token, and a mock client that returns an error", t, func() {

		req := httptest.NewRequest("GET", url, nil)

		expectedError := errors.New("broken")
		httpClient := getClientReturningError(expectedError)
		idClient := NewAPIClient(httpClient, zebedeeURL)

		Convey("When CheckRequest is called", func() {

			ctx, status, authFailure, err := idClient.CheckRequest(req, florenceToken, "")

			Convey("Then the identity service was called as expected", func() {
				So(len(httpClient.DoCalls()), ShouldEqual, 1)
				So(httpClient.DoCalls()[0].Req.URL.String(), ShouldEqual, expectedZebedeeURL)
			})

			Convey("Then the error and no context is returned", func() {
				So(authFailure, ShouldBeNil)
				So(err, ShouldEqual, expectedError)
				So(status, ShouldNotEqual, http.StatusOK)
				So(ctx, ShouldNotBeNil)
				So(common.IsUserPresent(ctx), ShouldBeFalse)
				So(common.IsCallerPresent(ctx), ShouldBeFalse)
			})
		})
	})
}

func TestHandler_IdentityServiceErrorResponseCode(t *testing.T) {

	Convey("Given a request with a florence token, and mock client that returns a non-200 response", t, func() {

		req := httptest.NewRequest("GET", url, nil)

		httpClient := &rchttp.ClienterMock{
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusNotFound,
				}, nil
			},
		}
		idClient := NewAPIClient(httpClient, zebedeeURL)

		Convey("When CheckRequest is called", func() {

			ctx, status, authFailure, err := idClient.CheckRequest(req, florenceToken, "")

			Convey("Then the identity service is called as expected", func() {
				So(len(httpClient.DoCalls()), ShouldEqual, 1)
				So(httpClient.DoCalls()[0].Req.URL.String(), ShouldEqual, expectedZebedeeURL)
			})

			Convey("Then there is no error but the response code matches the identity service", func() {
				So(authFailure.Error(), ShouldContainSubstring, "unexpected status code returned from AuthAPI: "+errUnableToIdentifyRequest.Error())
				So(err, ShouldBeNil)
				So(status, ShouldEqual, http.StatusNotFound)
				So(ctx, ShouldNotBeNil)
				So(common.IsUserPresent(ctx), ShouldBeFalse)
				So(common.IsCallerPresent(ctx), ShouldBeFalse)
			})
		})
	})
}

func TestHandler_florenceToken(t *testing.T) {

	Convey("Given a request with a florence token, and mock client that returns 200", t, func() {

		req := httptest.NewRequest("GET", url, nil)

		httpClient := getClientReturningIdentifier(userIdentifier)
		idClient := NewAPIClient(httpClient, zebedeeURL)

		Convey("When CheckRequest is called", func() {

			ctx, status, authFailure, err := idClient.CheckRequest(req, florenceToken, "")

			Convey("Then the identity service is called as expected", func() {
				So(authFailure, ShouldBeNil)
				So(err, ShouldBeNil)
				So(status, ShouldEqual, http.StatusOK)
				So(len(httpClient.DoCalls()), ShouldEqual, 1)
				zebedeeReq := httpClient.DoCalls()[0].Req
				So(zebedeeReq.URL.String(), ShouldEqual, expectedZebedeeURL)
				So(zebedeeReq.Header[common.FlorenceHeaderKey][0], ShouldEqual, florenceToken)
			})

			Convey("Then the downstream HTTP handler returned no error and expected context", func() {
				So(ctx, ShouldNotBeNil)
				So(common.Caller(ctx), ShouldEqual, userIdentifier)
				So(common.User(ctx), ShouldEqual, userIdentifier)
			})
		})
	})

	Convey("Given a request with a florence token as a cookie, and mock client that returns 200", t, func() {
		req := httptest.NewRequest("GET", url, nil)

		httpClient := getClientReturningIdentifier(userIdentifier)
		idClient := NewAPIClient(httpClient, zebedeeURL)

		Convey("When CheckRequest is called", func() {

			ctx, status, authFailure, err := idClient.CheckRequest(req, florenceToken, "")

			Convey("Then the identity service is called as expected", func() {
				So(authFailure, ShouldBeNil)
				So(err, ShouldBeNil)
				So(status, ShouldEqual, http.StatusOK)
				So(len(httpClient.DoCalls()), ShouldEqual, 1)
				zebedeeReq := httpClient.DoCalls()[0].Req
				So(zebedeeReq.URL.String(), ShouldEqual, expectedZebedeeURL)
				So(zebedeeReq.Header[common.FlorenceHeaderKey][0], ShouldEqual, florenceToken)
			})

			Convey("Then the downstream HTTP handler returned no error and expected context", func() {
				So(ctx, ShouldNotBeNil)
				So(common.Caller(ctx), ShouldEqual, userIdentifier)
				So(common.User(ctx), ShouldEqual, userIdentifier)
			})
		})
	})
}

func TestHandler_InvalidIdentityResponse(t *testing.T) {

	Convey("Given a request with a florence token, and mock client that returns invalid response JSON", t, func() {

		req := httptest.NewRequest("GET", url, nil)

		httpClient := &commontest.RCHTTPClienterMock{
			SetAuthTokenFunc: func(string) {},
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(bytes.NewBufferString("{ invalid JSON")),
				}, nil
			},
		}
		idClient := NewAPIClient(httpClient, zebedeeURL)

		Convey("When CheckRequest is called", func() {

			ctx, status, authFailure, err := idClient.CheckRequest(req, florenceToken, "")

			Convey("Then the identity service is called as expected", func() {
				So(len(httpClient.DoCalls()), ShouldEqual, 1)
				zebedeeReq := httpClient.DoCalls()[0].Req
				So(zebedeeReq.URL.String(), ShouldEqual, expectedZebedeeURL)
				So(zebedeeReq.Header[common.FlorenceHeaderKey][0], ShouldEqual, florenceToken)
			})

			Convey("Then the response is set as expected", func() {
				So(authFailure, ShouldBeNil)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "invalid character 'i' looking for beginning of object key string")
				So(status, ShouldEqual, http.StatusInternalServerError)
				So(ctx, ShouldNotBeNil)
				So(common.Caller(ctx), ShouldBeEmpty)
				So(common.User(ctx), ShouldBeEmpty)
			})
		})
	})
}

func TestHandler_authToken(t *testing.T) {

	Convey("Given a request with an auth token, and mock client that returns 200", t, func() {

		req := httptest.NewRequest("GET", url, nil)
		req.Header = map[string][]string{
			common.UserHeaderKey: {userIdentifier},
		}

		httpClient := getClientReturningIdentifier(callerIdentifier)
		idClient := NewAPIClient(httpClient, zebedeeURL)

		Convey("When CheckRequest is called", func() {

			ctx, status, authFailure, err := idClient.CheckRequest(req, "", callerAuthToken)
			So(err, ShouldBeNil)
			So(authFailure, ShouldBeNil)
			So(status, ShouldEqual, http.StatusOK)

			Convey("Then the identity service is called as expected", func() {
				So(len(httpClient.DoCalls()), ShouldEqual, 1)
				zebedeeReq := httpClient.DoCalls()[0].Req
				So(zebedeeReq.URL.String(), ShouldEqual, expectedZebedeeURL)
				So(len(zebedeeReq.Header[common.UserHeaderKey]), ShouldEqual, 0)
				So(len(zebedeeReq.Header[common.AuthHeaderKey]), ShouldEqual, 1)
				So(zebedeeReq.Header[common.AuthHeaderKey][0], ShouldEqual, callerAuthToken)
			})

			Convey("Then the downstream HTTP handler request has the expected context values", func() {
				So(len(httpClient.DoCalls()), ShouldEqual, 1)

				So(ctx, ShouldNotBeNil)
				So(common.IsCallerPresent(ctx), ShouldBeTrue)
				So(common.IsUserPresent(ctx), ShouldBeTrue)
				So(common.Caller(ctx), ShouldEqual, callerIdentifier)
				So(common.User(ctx), ShouldEqual, userIdentifier)
			})
		})
	})
}

func TestHandler_bothTokens(t *testing.T) {

	Convey("Given a request with both a florence token and service token", t, func() {

		req := httptest.NewRequest("GET", url, nil)
		req.Header = map[string][]string{
			common.FlorenceHeaderKey: {florenceToken},
			common.AuthHeaderKey:     {callerAuthToken},
		}

		httpClient := getClientReturningIdentifier(userIdentifier)
		idClient := NewAPIClient(httpClient, zebedeeURL)

		Convey("When CheckRequest is called", func() {

			ctx, status, authFailure, err := idClient.CheckRequest(req, florenceToken, callerAuthToken)
			So(err, ShouldBeNil)
			So(authFailure, ShouldBeNil)
			So(status, ShouldEqual, http.StatusOK)

			Convey("Then the identity service is called as expected - verifying florence, but ignoring auth header", func() {
				So(len(httpClient.DoCalls()), ShouldEqual, 1)
				zebedeeReq := httpClient.DoCalls()[0].Req
				So(zebedeeReq.URL.String(), ShouldEqual, expectedZebedeeURL)
				So(zebedeeReq.Header[common.FlorenceHeaderKey][0], ShouldEqual, florenceToken)
				So(len(zebedeeReq.Header[common.AuthHeaderKey]), ShouldEqual, 0)
			})

			Convey("Then the context returns with expected values", func() {
				So(ctx, ShouldNotBeNil)
				So(common.IsUserPresent(ctx), ShouldBeTrue)
				So(common.User(ctx), ShouldEqual, userIdentifier)
				So(common.Caller(ctx), ShouldEqual, userIdentifier)
			})
		})
	})
}

func TestSplitTokens(t *testing.T) {
	Convey("Given a service token and an empty florence token", t, func() {
		florenceToken := ""
		serviceToken := "Bearer 123456789"

		Convey("When we pass both tokens into splitTokens function", func() {
			logData := log.Data{}
			splitTokens(florenceToken, serviceToken, logData)

			Convey("Then the token objects are returned with the expected values", func() {
				So(logData["auth_token"], ShouldResemble, tokenObject{numberOfParts: 2, hasPrefix: true, tokenPart: "456789"})
				So(logData["florence_token"], ShouldBeNil)
			})
		})
	})

	Convey("Given a florence token and an empty service token", t, func() {
		florenceToken := "987654321"
		serviceToken := ""

		Convey("When we pass both tokens into splitTokens function", func() {
			logData := log.Data{}
			splitTokens(florenceToken, serviceToken, logData)

			Convey("Then the token objects are returned with the expected values", func() {
				So(logData["florence_token"], ShouldResemble, tokenObject{numberOfParts: 1, hasPrefix: false, tokenPart: "654321"})
				So(logData["auth_token"], ShouldBeNil)
			})
		})
	})

	Convey("Given a florence token and service token", t, func() {
		florenceToken := "987654321"
		serviceToken := "Bearer 123456789"

		Convey("When we pass both tokens into splitTokens function", func() {
			logData := log.Data{}
			splitTokens(florenceToken, serviceToken, logData)

			Convey("Then the token objects are returned with the expected values", func() {
				So(logData["florence_token"], ShouldResemble, tokenObject{numberOfParts: 1, hasPrefix: false, tokenPart: "654321"})
				So(logData["auth_token"], ShouldResemble, tokenObject{numberOfParts: 2, hasPrefix: true, tokenPart: "456789"})
			})
		})
	})

	Convey("Given a small service token", t, func() {
		florenceToken := "54321"
		serviceToken := "Bearer A 12"

		Convey("When we pass the tokens into splitTokens function", func() {
			logData := log.Data{}
			splitTokens(florenceToken, serviceToken, logData)

			Convey("Then the token objects are returned with the expected values", func() {
				So(logData["florence_token"], ShouldResemble, tokenObject{numberOfParts: 1, hasPrefix: false, tokenPart: "321"})
				So(logData["auth_token"], ShouldResemble, tokenObject{numberOfParts: 3, hasPrefix: true, tokenPart: "2"})
			})
		})
	})

}

func getClientReturningIdentifier(id string) *rchttp.ClienterMock {
	return &rchttp.ClienterMock{
		DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
			response := &common.IdentityResponse{Identifier: id}
			body, _ := json.Marshal(response)
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(bytes.NewBuffer(body)),
			}, nil
		},
	}
}
func getClientReturningError(err error) *commontest.RCHTTPClienterMock {
	return &commontest.RCHTTPClienterMock{
		DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
			return nil, err
		},
	}
}
