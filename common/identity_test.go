package common

import (
	"context"
	"net/http"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSetUser(t *testing.T) {

	Convey("Given a context", t, func() {

		ctx := context.Background()

		Convey("When SetUser is called", func() {

			user := "someone@ons.gov.uk"
			ctx = SetUser(ctx, user)

			Convey("Then the context had the caller identity", func() {
				So(ctx.Value(UserIdentityKey), ShouldEqual, user)
				So(IsUserPresent(ctx), ShouldBeTrue)
			})
		})
	})
}

func TestUser(t *testing.T) {

	Convey("Given a context with a user identity", t, func() {

		ctx := context.WithValue(context.Background(), UserIdentityKey, "Frederico")

		Convey("When User is called with the context", func() {

			user := User(ctx)

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

			user := User(ctx)

			Convey("Then the response is empty", func() {
				So(user, ShouldEqual, "")
			})
		})
	})
}

func TestUser_emptyUserIdentity(t *testing.T) {

	Convey("Given a context with an empty user identity", t, func() {

		ctx := context.WithValue(context.Background(), UserIdentityKey, "")

		Convey("When User is called with the context", func() {

			user := User(ctx)

			Convey("Then the response is empty", func() {
				So(user, ShouldEqual, "")
			})
		})
	})
}

func TestAddUserHeader(t *testing.T) {

	Convey("Given a request", t, func() {

		r, _ := http.NewRequest("POST", "http://localhost:21800/jobs", nil)

		Convey("When AddUserHeader is called", func() {

			user := "someone@ons.gov.uk"
			AddUserHeader(r, user)

			Convey("Then the request has the user header set", func() {
				So(r.Header.Get(UserHeaderKey), ShouldEqual, user)
			})
		})
	})
}

func TestAddServiceTokenHeader(t *testing.T) {

	Convey("Given a request", t, func() {

		r, _ := http.NewRequest("POST", "http://localhost:21800/jobs", nil)

		Convey("When AddServiceTokenHeader is called", func() {

			serviceToken := "123"
			AddServiceTokenHeader(r, serviceToken)

			Convey("Then the request has the service token header set", func() {
				So(r.Header.Get(AuthHeaderKey), ShouldEqual, BearerPrefix+serviceToken)
			})
		})
	})
}

func TestAddAuthHeaders(t *testing.T) {

	Convey("Given a fresh request", t, func() {

		Convey("When AddAuthHeaders is called with no auth", func() {

			r, _ := http.NewRequest("POST", "http://localhost:21800/jobs", nil)
			ctx := context.Background()
			AddAuthHeaders(ctx, r, "")

			Convey("Then the request has no auth headers set", func() {
				So(r.Header.Get(AuthHeaderKey), ShouldBeBlank)
				So(r.Header.Get(UserHeaderKey), ShouldBeBlank)
			})
		})
		Convey("When AddAuthHeaders is called with a service token", func() {

			serviceToken := "123"

			r, _ := http.NewRequest("POST", "http://localhost:21800/jobs", nil)
			ctx := context.Background()
			AddAuthHeaders(ctx, r, serviceToken)

			Convey("Then the request has the service token header set", func() {
				So(r.Header.Get(AuthHeaderKey), ShouldEqual, BearerPrefix+serviceToken)
				So(r.Header.Get(UserHeaderKey), ShouldBeBlank)
			})
		})

		Convey("When AddAuthHeaders is called with a service token and context has user ID", func() {

			serviceToken := "123"
			userID := "user@test"

			r, _ := http.NewRequest("POST", "http://localhost:21800/jobs", nil)
			ctx := SetUser(context.Background(), userID)
			AddAuthHeaders(ctx, r, serviceToken)

			Convey("Then the request has the service token header set", func() {
				So(r.Header.Get(AuthHeaderKey), ShouldEqual, BearerPrefix+serviceToken)
				So(r.Header.Get(UserHeaderKey), ShouldEqual, userID)
			})
		})

		Convey("When AddAuthHeaders is called with context that has user ID", func() {

			userID := "user@test"

			r, _ := http.NewRequest("POST", "http://localhost:21800/jobs", nil)
			ctx := SetUser(context.Background(), userID)
			AddAuthHeaders(ctx, r, "")

			Convey("Then the request has the user header set", func() {
				So(r.Header.Get(AuthHeaderKey), ShouldBeBlank)
				So(r.Header.Get(UserHeaderKey), ShouldEqual, userID)
			})
		})
	})
}

func TestNewRequestID(t *testing.T) {
	Convey("create a requestID with length of 12", t, func() {
		requestID := NewRequestID(12)
		So(len(requestID), ShouldEqual, 12)

		Convey("create a second requestID with length of 12", func() {
			secondRequestID := NewRequestID(12)
			So(len(secondRequestID), ShouldEqual, 12)
			So(secondRequestID, ShouldNotEqual, requestID)
		})
	})
}
