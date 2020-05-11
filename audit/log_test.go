package audit

import (
	"context"
	"testing"

	nethttp "github.com/ONSdigital/dp-net/http"
	"github.com/ONSdigital/log.go/log"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	ctx = context.Background()
)

func TestAddLogData(t *testing.T) {
	Convey("with empty context", t, func() {
		data := addLogData(ctx, nil)

		So(data, ShouldResemble, log.Data{})
	})

	Convey("context with user", t, func() {
		contextWithUser := context.WithValue(ctx, nethttp.UserIdentityKey, "user email")
		data := addLogData(contextWithUser, nil)

		So(data, ShouldResemble, log.Data{reqUser: "user email"})
	})

	Convey("context with caller id", t, func() {
		contextWithCaller := context.WithValue(ctx, nethttp.CallerIdentityKey, "api service")
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
	ctx := nethttp.WithRequestId(context.Background(), "request id")
	ctx = context.WithValue(ctx, nethttp.UserIdentityKey, "user email")
	ctx = context.WithValue(ctx, nethttp.CallerIdentityKey, "api service")
	return ctx
}
