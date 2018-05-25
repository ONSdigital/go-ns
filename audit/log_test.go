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

func TestErrorLog(t *testing.T) {

	var eventName, eventCorrelationKey string
	var eventData log.Data
	log.Event = func(name string, correlationKey string, data log.Data) {
		eventName = name
		eventCorrelationKey = correlationKey
		eventData = data
	}

	ctx := context.Background()

	Convey("with empty context", t, func() {
		LogError(ctx, errors.New("test error"), nil)
		So(eventName, ShouldEqual, "error")
		So(eventCorrelationKey, ShouldEqual, "")
		So(eventData, ShouldContainKey, "message")
		So(eventData["message"], ShouldEqual, "test error")
		So(eventData, ShouldContainKey, "error")
		So(eventData["error"], ShouldHaveSameTypeAs, errors.New(""))
		So(eventData["error"].(error).Error(), ShouldEqual, "test error")
		So(eventData, ShouldNotContainKey, reqCaller)
		So(eventData, ShouldNotContainKey, reqUser)

		eventData = log.Data{}
	})

	Convey("context with request id", t, func() {
		contextWithRequestID := context.WithValue(ctx, requestID.ContextKey, "request id")
		LogError(contextWithRequestID, errors.New("test error"), nil)
		So(eventName, ShouldEqual, "error")
		So(eventCorrelationKey, ShouldEqual, "request id")
		So(eventData, ShouldContainKey, "message")
		So(eventData["message"], ShouldEqual, "test error")
		So(eventData, ShouldContainKey, "error")
		So(eventData["error"], ShouldHaveSameTypeAs, errors.New(""))
		So(eventData["error"].(error).Error(), ShouldEqual, "test error")
		So(eventData, ShouldNotContainKey, reqCaller)
		So(eventData, ShouldNotContainKey, reqUser)

		eventData = log.Data{}
	})

	Convey("context with user", t, func() {
		contextWithUser := context.WithValue(ctx, common.UserIdentityKey, "user email")
		LogError(contextWithUser, errors.New("test error"), nil)
		So(eventName, ShouldEqual, "error")
		So(eventCorrelationKey, ShouldEqual, "")
		So(eventData, ShouldContainKey, "message")
		So(eventData["message"], ShouldEqual, "test error")
		So(eventData, ShouldContainKey, "error")
		So(eventData["error"], ShouldHaveSameTypeAs, errors.New(""))
		So(eventData["error"].(error).Error(), ShouldEqual, "test error")
		So(eventData, ShouldContainKey, reqUser)
		So(eventData[reqUser], ShouldHaveSameTypeAs, "")
		So(eventData[reqUser], ShouldEqual, "user email")
		So(eventData, ShouldNotContainKey, reqCaller)

		eventData = log.Data{}
	})

	Convey("context with caller id", t, func() {
		contextWithCaller := context.WithValue(ctx, common.CallerIdentityKey, "api service")
		LogError(contextWithCaller, errors.New("test error"), nil)
		So(eventName, ShouldEqual, "error")
		So(eventCorrelationKey, ShouldEqual, "")
		So(eventData, ShouldContainKey, "message")
		So(eventData["message"], ShouldEqual, "test error")
		So(eventData, ShouldContainKey, "error")
		So(eventData["error"], ShouldHaveSameTypeAs, errors.New(""))
		So(eventData["error"].(error).Error(), ShouldEqual, "test error")
		So(eventData, ShouldContainKey, reqCaller)
		So(eventData[reqCaller], ShouldHaveSameTypeAs, "")
		So(eventData[reqCaller], ShouldEqual, "api service")
		So(eventData, ShouldNotContainKey, reqUser)

		eventData = log.Data{}
	})

	Convey("context with request, user and caller ids", t, func() {
		ctx := getContextWithCallerUserAndRequestIDcontext()
		LogError(ctx, errors.New("test error"), nil)
		So(eventName, ShouldEqual, "error")
		So(eventCorrelationKey, ShouldEqual, "request id")
		So(eventData, ShouldContainKey, "message")
		So(eventData["message"], ShouldEqual, "test error")
		So(eventData, ShouldContainKey, "error")
		So(eventData["error"], ShouldHaveSameTypeAs, errors.New(""))
		So(eventData["error"].(error).Error(), ShouldEqual, "test error")
		So(eventData, ShouldContainKey, reqCaller)
		So(eventData[reqCaller], ShouldHaveSameTypeAs, "")
		So(eventData[reqCaller], ShouldEqual, "api service")
		So(eventData, ShouldContainKey, reqUser)
		So(eventData[reqUser], ShouldHaveSameTypeAs, "")
		So(eventData[reqUser], ShouldEqual, "user email")

		eventData = log.Data{}
	})
}

func TestInfoLog(t *testing.T) {

	var eventName, eventCorrelationKey string
	var eventData log.Data
	log.Event = func(name string, correlationKey string, data log.Data) {
		eventName = name
		eventCorrelationKey = correlationKey
		eventData = data
	}
	Convey("Info", t, func() {
		ctx := context.Background()

		Convey("with empty context", func() {
			LogInfo(ctx, "info message", nil)
			So(eventName, ShouldEqual, "info")
			So(eventCorrelationKey, ShouldEqual, "")
			So(eventData, ShouldContainKey, "message")
			So(eventData["message"], ShouldEqual, "info message")
			So(eventData, ShouldNotContainKey, "caller")
			So(eventData, ShouldNotContainKey, "user")

			eventData = log.Data{}
		})

		Convey("context with request id", func() {
			contextWithRequestID := context.WithValue(ctx, requestID.ContextKey, "request id")
			LogInfo(contextWithRequestID, "info message", nil)
			So(eventName, ShouldEqual, "info")
			So(eventCorrelationKey, ShouldEqual, "request id")
			So(eventData, ShouldContainKey, "message")
			So(eventData["message"], ShouldEqual, "info message")
			So(eventData, ShouldNotContainKey, "caller")
			So(eventData, ShouldNotContainKey, "user")

			eventData = log.Data{}
		})

		Convey("context with user", func() {
			contextWithUser := context.WithValue(ctx, common.UserIdentityKey, "user email")
			LogInfo(contextWithUser, "info message", nil)
			So(eventName, ShouldEqual, "info")
			So(eventCorrelationKey, ShouldEqual, "")
			So(eventData, ShouldContainKey, "message")
			So(eventData["message"], ShouldEqual, "info message")
			So(eventData, ShouldContainKey, reqUser)
			So(eventData[reqUser], ShouldHaveSameTypeAs, "")
			So(eventData[reqUser], ShouldEqual, "user email")
			So(eventData, ShouldNotContainKey, reqCaller)

			eventData = log.Data{}
		})

		Convey("context with caller id", func() {
			contextWithCaller := context.WithValue(ctx, common.CallerIdentityKey, "api service")
			LogInfo(contextWithCaller, "info message", nil)
			So(eventName, ShouldEqual, "info")
			So(eventCorrelationKey, ShouldEqual, "")
			So(eventData, ShouldContainKey, "message")
			So(eventData["message"], ShouldEqual, "info message")
			So(eventData, ShouldContainKey, reqCaller)
			So(eventData[reqCaller], ShouldHaveSameTypeAs, "")
			So(eventData[reqCaller], ShouldEqual, "api service")
			So(eventData, ShouldNotContainKey, reqUser)

			eventData = log.Data{}
		})

		Convey("context with request, user and caller ids", func() {
			ctx := getContextWithCallerUserAndRequestIDcontext()
			LogInfo(ctx, "info message", nil)
			So(eventName, ShouldEqual, "info")
			So(eventCorrelationKey, ShouldEqual, "request id")
			So(eventData, ShouldContainKey, "message")
			So(eventData["message"], ShouldEqual, "info message")
			So(eventData, ShouldContainKey, reqCaller)
			So(eventData[reqCaller], ShouldHaveSameTypeAs, "")
			So(eventData[reqCaller], ShouldEqual, "api service")
			So(eventData, ShouldContainKey, reqUser)
			So(eventData[reqUser], ShouldHaveSameTypeAs, "")
			So(eventData[reqUser], ShouldEqual, "user email")

			eventData = log.Data{}
		})
	})
}

func getContextWithCallerUserAndRequestIDcontext() context.Context {
	ctx := context.WithValue(context.Background(), requestID.ContextKey, "request id")
	ctx = context.WithValue(ctx, common.UserIdentityKey, "user email")
	ctx = context.WithValue(ctx, common.CallerIdentityKey, "api service")
	return ctx
}
