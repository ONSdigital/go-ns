package audit

import (
	"context"
	"errors"
	"testing"

	"github.com/ONSdigital/go-ns/common"
	"github.com/ONSdigital/go-ns/handlers/requestID"
	"github.com/ONSdigital/go-ns/log"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	eventRes eventResults

	ctx = context.Background()
)

type eventResults struct {
	name string
	cKey string
	data log.Data
}

func init() {
	log.Event = func(name string, correlationKey string, data log.Data) {
		eventRes = eventResults{
			name: name,
			cKey: correlationKey,
			data: data,
		}
	}
}

func TestErrorLog(t *testing.T) {

	errTest := errors.New("test error")

	Convey("with empty context", t, func() {
		LogError(ctx, errTest, nil)

		So(eventRes, ShouldResemble, eventResults{
			name: "error",
			cKey: "",
			data: log.Data{"message": "test error", "error": errTest},
		})
	})

	Convey("context with request id", t, func() {
		contextWithRequestID := context.WithValue(ctx, requestID.ContextKey, "request id")
		LogError(contextWithRequestID, errTest, nil)

		So(eventRes, ShouldResemble, eventResults{
			name: "error",
			cKey: "request id",
			data: log.Data{"message": "test error", "error": errTest},
		})
	})

	Convey("context with user", t, func() {
		contextWithUser := context.WithValue(ctx, common.UserIdentityKey, "user email")
		LogError(contextWithUser, errTest, nil)

		So(eventRes, ShouldResemble, eventResults{
			name: "error",
			cKey: "",
			data: log.Data{"message": "test error", "error": errTest, reqUser: "user email"},
		})
	})

	Convey("context with caller id", t, func() {
		contextWithCaller := context.WithValue(ctx, common.CallerIdentityKey, "api service")
		LogError(contextWithCaller, errTest, nil)

		So(eventRes, ShouldResemble, eventResults{
			name: "error",
			cKey: "",
			data: log.Data{"message": "test error", "error": errTest, reqCaller: "api service"},
		})
	})

	Convey("context with request, user and caller ids", t, func() {
		ctxWithIDs := getContextWithCallerUserAndRequestIDcontext()
		LogError(ctxWithIDs, errTest, nil)

		So(eventRes, ShouldResemble, eventResults{
			name: "error",
			cKey: "request id",
			data: log.Data{
				"message": "test error",
				"error":   errTest,
				reqCaller: "api service",
				reqUser:   "user email",
			},
		})
	})
}

func TestInfoLog(t *testing.T) {

	Convey("with empty context", t, func() {
		LogInfo(ctx, "info message", nil)

		So(eventRes, ShouldResemble, eventResults{
			name: "info",
			cKey: "",
			data: log.Data{"message": "info message"},
		})
	})

	Convey("context with request id", t, func() {
		contextWithRequestID := context.WithValue(ctx, requestID.ContextKey, "request id")
		LogInfo(contextWithRequestID, "info message", nil)

		So(eventRes, ShouldResemble, eventResults{
			name: "info",
			cKey: "request id",
			data: log.Data{"message": "info message"},
		})
	})

	Convey("context with user", t, func() {
		contextWithUser := context.WithValue(ctx, common.UserIdentityKey, "user email")
		LogInfo(contextWithUser, "info message", nil)

		So(eventRes, ShouldResemble, eventResults{
			name: "info",
			cKey: "",
			data: log.Data{"message": "info message", reqUser: "user email"},
		})
	})

	Convey("context with caller id", t, func() {
		contextWithCaller := context.WithValue(ctx, common.CallerIdentityKey, "api service")
		LogInfo(contextWithCaller, "info message", nil)

		So(eventRes, ShouldResemble, eventResults{
			name: "info",
			cKey: "",
			data: log.Data{"message": "info message", reqCaller: "api service"},
		})
	})

	Convey("context with request, user and caller ids", t, func() {
		ctxWithIDs := getContextWithCallerUserAndRequestIDcontext()
		LogInfo(ctxWithIDs, "info message", nil)

		So(eventRes, ShouldResemble, eventResults{
			name: "info",
			cKey: "request id",
			data: log.Data{
				"message": "info message",
				reqCaller: "api service",
				reqUser:   "user email",
			},
		})
	})
}

func TestAddLogData(t *testing.T) {
	Convey("with empty context", t, func() {
		data := addLogData(ctx, nil)

		So(data, ShouldResemble, log.Data{})
	})

	Convey("context with user", t, func() {
		contextWithUser := context.WithValue(ctx, common.UserIdentityKey, "user email")
		data := addLogData(contextWithUser, nil)

		So(data, ShouldResemble, log.Data{reqUser: "user email"})
	})

	Convey("context with caller id", t, func() {
		contextWithCaller := context.WithValue(ctx, common.CallerIdentityKey, "api service")
		data := addLogData(contextWithCaller, nil)

		So(data, ShouldResemble, log.Data{reqCaller: "api service"})
	})

	Convey("context with request, user and caller ids", t, func() {
		ctx := getContextWithCallerUserAndRequestIDcontext()
		data := addLogData(ctx, nil)

		So(data, ShouldResemble, log.Data{
			reqCaller: "api service",
			reqUser:   "user email",
		})
	})
}

func getContextWithCallerUserAndRequestIDcontext() context.Context {
	ctx := context.WithValue(context.Background(), requestID.ContextKey, "request id")
	ctx = context.WithValue(ctx, common.UserIdentityKey, "user email")
	ctx = context.WithValue(ctx, common.CallerIdentityKey, "api service")
	return ctx
}
