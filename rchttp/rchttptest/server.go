package rchttptest

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"
)

const (
	JsonContentType   = "application/json"
	FormEncodedType   = "application/x-www-form-urlencoded"
	ServiceToken      = "APIAmWhoAPIAm"
	UserId            = "User@Test"
	ContentTypeHeader = "Content-Type"
)

type TestServer struct {
	Server *httptest.Server
	URL    string
}

type Responder struct {
	Body      string              `json:"body"`
	CallCount int                 `json:"call_count"`
	Method    string              `json:"method"`
	Error     string              `json:"error"`
	Headers   map[string][]string `json:"headers"`
}

type RequestTester struct {
	Delay         string `json:"delay"`
	DelayDuration time.Duration
	DelayOnCall   int `json:"delay_on_call"`
}

func NewTestServer() *TestServer {
	callCount := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		contentType := r.Header.Get(ContentTypeHeader)
		b := GetBody(r.Body)
		headers := make(map[string][]string)
		for h, v := range r.Header {
			headers[h] = v
		}
		jsonResponse, err := json.Marshal(Responder{
			Method:    r.Method,
			CallCount: callCount,
			Body:      string(b),
			Headers:   headers,
		})
		if err != nil {
			convertErrorToOutput(w, contentType, err)
			return
		}

		// when we see JSON, decode it to see if we need to sleep
		if contentType == JsonContentType {
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
		Server: ts,
		URL:    ts.URL,
	}
}

func (ts *TestServer) Close() {
	ts.Server.Close()
}

func convertErrorToOutput(w io.Writer, contentType string, err error) {
	if contentType != JsonContentType {
		fmt.Fprint(w, err)
	} else {
		errJson := `{"error":"` + strings.Replace(err.Error(), `"`, "`", -1) + `"}` // replaces " with `
		fmt.Fprint(w, errJson)
	}
}

func GetBody(body io.ReadCloser) []byte {
	b, err := ioutil.ReadAll(body)
	if err != nil {
		panic(err)
	}
	body.Close()
	return b
}
