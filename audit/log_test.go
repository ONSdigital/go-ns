package audit

import (
	"context"
	"testing"

	"github.com/ONSdigital/go-ns/common"
	"github.com/ONSdigital/log.go/v2/log"
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
	ctx := common.WithRequestId(context.Background(), "request id")
	ctx = context.WithValue(ctx, common.UserIdentityKey, "user email")
	ctx = context.WithValue(ctx, common.CallerIdentityKey, "api service")
	return ctx
}
