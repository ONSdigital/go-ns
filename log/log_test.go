package log

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

var dummyHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(200)
})

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
		So(HumanReadable, ShouldBeFalse)

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

func TestContext(t *testing.T) {
	Convey("Context should retrieve the X-Request-Id from a request", t, func() {
		req, err := http.NewRequest("GET", "/", nil)
		So(err, ShouldBeNil)

		ctx := Context(req)
		So(ctx, ShouldBeEmpty)

		req.Header.Set("X-Request-Id", "test")
		ctx = Context(req)
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

	var eventName, eventContext string
	var eventData Data
	Event = func(name string, context string, data Data) {
		eventName = name
		eventContext = context
		eventData = data
	}

	Convey("Error", t, func() {
		Error(errors.New("test error"), nil)
		So(eventName, ShouldEqual, "error")
		So(eventContext, ShouldEqual, "")
		So(eventData, ShouldContainKey, "message")
		So(eventData["message"], ShouldEqual, "test error")
		So(eventData, ShouldContainKey, "error")
		So(eventData["error"], ShouldHaveSameTypeAs, errors.New(""))
		So(eventData["error"].(error).Error(), ShouldEqual, "test error")
	})

	Convey("ErrorC", t, func() {
		ErrorC("context", errors.New("test error"), nil)
		So(eventName, ShouldEqual, "error")
		So(eventContext, ShouldEqual, "context")
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
		So(eventContext, ShouldEqual, "test-request-id")
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

	var eventName, eventContext string
	var eventData Data
	Event = func(name string, context string, data Data) {
		eventName = name
		eventContext = context
		eventData = data
	}

	Convey("Debug", t, func() {
		Debug("test message", nil)
		So(eventName, ShouldEqual, "debug")
		So(eventContext, ShouldEqual, "")
		So(eventData, ShouldContainKey, "message")
		So(eventData["message"], ShouldEqual, "test message")
	})

	Convey("DebugC", t, func() {
		DebugC("context", "test message", nil)
		So(eventName, ShouldEqual, "debug")
		So(eventContext, ShouldEqual, "context")
		So(eventData, ShouldContainKey, "message")
		So(eventData["message"], ShouldEqual, "test message")
	})

	Convey("DebugR", t, func() {
		req, err := http.NewRequest("GET", "/", nil)
		So(err, ShouldBeNil)

		req.Header.Set("X-Request-Id", "test-request-id")

		DebugR(req, "test message", nil)
		So(eventName, ShouldEqual, "debug")
		So(eventContext, ShouldEqual, "test-request-id")
		So(eventData, ShouldContainKey, "message")
		So(eventData["message"], ShouldEqual, "test message")
	})
}

func TestTrace(t *testing.T) {
	oldEvent := Event
	defer func() {
		Event = oldEvent
	}()

	var eventName, eventContext string
	var eventData Data
	Event = func(name string, context string, data Data) {
		eventName = name
		eventContext = context
		eventData = data
	}

	Convey("Trace", t, func() {
		Trace("test message", nil)
		So(eventName, ShouldEqual, "trace")
		So(eventContext, ShouldEqual, "")
		So(eventData, ShouldContainKey, "message")
		So(eventData["message"], ShouldEqual, "test message")
	})

	Convey("TraceC", t, func() {
		TraceC("context", "test message", nil)
		So(eventName, ShouldEqual, "trace")
		So(eventContext, ShouldEqual, "context")
		So(eventData, ShouldContainKey, "message")
		So(eventData["message"], ShouldEqual, "test message")
	})

	Convey("TraceR", t, func() {
		req, err := http.NewRequest("GET", "/", nil)
		So(err, ShouldBeNil)

		req.Header.Set("X-Request-Id", "test-request-id")

		TraceR(req, "test message", nil)
		So(eventName, ShouldEqual, "trace")
		So(eventContext, ShouldEqual, "test-request-id")
		So(eventData, ShouldContainKey, "message")
		So(eventData["message"], ShouldEqual, "test message")
	})
}

func TestEvent(t *testing.T) {
	Convey("event should output JSON", t, func() {
		Namespace = "namespace"

		stdout := captureOutput(func() {
			event("test", "context", Data{"foo": "bar"})
		})
		var m map[string]interface{}
		err := json.Unmarshal([]byte(stdout), &m)
		So(err, ShouldBeNil)
		log.Println(err)

		So(m, ShouldContainKey, "created")
		So(m, ShouldContainKey, "event")
		So(m["event"], ShouldEqual, "test")
		So(m, ShouldContainKey, "namespace")
		So(m["namespace"], ShouldEqual, "namespace")
		So(m, ShouldContainKey, "context")
		So(m["context"], ShouldEqual, "context")
		So(m, ShouldContainKey, "data")
		So(m["data"], ShouldHaveSameTypeAs, map[string]interface{}{})
		So(m["data"].(map[string]interface{})["foo"], ShouldEqual, "bar")
	})
}
