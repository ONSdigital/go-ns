package identity

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ONSdigital/go-ns/common"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCheck_nilIdentity(t *testing.T) {
	Convey("Given a request with no identity provided in the request context", t, func() {

		req, err := http.NewRequest("POST", "http://localhost:21800/jobs", nil)
		So(err, ShouldBeNil)
		responseRecorder := httptest.NewRecorder()

		handlerCalled := false
		httpHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			handlerCalled = true
		})

		Convey("When the authentication handler is called", func() {

			Check(httpHandler)(responseRecorder, req)

			Convey("Then a 404 response is returned", func() {
				So(responseRecorder.Code, ShouldEqual, http.StatusNotFound)
			})

			Convey("Then the downstream HTTP handler is not called", func() {
				So(handlerCalled, ShouldBeFalse)
			})
		})
	})
}

func TestCheck_emptyIdentity(t *testing.T) {
	Convey("Given a request with an empty identity provided in the request context", t, func() {

		req, err := http.NewRequest("POST", "http://localhost:21800/jobs", nil)

		ctx := context.WithValue(req.Context(), common.CallerIdentityKey, "")
		req = req.WithContext(ctx)

		So(err, ShouldBeNil)
		responseRecorder := httptest.NewRecorder()

		handlerCalled := false
		httpHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			handlerCalled = true
		})

		Convey("When the authentication handler is called", func() {

			Check(httpHandler)(responseRecorder, req)

			Convey("Then a 404 response is returned", func() {
				So(responseRecorder.Code, ShouldEqual, http.StatusNotFound)
			})

			Convey("Then the downstream HTTP handler is not called", func() {
				So(handlerCalled, ShouldBeFalse)
			})
		})
	})
}

func TestCheck_identityProvided(t *testing.T) {

	Convey("Given a request with an identity provided in the request context", t, func() {

		req, err := http.NewRequest("POST", "http://localhost:21800/jobs", nil)

		ctx := context.WithValue(req.Context(), common.CallerIdentityKey, "user@ons.gov.uk")
		req = req.WithContext(ctx)

		So(err, ShouldBeNil)
		responseRecorder := httptest.NewRecorder()

		handlerCalled := false
		httpHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			handlerCalled = true
		})

		Convey("When the authentication handler is called", func() {

			Check(httpHandler)(responseRecorder, req)

			Convey("Then the response is true", func() {
				So(responseRecorder.Code, ShouldEqual, http.StatusOK)
			})

			Convey("Then the downstream HTTP handler is called", func() {
				So(handlerCalled, ShouldBeTrue)
			})
		})
	})
}

func TestIsPresent_withIdentity(t *testing.T) {

	Convey("Given a context with an identity", t, func() {

		ctx := context.WithValue(context.Background(), common.CallerIdentityKey, "user@ons.gov.uk")

		Convey("When IsPresent is called with the context", func() {

			identityIsPresent := common.IsPresent(ctx)

			Convey("Then the response is true", func() {
				So(identityIsPresent, ShouldBeTrue)
			})
		})
	})
}

func TestIsPresent_withNoIdentity(t *testing.T) {

	Convey("Given a context with no identity", t, func() {

		ctx := context.Background()

		Convey("When IsPresent is called with the context", func() {

			identityIsPresent := common.IsPresent(ctx)

			Convey("Then the response is false", func() {
				So(identityIsPresent, ShouldBeFalse)
			})
		})
	})
}

func TestIsPresent_withEmptyIdentity(t *testing.T) {
	Convey("Given a context with an empty identity", t, func() {

		ctx := context.WithValue(context.Background(), common.CallerIdentityKey, "")

		Convey("When IsPresent is called with the context", func() {

			identityIsPresent := common.IsPresent(ctx)

			Convey("Then the response is false", func() {
				So(identityIsPresent, ShouldBeFalse)
			})
		})
	})
}

func TestSetUser(t *testing.T) {

	Convey("Given a context", t, func() {

		ctx := context.Background()

		Convey("When SetUser is called", func() {

			user := "someone@ons.gov.uk"
			ctx = common.SetUser(ctx, user)

			Convey("Then the response had the caller identity", func() {
				So(ctx.Value(common.UserIdentityKey), ShouldEqual, user)
			})
		})
	})
}

func TestUser(t *testing.T) {

	Convey("Given a context with a user identity", t, func() {

		ctx := context.WithValue(context.Background(), common.UserIdentityKey, "Frederico")

		Convey("When User is called with the context", func() {

			user := common.User(ctx)

			Convey("Then the response had the user identity", func() {
				So(user, ShouldEqual, "Frederico")
			})
		})
	})
}

func TestUser_noUserIdentity(t *testing.T) {

	Convey("Given a context with no user identity", t, func() {

		ctx := context.Background()

		Convey("When User is called with the context", func() {

			user := common.User(ctx)

			Convey("Then the response is empty", func() {
				So(user, ShouldEqual, "")
			})
		})
	})
}

func TestUser_emptyUserIdentity(t *testing.T) {

	Convey("Given a context with an empty user identity", t, func() {

		ctx := context.WithValue(context.Background(), common.UserIdentityKey, "")

		Convey("When User is called with the context", func() {

			user := common.User(ctx)

			Convey("Then the response is empty", func() {
				So(user, ShouldEqual, "")
			})
		})
	})
}

func TestCaller(t *testing.T) {

	Convey("Given a context with a caller identity", t, func() {

		ctx := context.WithValue(context.Background(), common.CallerIdentityKey, "Frederico")

		Convey("When Caller is called with the context", func() {

			caller := common.Caller(ctx)

			Convey("Then the response had the caller identity", func() {
				So(caller, ShouldEqual, "Frederico")
			})
		})
	})
}

func TestSetCaller(t *testing.T) {

	Convey("Given a context", t, func() {

		ctx := context.Background()

		Convey("When SetCaller is called", func() {

			caller := "dp-dataset-api"
			ctx = common.SetCaller(ctx, caller)

			Convey("Then the response had the caller identity", func() {
				So(ctx.Value(common.CallerIdentityKey), ShouldEqual, caller)
			})
		})
	})
}

func TestCaller_noCallerIdentity(t *testing.T) {

	Convey("Given a context with no caller identity", t, func() {

		ctx := context.Background()

		Convey("When Caller is called with the context", func() {

			caller := common.Caller(ctx)

			Convey("Then the response is empty", func() {
				So(caller, ShouldEqual, "")
			})
		})
	})
}

func TestCaller_emptyCallerIdentity(t *testing.T) {

	Convey("Given a context with an empty caller identity", t, func() {

		ctx := context.WithValue(context.Background(), common.CallerIdentityKey, "")

		Convey("When Caller is called with the context", func() {

			caller := common.Caller(ctx)

			Convey("Then the response is empty", func() {
				So(caller, ShouldEqual, "")
			})
		})
	})
}

func TestAddUserHeader(t *testing.T) {

	Convey("Given a request", t, func() {

		r, _ := http.NewRequest("POST", "http://localhost:21800/jobs", nil)

		Convey("When AddUserHeader is called", func() {

			user := "someone@ons.gov.uk"
			common.AddUserHeader(r, user)

			Convey("Then the request has the user header set", func() {
				So(r.Header.Get(common.UserHeaderKey), ShouldEqual, user)
			})
		})
	})
}

func TestAddServiceTokenHeader(t *testing.T) {

	Convey("Given a request", t, func() {

		r, _ := http.NewRequest("POST", "http://localhost:21800/jobs", nil)

		Convey("When AddServiceTokenHeader is called", func() {

			serviceToken := "123"
			common.AddServiceTokenHeader(r, serviceToken)

			Convey("Then the request has the service token header set", func() {
				So(r.Header.Get(common.AuthHeaderKey), ShouldEqual, common.BearerPrefix+serviceToken)
			})
		})
	})
}
