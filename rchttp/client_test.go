package rchttp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ONSdigital/go-ns/common"
	"github.com/ONSdigital/go-ns/rchttp/rchttptest"
	. "github.com/smartystreets/goconvey/convey"
)

func TestHappyPaths(t *testing.T) {
	ts := rchttptest.NewTestServer()
	defer ts.Close()
	expectedCallCount := 0

	Convey("Given a default rchttp client and happy paths", t, func() {
		httpClient := NewClient()

		Convey("When Get() is called on a URL", func() {
			expectedCallCount++
			resp, err := httpClient.Get(context.Background(), ts.URL)
			So(resp, ShouldNotBeNil)
			So(err, ShouldBeNil)

			call, err := unmarshallResp(resp)
			So(err, ShouldBeNil)

			Convey("Then the server sees a GET with no body", func() {
				So(call.CallCount, ShouldEqual, expectedCallCount)
				So(call.Method, ShouldEqual, "GET")
				So(call.Body, ShouldEqual, "")
				So(call.Error, ShouldEqual, "")
				So(resp.Header.Get("Content-Type"), ShouldContainSubstring, "text/plain")
			})
		})

		Convey("When Post() is called on a URL", func() {
			expectedCallCount++
			resp, err := httpClient.Post(context.Background(), ts.URL, rchttptest.JsonContentType, strings.NewReader(`{"dummy":"ook"}`))
			So(resp, ShouldNotBeNil)
			So(err, ShouldBeNil)

			call, err := unmarshallResp(resp)
			So(err, ShouldBeNil)

			Convey("Then the server sees a POST with that body as JSON", func() {
				So(call.CallCount, ShouldEqual, expectedCallCount)
				So(call.Method, ShouldEqual, "POST")
				So(call.Body, ShouldEqual, `{"dummy":"ook"}`)
				So(call.Error, ShouldEqual, "")
				So(call.Headers[rchttptest.ContentTypeHeader], ShouldResemble, []string{rchttptest.JsonContentType})
			})
		})

		Convey("When Put() is called on a URL", func() {
			expectedCallCount++
			resp, err := httpClient.Put(context.Background(), ts.URL, rchttptest.JsonContentType, strings.NewReader(`{"dummy":"ook2"}`))
			So(resp, ShouldNotBeNil)
			So(err, ShouldBeNil)

			call, err := unmarshallResp(resp)
			So(err, ShouldBeNil)

			Convey("Then the server sees a PUT with that body as JSON", func() {
				So(call.CallCount, ShouldEqual, expectedCallCount)
				So(call.Method, ShouldEqual, "PUT")
				So(call.Body, ShouldEqual, `{"dummy":"ook2"}`)
				So(call.Error, ShouldEqual, "")
				So(call.Headers[rchttptest.ContentTypeHeader], ShouldResemble, []string{rchttptest.JsonContentType})
			})
		})

		Convey("When PostForm() is called on a URL", func() {
			expectedCallCount++
			resp, err := httpClient.PostForm(context.Background(), ts.URL, url.Values{"ook": {"koo"}, "zoo": {"ooz"}})
			So(resp, ShouldNotBeNil)
			So(err, ShouldBeNil)

			call, err := unmarshallResp(resp)
			So(err, ShouldBeNil)

			Convey("Then the server sees a POST with those values encoded", func() {
				So(call.CallCount, ShouldEqual, expectedCallCount)
				So(call.Method, ShouldEqual, "POST")
				So(call.Body, ShouldEqual, "ook=koo&zoo=ooz")
				So(call.Error, ShouldEqual, "")
				So(call.Headers[rchttptest.ContentTypeHeader], ShouldResemble, []string{rchttptest.FormEncodedType})
			})
		})
	})
}

func TestClientDoesRetry(t *testing.T) {
	ts := rchttptest.NewTestServer()
	defer ts.Close()
	expectedCallCount := 0

	Convey("Given an rchttp client with small client timeout", t, func() {
		// force client to abandon requests before the requested one second delay on the (next) server response
		httpClient := ClientWithTimeout(nil, 100*time.Millisecond)

		Convey("When Post() is called on a URL with a delay on the first response", func() {
			/// XXX this is two for the retry due to the delayed first POST
			delayByOneSecondOnNext := delayByOneSecondOn(expectedCallCount + 1)
			expectedCallCount += 2
			resp, err := httpClient.Post(context.Background(), ts.URL, rchttptest.JsonContentType, strings.NewReader(delayByOneSecondOnNext))
			So(resp, ShouldNotBeNil)
			So(err, ShouldBeNil)

			call, err := unmarshallResp(resp)
			So(err, ShouldBeNil)

			Convey("Then the server sees two POST calls", func() {
				So(call.CallCount, ShouldEqual, expectedCallCount)
				So(call.Method, ShouldEqual, "POST")
				So(call.Body, ShouldEqual, delayByOneSecondOnNext)
				So(call.Error, ShouldEqual, "")
				So(resp.Header.Get(rchttptest.ContentTypeHeader), ShouldContainSubstring, "text/plain")
			})
		})
	})
}

func TestClientNoRetries(t *testing.T) {
	ts := rchttptest.NewTestServer()
	defer ts.Close()
	expectedCallCount := 0

	Convey("Given an rchttp client with no retries", t, func() {
		httpClient := ClientWithTimeout(nil, 100*time.Millisecond)
		httpClient.SetMaxRetries(0)

		Convey("When Post() is called on a URL with a delay on the first call", func() {
			delayByOneSecondOnNext := delayByOneSecondOn(expectedCallCount + 1)
			resp, err := httpClient.Post(context.Background(), ts.URL, rchttptest.JsonContentType, strings.NewReader(delayByOneSecondOnNext))
			Convey("Then the server has no response", func() {
				So(resp, ShouldBeNil)
				So(err.Error(), ShouldContainSubstring, "Timeout exceeded")
			})
		})
	})
}

func TestClientWithServiceTokenHasAuthHeader(t *testing.T) {
	ts := rchttptest.NewTestServer()
	defer ts.Close()
	expectedCallCount := 0

	Convey("Given an rchttp client with an auth token", t, func() {
		expectedAuthToken := "IAmWhoIAm"
		// throw in a check for wrapped client instantiation
		httpClient := ClientWithServiceToken(ClientWithTimeout(nil, 5*time.Second), expectedAuthToken)

		Convey("When Post() is called on a URL", func() {
			expectedCallCount++
			resp, err := httpClient.Post(context.Background(), ts.URL, rchttptest.JsonContentType, strings.NewReader(`{"hello":"there"}`))
			So(resp, ShouldNotBeNil)
			So(err, ShouldBeNil)

			call, err := unmarshallResp(resp)
			So(err, ShouldBeNil)

			Convey("Then the server sees the auth header", func() {
				So(call.CallCount, ShouldEqual, expectedCallCount)
				So(call.Method, ShouldEqual, "POST")
				So(call.Body, ShouldEqual, `{"hello":"there"}`)
				So(call.Error, ShouldEqual, "")
				So(call.Headers[common.AuthHeaderKey], ShouldResemble, []string{common.BearerPrefix + expectedAuthToken})
			})
		})
	})
}

// end of tests //

// delayByOneSecondOn returns the json which will instruct the server to delay responding on call-number `delayOnCall`
func delayByOneSecondOn(delayOnCall int) string {
	return `{"delay":"1s","delay_on_call":` + strconv.Itoa(delayOnCall) + `}`
}

func unmarshallResp(resp *http.Response) (*rchttptest.Responder, error) {
	responder := &rchttptest.Responder{}
	body := rchttptest.GetBody(resp.Body)
	err := json.Unmarshal(body, responder)
	if err != nil {
		panic(err.Error() + string(body))
	}
	return responder, err
}
