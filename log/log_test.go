package log

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/ONSdigital/go-ns/handlers/requestID"
	"github.com/mgutz/ansi"
	. "github.com/smartystreets/goconvey/convey"
)

var dummyHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(200)
})

var contextWithRequestHeader = context.WithValue(context.Background(), requestID.ContextKey, "request id")

func captureOutput(f func()) string {
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() {
		os.Stdout = stdout
	}()

	outC := make(chan string)
	// copy the output in a separate goroutine so printing can't block indefinitely
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	f()

	w.Close()
	out := <-outC
	return out
}

func TestHumanLog(t *testing.T) {
	Convey("HUMAN_LOG environment variable should configure human log output", t, func() {
		// Cannot guarentee the variable has not been set already
		if os.Getenv("HUMAN_LOG") == "" {
			So(HumanReadable, ShouldBeFalse)
		}

		os.Setenv("HUMAN_LOG", "true")
		configureHumanReadable()
		So(HumanReadable, ShouldBeTrue)

		os.Setenv("HUMAN_LOG", "false")
		configureHumanReadable()
		So(HumanReadable, ShouldBeFalse)

		os.Setenv("HUMAN_LOG", "1")
		configureHumanReadable()
		So(HumanReadable, ShouldBeTrue)

		os.Setenv("HUMAN_LOG", "")
		configureHumanReadable()
		So(HumanReadable, ShouldBeFalse)
	})
}

func TestGetRequestID(t *testing.T) {
	Convey("GetRequestID should retrieve the X-Request-Id from a request", t, func() {
		req, err := http.NewRequest("GET", "/", nil)
		So(err, ShouldBeNil)

		ctx := GetRequestID(req)
		So(ctx, ShouldBeEmpty)

		req.Header.Set("X-Request-Id", "test")
		ctx = GetRequestID(req)
		So(ctx, ShouldEqual, "test")
	})
}

func TestHandler(t *testing.T) {
	Convey("Handler should wrap another handler", t, func() {
		wrapped := Handler(dummyHandler)
		So(wrapped, ShouldHaveSameTypeAs, dummyHandler)
	})

	Convey("Handler should capture stuff", t, func() {
		oldEvent := Event
		defer func() {
			Event = oldEvent
		}()

		wrapped := Handler(dummyHandler)

		var eventName, eventContext string
		var eventData Data
		Event = func(name string, context string, data Data) {
			eventName = name
			eventContext = context
			eventData = data
		}

		req, err := http.NewRequest("GET", "/", nil)
		So(err, ShouldBeNil)
		req.Header.Set("X-Request-Id", "test")

		w := httptest.NewRecorder()

		wrapped.ServeHTTP(w, req)
		So(eventName, ShouldEqual, "request")
		So(eventContext, ShouldEqual, "test")
		So(eventData, ShouldContainKey, "start")
		So(eventData["start"], ShouldHaveSameTypeAs, time.Now())
		So(eventData, ShouldContainKey, "end")
		So(eventData["end"], ShouldHaveSameTypeAs, time.Now())
		So(eventData, ShouldContainKey, "duration")
		So(eventData["duration"], ShouldHaveSameTypeAs, time.Now().Sub(time.Now()))
		So(eventData, ShouldContainKey, "status")
		So(eventData["status"], ShouldEqual, 200)
		So(eventData, ShouldContainKey, "method")
		So(eventData["method"], ShouldEqual, "GET")
		So(eventData, ShouldContainKey, "path")
		So(eventData["path"], ShouldEqual, "/")

		// query should not be included if no query parameter exists
		So(eventData, ShouldNotContainKey, "query")
	})

	Convey("Handler should capture query string values", t, func() {
		oldEvent := Event
		defer func() {
			Event = oldEvent
		}()

		wrapped := Handler(dummyHandler)

		var eventData Data
		Event = func(name string, context string, data Data) {
			eventData = data
		}

		req, err := http.NewRequest("GET", "/", nil)
		So(err, ShouldBeNil)
		req.Header.Set("X-Request-Id", "test")

		q := req.URL.Query()
		q.Add("foo", "bar")
		q.Add("foo", "baz")
		q.Add("uri", "/a/b/c")
		req.URL.RawQuery = q.Encode()

		w := httptest.NewRecorder()

		wrapped.ServeHTTP(w, req)

		So(eventData, ShouldContainKey, "query")
		So(eventData["query"], ShouldHaveSameTypeAs, q)
		So(eventData["query"], ShouldResemble, url.Values{"foo": []string{"bar", "baz"}, "uri": []string{"/a/b/c"}})
	})
}

func TestResponseCapture(t *testing.T) {
	Convey("responseCapture should capture a response status code", t, func() {
		w := httptest.NewRecorder()
		c := responseCapture{w, 0}
		So(c.statusCode, ShouldEqual, 0)

		c.WriteHeader(200)
		So(c.statusCode, ShouldEqual, 200)
	})

	Convey("responseCapture should pass through a Flush call", t, func() {
		w := httptest.NewRecorder()
		c := responseCapture{w, 0}
		So(w.Flushed, ShouldBeFalse)

		c.Flush()
		So(w.Flushed, ShouldBeTrue)
	})
}

func TestError(t *testing.T) {
	oldEvent := Event
	defer func() {
		Event = oldEvent
	}()

	var eventName, eventCorrelationKey string
	var eventData Data
	Event = func(name string, correlationKey string, data Data) {
		eventName = name
		eventCorrelationKey = correlationKey
		eventData = data
	}

	Convey("Error", t, func() {
		Error(errors.New("test error"), nil)
		So(eventName, ShouldEqual, "error")
		So(eventCorrelationKey, ShouldEqual, "")
		So(eventData, ShouldContainKey, "message")
		So(eventData["message"], ShouldEqual, "test error")
		So(eventData, ShouldContainKey, "error")
		So(eventData["error"], ShouldHaveSameTypeAs, errors.New(""))
		So(eventData["error"].(error).Error(), ShouldEqual, "test error")
	})

	Convey("ErrorCtx", t, func() {
		ErrorCtx(context.Background(), errors.New("test error"), nil)
		So(eventName, ShouldEqual, "error")
		So(eventCorrelationKey, ShouldEqual, "")
		So(eventData, ShouldContainKey, "message")
		So(eventData["message"], ShouldEqual, "test error")
		So(eventData, ShouldContainKey, "error")
		So(eventData["error"], ShouldHaveSameTypeAs, errors.New(""))
		So(eventData["error"].(error).Error(), ShouldEqual, "test error")

		ErrorCtx(contextWithRequestHeader, errors.New("test error"), nil)
		So(eventName, ShouldEqual, "error")
		So(eventCorrelationKey, ShouldEqual, "request id")
		So(eventData, ShouldContainKey, "message")
		So(eventData["message"], ShouldEqual, "test error")
		So(eventData, ShouldContainKey, "error")
		So(eventData["error"], ShouldHaveSameTypeAs, errors.New(""))
		So(eventData["error"].(error).Error(), ShouldEqual, "test error")
	})

	Convey("ErrorC", t, func() {
		ErrorC("correlation id", errors.New("test error"), nil)
		So(eventName, ShouldEqual, "error")
		So(eventCorrelationKey, ShouldEqual, "correlation id")
		So(eventData, ShouldContainKey, "message")
		So(eventData["message"], ShouldEqual, "test error")
		So(eventData, ShouldContainKey, "error")
		So(eventData["error"], ShouldHaveSameTypeAs, errors.New(""))
		So(eventData["error"].(error).Error(), ShouldEqual, "test error")
	})

	Convey("ErrorR", t, func() {
		req, err := http.NewRequest("GET", "/", nil)
		So(err, ShouldBeNil)

		req.Header.Set("X-Request-Id", "test-request-id")

		ErrorR(req, errors.New("test error"), nil)
		So(eventName, ShouldEqual, "error")
		So(eventCorrelationKey, ShouldEqual, "test-request-id")
		So(eventData, ShouldContainKey, "message")
		So(eventData["message"], ShouldEqual, "test error")
		So(eventData, ShouldContainKey, "error")
		So(eventData["error"], ShouldHaveSameTypeAs, errors.New(""))
		So(eventData["error"].(error).Error(), ShouldEqual, "test error")
	})
}

func TestDebug(t *testing.T) {
	oldEvent := Event
	defer func() {
		Event = oldEvent
	}()

	var eventName, eventCorrelationKey string
	var eventData Data
	Event = func(name string, correlationKey string, data Data) {
		eventName = name
		eventCorrelationKey = correlationKey
		eventData = data
	}

	Convey("Debug", t, func() {
		Debug("test message", nil)
		So(eventName, ShouldEqual, "debug")
		So(eventCorrelationKey, ShouldEqual, "")
		So(eventData, ShouldContainKey, "message")
		So(eventData["message"], ShouldEqual, "test message")
	})

	Convey("DebugCtx", t, func() {
		DebugCtx(context.Background(), "test message", nil)
		So(eventName, ShouldEqual, "debug")
		So(eventCorrelationKey, ShouldEqual, "")
		So(eventData, ShouldContainKey, "message")
		So(eventData["message"], ShouldEqual, "test message")

		DebugCtx(contextWithRequestHeader, "test message", nil)
		So(eventName, ShouldEqual, "debug")
		So(eventCorrelationKey, ShouldEqual, "request id")
		So(eventData, ShouldContainKey, "message")
		So(eventData["message"], ShouldEqual, "test message")
	})

	Convey("DebugC", t, func() {
		DebugC("correlation id", "test message", nil)
		So(eventName, ShouldEqual, "debug")
		So(eventCorrelationKey, ShouldEqual, "correlation id")
		So(eventData, ShouldContainKey, "message")
		So(eventData["message"], ShouldEqual, "test message")
	})

	Convey("DebugR", t, func() {
		req, err := http.NewRequest("GET", "/", nil)
		So(err, ShouldBeNil)

		req.Header.Set("X-Request-Id", "test-request-id")

		DebugR(req, "test message", nil)
		So(eventName, ShouldEqual, "debug")
		So(eventCorrelationKey, ShouldEqual, "test-request-id")
		So(eventData, ShouldContainKey, "message")
		So(eventData["message"], ShouldEqual, "test message")
	})
}

func TestTrace(t *testing.T) {
	oldEvent := Event
	defer func() {
		Event = oldEvent
	}()

	var eventName, eventCorrelationKey string
	var eventData Data
	Event = func(name string, correlationKey string, data Data) {
		eventName = name
		eventCorrelationKey = correlationKey
		eventData = data
	}

	Convey("Trace", t, func() {
		Trace("test message", nil)
		So(eventName, ShouldEqual, "trace")
		So(eventCorrelationKey, ShouldEqual, "")
		So(eventData, ShouldContainKey, "message")
		So(eventData["message"], ShouldEqual, "test message")
	})

	Convey("TraceCtx", t, func() {
		TraceCtx(context.Background(), "test message", nil)
		So(eventName, ShouldEqual, "trace")
		So(eventCorrelationKey, ShouldEqual, "")
		So(eventData, ShouldContainKey, "message")
		So(eventData["message"], ShouldEqual, "test message")

		TraceCtx(contextWithRequestHeader, "test message", nil)
		So(eventName, ShouldEqual, "trace")
		So(eventCorrelationKey, ShouldEqual, "request id")
		So(eventData, ShouldContainKey, "message")
		So(eventData["message"], ShouldEqual, "test message")
	})

	Convey("TraceC", t, func() {
		TraceC("correlation id", "test message", nil)
		So(eventName, ShouldEqual, "trace")
		So(eventCorrelationKey, ShouldEqual, "correlation id")
		So(eventData, ShouldContainKey, "message")
		So(eventData["message"], ShouldEqual, "test message")
	})

	Convey("TraceR", t, func() {
		req, err := http.NewRequest("GET", "/", nil)
		So(err, ShouldBeNil)

		req.Header.Set("X-Request-Id", "test-request-id")

		TraceR(req, "test message", nil)
		So(eventName, ShouldEqual, "trace")
		So(eventCorrelationKey, ShouldEqual, "test-request-id")
		So(eventData, ShouldContainKey, "message")
		So(eventData["message"], ShouldEqual, "test message")
	})
}

func TestInfo(t *testing.T) {
	oldEvent := Event
	defer func() {
		Event = oldEvent
	}()

	var eventName, eventCorrelationKey string
	var eventData Data
	Event = func(name string, correlationKey string, data Data) {
		eventName = name
		eventCorrelationKey = correlationKey
		eventData = data
	}

	Convey("Info", t, func() {
		Info("test message", nil)
		So(eventName, ShouldEqual, "info")
		So(eventCorrelationKey, ShouldEqual, "")
		So(eventData, ShouldContainKey, "message")
		So(eventData["message"], ShouldEqual, "test message")
	})

	Convey("InfoCtx", t, func() {
		InfoCtx(context.Background(), "test message", nil)
		So(eventName, ShouldEqual, "info")
		So(eventCorrelationKey, ShouldEqual, "")
		So(eventData, ShouldContainKey, "message")
		So(eventData["message"], ShouldEqual, "test message")

		InfoCtx(contextWithRequestHeader, "test message", nil)
		So(eventName, ShouldEqual, "info")
		So(eventCorrelationKey, ShouldEqual, "request id")
		So(eventData, ShouldContainKey, "message")
		So(eventData["message"], ShouldEqual, "test message")
	})

	Convey("InfoC", t, func() {
		InfoC("correlation id", "test message", nil)
		So(eventName, ShouldEqual, "info")
		So(eventCorrelationKey, ShouldEqual, "correlation id")
		So(eventData, ShouldContainKey, "message")
		So(eventData["message"], ShouldEqual, "test message")
	})

	Convey("InfoR", t, func() {
		req, err := http.NewRequest("GET", "/", nil)
		So(err, ShouldBeNil)

		req.Header.Set("X-Request-Id", "test-request-id")

		InfoR(req, "test message", nil)
		So(eventName, ShouldEqual, "info")
		So(eventCorrelationKey, ShouldEqual, "test-request-id")
		So(eventData, ShouldContainKey, "message")
		So(eventData["message"], ShouldEqual, "test message")
	})
}

func TestEvent(t *testing.T) {
	Convey("event should output JSON", t, func() {
		Namespace = "namespace"

		stdout := captureOutput(func() {
			event("test", "request id", Data{"foo": "bar"})
		})
		var m map[string]interface{}
		err := json.Unmarshal([]byte(stdout), &m)
		So(err, ShouldBeNil)

		So(m, ShouldContainKey, "created")
		So(m, ShouldContainKey, "event")
		So(m["event"], ShouldEqual, "test")
		So(m, ShouldContainKey, "namespace")
		So(m["namespace"], ShouldEqual, "namespace")
		So(m, ShouldContainKey, "correlationKey")
		So(m["correlationKey"], ShouldEqual, "request id")
		So(m, ShouldContainKey, "data")
		So(m["data"], ShouldHaveSameTypeAs, map[string]interface{}{})
		So(m["data"].(map[string]interface{})["foo"], ShouldEqual, "bar")
	})

	Convey("event with invalid data value should fail", t, func() {
		Namespace = "namespace"
		HumanReadable = false

		stdout := captureOutput(func() {
			event("test", "request id", Data{"foo": func() {}})
		})
		var m map[string]interface{}
		err := json.Unmarshal([]byte(stdout), &m)
		So(err, ShouldBeNil)

		So(m, ShouldContainKey, "created")
		So(m, ShouldContainKey, "event")
		So(m["event"], ShouldEqual, "log_error")
		So(m, ShouldContainKey, "namespace")
		So(m["namespace"], ShouldEqual, "namespace")
		So(m, ShouldContainKey, "correlationKey")
		So(m["correlationKey"], ShouldEqual, "request id")
		So(m, ShouldContainKey, "data")
		So(m["data"], ShouldHaveSameTypeAs, map[string]interface{}{})
		So(m["data"].(map[string]interface{})["error"], ShouldEqual, "json: unsupported type: func()")
	})
}

type humanReadableTest struct {
	name, correlationKey string
	data                 Data
	m                    map[string]interface{}
	result               string
}

func TestPrintHumanReadable(t *testing.T) {
	now := time.Now()
	var tests = []humanReadableTest{
		{
			"name", "correlation_id", Data{"foo": "bar"},
			map[string]interface{}{"created": now},
			fmt.Sprintf("%s%s [%s] %s%s\n  -> %s: %+v\n", ansi.DefaultFG, now, "correlation_id", "name", ansi.DefaultFG, "foo", "bar"),
		},
		{
			"name", "correlation_id", Data{"message": "test message"},
			map[string]interface{}{"created": now},
			fmt.Sprintf("%s%s [%s] %s: %s%s\n", ansi.DefaultFG, now, "correlation_id", "name", "test message", ansi.DefaultFG),
		},
		{
			"error", "correlation_id", Data{"error": errors.New("test error")},
			map[string]interface{}{"created": now},
			fmt.Sprintf("%s%s [%s] %s: %s%s\n", ansi.LightRed, now, "correlation_id", "error", "test error", ansi.DefaultFG),
		},
		{
			"trace", "correlation_id", Data{"message": "test message"},
			map[string]interface{}{"created": now},
			fmt.Sprintf("%s%s [%s] %s: %s%s\n", ansi.Blue, now, "correlation_id", "trace", "test message", ansi.DefaultFG),
		},
		{
			"debug", "correlation_id", Data{"message": "test message"},
			map[string]interface{}{"created": now},
			fmt.Sprintf("%s%s [%s] %s: %s%s\n", ansi.Green, now, "correlation_id", "debug", "test message", ansi.DefaultFG),
		},
		{
			"info", "correlation_id", Data{"message": "test message"},
			map[string]interface{}{"created": now},
			fmt.Sprintf("%s%s [%s] %s: %s%s\n", ansi.LightCyan, now, "correlation_id", "info", "test message", ansi.DefaultFG),
		},
		{
			"request", "correlation_id", Data{"message": "test message"},
			map[string]interface{}{"created": now},
			fmt.Sprintf("%s%s [%s] %s: %s%s\n", ansi.Cyan, now, "correlation_id", "request", "test message", ansi.DefaultFG),
		},
	}

	Convey("printHumanReadable should output human readable log messages", t, func() {
		Namespace = "namespace"
		HumanReadable = true

		for _, test := range tests {
			stdout := captureOutput(func() {
				printHumanReadable(test.name, test.correlationKey, test.data, test.m)
			})
			So(stdout, ShouldEqual, test.result)
		}
	})

	Convey("event should call printHumanReadable if HumanReadable is set", t, func() {
		Namespace = "namespace"
		HumanReadable = true
		stdout := captureOutput(func() {
			event("debug", "context", Data{"message": "test message"})
		})
		endWith := fmt.Sprintf("[%s] %s: %s%s\n", "context", "debug", "test message", ansi.DefaultFG)
		So(stdout, ShouldEndWith, endWith)
	})
}
