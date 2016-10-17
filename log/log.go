package log

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/mgutz/ansi"
)

// Namespace is the service namespace used for logging
var Namespace = "service-namespace"

// HumanReadable, if true, outputs log events in a human readable format
var HumanReadable bool

func init() {
	configureHumanReadable()
}

func configureHumanReadable() {
	HumanReadable, _ = strconv.ParseBool(os.Getenv("HUMAN_LOG"))
}

// Data contains structured log data
type Data map[string]interface{}

// Context returns a context ID from a request (using X-Request-Id)
func Context(req *http.Request) string {
	return req.Header.Get("X-Request-Id")
}

// Handler wraps a http.Handler and logs the status code and total response time
func Handler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		rc := &responseCapture{w, 0}

		s := time.Now()
		h.ServeHTTP(rc, req)
		e := time.Now()
		d := e.Sub(s)

		Event("request", Context(req), Data{
			"start":    s,
			"end":      e,
			"duration": d,
			"status":   rc.statusCode,
			"method":   req.Method,
			"path":     req.URL.Path,
		})
	})
}

type responseCapture struct {
	http.ResponseWriter
	statusCode int
}

func (r *responseCapture) WriteHeader(status int) {
	r.statusCode = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *responseCapture) Flush() {
	if f, ok := r.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// Event records an event
var Event = event

func event(name string, context string, data Data) {
	m := map[string]interface{}{
		"created":   time.Now(),
		"event":     name,
		"namespace": Namespace,
	}

	if len(context) > 0 {
		m["context"] = context
	}

	if data != nil {
		m["data"] = data
	}

	if HumanReadable {
		printHumanReadable(name, context, data, m)
		return
	}

	b, err := json.Marshal(&m)
	if err != nil {
		// This should never happen
		// FIXME: not sure what we should do here
	}

	fmt.Fprintf(os.Stdout, "%s\n", b)
}

func printHumanReadable(name, context string, data Data, m map[string]interface{}) {
	ctx := ""
	if len(context) > 0 {
		ctx = "[" + context + "] "
	}
	msg := ""
	if message, ok := data["message"]; ok {
		msg = ": " + fmt.Sprintf("%s", message)
		delete(data, "message")
	}
	if name == "error" && len(msg) == 0 {
		if err, ok := data["error"]; ok {
			msg = ": " + fmt.Sprintf("%s", err)
			delete(data, "error")
		}
	}
	col := ansi.DefaultFG
	switch name {
	case "error":
		col = ansi.LightRed
	case "trace":
		col = ansi.Blue
	case "debug":
		col = ansi.Green
	case "request":
		col = ansi.Cyan
	}

	fmt.Fprintf(os.Stdout, "%s%s %s%s%s%s\n", col, m["created"], ctx, name, msg, ansi.DefaultFG)
	if data != nil {
		for k, v := range data {
			fmt.Fprintf(os.Stdout, "  -> %s: %+v\n", k, v)
		}
	}
}

// ErrorC is a structured error message with context
func ErrorC(context string, err error, data Data) {
	if data == nil {
		data = Data{}
	}
	if _, ok := data["error"]; !ok {
		data["message"] = err.Error()
		data["error"] = err
	}
	Event("error", context, data)
}

// ErrorR is a structured error message for a request
func ErrorR(req *http.Request, err error, data Data) {
	ErrorC(Context(req), err, data)
}

// Error is a structured error message
func Error(err error, data Data) {
	ErrorC("", err, data)
}

// DebugC is a structured debug message with context
func DebugC(context string, message string, data Data) {
	if data == nil {
		data = Data{}
	}
	if _, ok := data["message"]; !ok {
		data["message"] = message
	}
	Event("debug", context, data)
}

// DebugR is a structured debug message for a request
func DebugR(req *http.Request, message string, data Data) {
	DebugC(Context(req), message, data)
}

// Debug is a structured trace message
func Debug(message string, data Data) {
	DebugC("", message, data)
}

// TraceC is a structured trace message with context
func TraceC(context string, message string, data Data) {
	if data == nil {
		data = Data{}
	}
	if _, ok := data["message"]; !ok {
		data["message"] = message
	}
	Event("trace", context, data)
}

// TraceR is a structured trace message for a request
func TraceR(req *http.Request, message string, data Data) {
	TraceC(Context(req), message, data)
}

// Trace is a structured trace message
func Trace(message string, data Data) {
	TraceC("", message, data)
}
