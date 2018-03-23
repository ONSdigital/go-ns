package rchttp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

const (
	jsonContentType = "application/json"
	formEncodedType = "application/x-www-form-urlencoded"
)

type TestServer struct {
	ts  *httptest.Server
	url string
}

type Responder struct {
	Body        string `json:"body"`
	CallCount   int    `json:"call_count"`
	ContentType string `json:"content_type"`
	Method      string `json:"method"`
	Error       string `json:"error"`
}

type RequestTester struct {
	Delay         string `json:"delay"`
	DelayDuration time.Duration
	DelayOnCall   int `json:"delay_on_call"`
}

func TestHappyPaths(t *testing.T) {
	ts := NewTestServer()
	defer ts.Close()
	expectedCallCount := 0

	Convey("Given a default rchttp client and happy paths", t, func() {
		httpClient := DefaultClient

		Convey("When Get() is called on a URL", func() {
			expectedCallCount++
			resp, err := httpClient.Get(context.Background(), ts.url)
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
			resp, err := httpClient.Post(context.Background(), ts.url, jsonContentType, strings.NewReader(`{"dummy":"ook"}`))
			So(resp, ShouldNotBeNil)
			So(err, ShouldBeNil)

			call, err := unmarshallResp(resp)
			So(err, ShouldBeNil)

			Convey("Then the server sees a POST with that body as JSON", func() {
				So(call.CallCount, ShouldEqual, expectedCallCount)
				So(call.Method, ShouldEqual, "POST")
				So(call.Body, ShouldEqual, `{"dummy":"ook"}`)
				So(call.Error, ShouldEqual, "")
				So(call.ContentType, ShouldEqual, jsonContentType)
			})
		})

		Convey("When Put() is called on a URL", func() {
			expectedCallCount++
			resp, err := httpClient.Put(context.Background(), ts.url, jsonContentType, strings.NewReader(`{"dummy":"ook2"}`))
			So(resp, ShouldNotBeNil)
			So(err, ShouldBeNil)

			call, err := unmarshallResp(resp)
			So(err, ShouldBeNil)

			Convey("Then the server sees a PUT with that body as JSON", func() {
				So(call.CallCount, ShouldEqual, expectedCallCount)
				So(call.Method, ShouldEqual, "PUT")
				So(call.Body, ShouldEqual, `{"dummy":"ook2"}`)
				So(call.Error, ShouldEqual, "")
				So(call.ContentType, ShouldEqual, jsonContentType)
			})
		})

		Convey("When PostForm() is called on a URL", func() {
			expectedCallCount++
			resp, err := httpClient.PostForm(context.Background(), ts.url, url.Values{"ook": {"koo"}, "zoo": {"ooz"}})
			So(resp, ShouldNotBeNil)
			So(err, ShouldBeNil)

			call, err := unmarshallResp(resp)
			So(err, ShouldBeNil)

			Convey("Then the server sees a POST with those values encoded", func() {
				So(call.CallCount, ShouldEqual, expectedCallCount)
				So(call.Method, ShouldEqual, "POST")
				So(call.Body, ShouldEqual, "ook=koo&zoo=ooz")
				So(call.Error, ShouldEqual, "")
				So(call.ContentType, ShouldEqual, formEncodedType)
			})
		})
	})
}

func TestClientDoesRetry(t *testing.T) {
	ts := NewTestServer()
	defer ts.Close()
	expectedCallCount := 0

	Convey("Given an rchttp client with small client timeout", t, func() {
		// force client to abandon requests before the requested one second delay on the (next) server response
		httpClient := ClientWithTimeout(100 * time.Millisecond)

		Convey("When Post() is called on a URL with a delay on the first response", func() {
			/// XXX this is two for the retry due to the delayed first POST
			delayByOneSecondOnNext := delayByOneSecondOn(expectedCallCount + 1)
			expectedCallCount += 2
			resp, err := httpClient.Post(context.Background(), ts.url, jsonContentType, strings.NewReader(delayByOneSecondOnNext))
			So(resp, ShouldNotBeNil)
			So(err, ShouldBeNil)

			call, err := unmarshallResp(resp)
			So(err, ShouldBeNil)

			Convey("Then the server sees two POST calls", func() {
				So(call.CallCount, ShouldEqual, expectedCallCount)
				So(call.Method, ShouldEqual, "POST")
				So(call.Body, ShouldEqual, delayByOneSecondOnNext)
				So(call.Error, ShouldEqual, "")
				So(resp.Header.Get("Content-Type"), ShouldContainSubstring, "text/plain")
			})
		})
	})
}

func TestClientNoRetries(t *testing.T) {
	ts := NewTestServer()
	defer ts.Close()
	expectedCallCount := 0

	Convey("Given an rchttp client with a no retries", t, func() {
		httpClient := ClientWithTimeout(100 * time.Millisecond)
		httpClient.MaxRetries = 0

		Convey("When Post() is called on a URL with a delay on the first call", func() {
			delayByOneSecondOnNext := delayByOneSecondOn(expectedCallCount + 1)
			resp, err := httpClient.Post(context.Background(), ts.url, jsonContentType, strings.NewReader(delayByOneSecondOnNext))
			Convey("Then the server has no response", func() {
				So(resp, ShouldBeNil)
				So(err.Error(), ShouldContainSubstring, "Timeout exceeded")
			})
		})
	})
}

// end of tests //

func NewTestServer() *TestServer {
	callCount := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		contentType := r.Header.Get("Content-type")
		b := getBody(r.Body)
		jsonResponse, err := json.Marshal(Responder{
			Method:      r.Method,
			ContentType: contentType,
			CallCount:   callCount,
			Body:        string(b),
		})
		if err != nil {
			convertErrorToOutput(w, contentType, err)
			return
		}

		// when we see JSON, decode it to see if we need to sleep
		if contentType == jsonContentType {
			reqTest := &RequestTester{}
			if err := json.Unmarshal(b, reqTest); err != nil {
				convertErrorToOutput(w, contentType, err)
				return
			}
			if reqTest.Delay != "" {
				delayDuration, err := time.ParseDuration(reqTest.Delay)
				if err != nil {
					convertErrorToOutput(w, contentType, err)
					return
				}
				if reqTest.DelayOnCall == callCount {
					time.Sleep(delayDuration)
				}
			}
		}

		fmt.Fprint(w, string(jsonResponse))
	}))
	return &TestServer{
		ts:  ts,
		url: ts.URL,
	}
}

func (ts *TestServer) Close() {
	ts.ts.Close()
}

func convertErrorToOutput(w io.Writer, contentType string, err error) {
	if contentType != jsonContentType {
		fmt.Fprint(w, err)
	} else {
		errJson := `{"error":"` + strings.Replace(err.Error(), `"`, "`", -1) + `"}` // replaces " with `
		fmt.Fprint(w, errJson)
	}
}

func delayByOneSecondOn(delayOnCall int) string {
	return `{"delay":"1s","delay_on_call":` + strconv.Itoa(delayOnCall) + `}`
}

func getBody(body io.ReadCloser) []byte {
	b, err := ioutil.ReadAll(body)
	if err != nil {
		panic(err)
	}
	body.Close()
	return b
}

func unmarshallResp(resp *http.Response) (*Responder, error) {
	responder := &Responder{}
	body := getBody(resp.Body)
	err := json.Unmarshal(body, responder)
	if err != nil {
		panic(err.Error() + string(body))
	}
	return responder, err
}
