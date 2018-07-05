package log

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/ONSdigital/go-ns/common"
	"github.com/mgutz/ansi"
)

// Namespace is the service namespace used for logging
var Namespace = "service-namespace"

// HumanReadable represents a flag to determine if log events
// will be in a human readable format
var HumanReadable bool
var hrMutex sync.Mutex

func init() {
	configureHumanReadable()
}

func configureHumanReadable() {
	HumanReadable, _ = strconv.ParseBool(os.Getenv("HUMAN_LOG"))
}

// Data contains structured log data
type Data map[string]interface{}

// GetRequestID returns the request ID from a request (using X-Request-Id)
func GetRequestID(req *http.Request) string {
	return req.Header.Get(common.RequestHeaderKey)
}

// Handler wraps a http.Handler and logs the status code and total response time
func Handler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		rc := &responseCapture{w, 0}

		s := time.Now()
		h.ServeHTTP(rc, req)
		e := time.Now()
		d := e.Sub(s)

		data := Data{
			"start":    s,
			"end":      e,
			"duration": d,
			"status":   rc.statusCode,
			"method":   req.Method,
			"path":     req.URL.Path,
		}
		if len(req.URL.RawQuery) > 0 {
			data["query"] = req.URL.Query()
		}
		Event("request", GetRequestID(req), data)
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

func (r *responseCapture) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if h, ok := r.ResponseWriter.(http.Hijacker); ok {
		return h.Hijack()
	}
	return nil, nil, errors.New("log: response does not implement http.Hijacker")
}

// Event records an event
var Event = event

func event(name string, correlationKey string, data Data) {
	m := map[string]interface{}{
		"created":   time.Now(),
		"event":     name,
		"namespace": Namespace,
	}

	if len(correlationKey) > 0 {
		m["correlation_key"] = correlationKey
	}

	if data != nil {
		m["data"] = data
	}

	if HumanReadable {
		printHumanReadable(name, correlationKey, data, m)
		return
	}

	b, err := json.Marshal(&m)
	if err != nil {
		// This should never happen
		// We'll log the error (which for our purposes, can't fail), which
		// gives us an indication we have something to investigate
		b, _ = json.Marshal(map[string]interface{}{
			"created":         time.Now(),
			"event":           "log_error",
			"namespace":       Namespace,
			"correlation_key": correlationKey,
			"data":            map[string]interface{}{"error": err.Error()},
		})
	}

	fmt.Fprintf(os.Stdout, "%s\n", b)
}

func printHumanReadable(name, correlationKey string, data Data, m map[string]interface{}) {
	hrMutex.Lock()
	defer hrMutex.Unlock()

	ctx := ""
	if len(correlationKey) > 0 {
		ctx = "[" + correlationKey + "] "
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
	case "info":
		col = ansi.LightCyan
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

// ErrorC is a structured error message with correlationKey
func ErrorC(correlationKey string, err error, data Data) {
	if data == nil {
		data = Data{}
	}
	if err != nil {
		data["message"] = err.Error()
		data["error"] = err
	}
	Event("error", correlationKey, data)
}

// ErrorCtx is a structured error message and retrieves the correlationKey from go context
func ErrorCtx(ctx context.Context, err error, data Data) {
	correlationKey := common.GetRequestId(ctx)
	ErrorC(correlationKey, err, data)
}

// ErrorR is a structured error message for a request
func ErrorR(req *http.Request, err error, data Data) {
	ErrorC(GetRequestID(req), err, data)
}

// Error is a structured error message
func Error(err error, data Data) {
	ErrorC("", err, data)
}

// DebugC is a structured debug message with correlationKey
func DebugC(correlationKey string, message string, data Data) {
	if data == nil {
		data = Data{}
	}
	if len(message) > 0 {
		data["message"] = message
	}
	Event("debug", correlationKey, data)
}

// DebugCtx is a structured debug message and retrieves the correlationKey from go context
func DebugCtx(ctx context.Context, message string, data Data) {
	correlationKey := common.GetRequestId(ctx)
	DebugC(correlationKey, message, data)
}

// DebugR is a structured debug message for a request
func DebugR(req *http.Request, message string, data Data) {
	DebugC(GetRequestID(req), message, data)
}

// Debug is a structured trace message
func Debug(message string, data Data) {
	DebugC("", message, data)
}

// TraceC is a structured trace message with correlationKey
func TraceC(correlationKey string, message string, data Data) {
	if data == nil {
		data = Data{}
	}
	if len(message) > 0 {
		data["message"] = message
	}
	Event("trace", correlationKey, data)
}

// TraceCtx is a structured trace message and retrieves the correlationKey from go context
func TraceCtx(ctx context.Context, message string, data Data) {
	correlationKey := common.GetRequestId(ctx)
	TraceC(correlationKey, message, data)
}

// TraceR is a structured trace message for a request
func TraceR(req *http.Request, message string, data Data) {
	TraceC(GetRequestID(req), message, data)
}

// Trace is a structured trace message
func Trace(message string, data Data) {
	TraceC("", message, data)
}

// InfoC is a structured info message with correlationKey
func InfoC(correlationKey string, message string, data Data) {
	if data == nil {
		data = Data{}
	}
	if len(message) > 0 {
		data["message"] = message
	}
	Event("info", correlationKey, data)
}

// InfoCtx is a structured info message and retrieves the correlationKey from go context
func InfoCtx(ctx context.Context, message string, data Data) {
	correlationKey := common.GetRequestId(ctx)
	InfoC(correlationKey, message, data)
}

// InfoR is a structured info message for a request
func InfoR(req *http.Request, message string, data Data) {
	InfoC(GetRequestID(req), message, data)
}

// Info is a structured info message
func Info(message string, data Data) {
	InfoC("", message, data)
}
