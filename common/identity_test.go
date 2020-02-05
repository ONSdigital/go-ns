package common

import (
	"context"
	"net/http"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/headers"
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

func TestAddAuthHeaders(t *testing.T) {

	Convey("Given a fresh request", t, func() {

		Convey("When AddAuthHeaders is called with no auth", func() {

			r, _ := http.NewRequest("POST", "http://localhost:21800/jobs", nil)
			ctx := context.Background()
			AddAuthHeaders(ctx, r, "")

			Convey("Then the request has no auth headers set", func() {
				userIdentityHeader, err := headers.GetUserIdentity(r)
				So(err, ShouldResemble, headers.ErrHeaderNotFound)
				So(userIdentityHeader, ShouldBeBlank)

				serviceAuthToken, err := headers.GetServiceAuthToken(r)
				So(err, ShouldResemble, headers.ErrHeaderNotFound)
				So(serviceAuthToken, ShouldBeBlank)
			})
		})
		Convey("When AddAuthHeaders is called with a service token", func() {

			serviceToken := "123"

			r, _ := http.NewRequest("POST", "http://localhost:21800/jobs", nil)
			ctx := context.Background()
			AddAuthHeaders(ctx, r, serviceToken)

			Convey("Then the request has the service token header set", func() {
				userIdentityHeader, err := headers.GetUserIdentity(r)
				So(err, ShouldResemble, headers.ErrHeaderNotFound)
				So(userIdentityHeader, ShouldBeBlank)

				serviceAuthToken, err := headers.GetServiceAuthToken(r)
				So(err, ShouldBeNil)
				So(serviceAuthToken, ShouldEqual, serviceToken)
			})
		})

		Convey("When AddAuthHeaders is called with a service token and context has user ID", func() {

			serviceToken := "123"
			userID := "user@test"

			r, _ := http.NewRequest("POST", "http://localhost:21800/jobs", nil)
			ctx := SetUser(context.Background(), userID)
			AddAuthHeaders(ctx, r, serviceToken)

			Convey("Then the request has the service token header set", func() {
				userIdentityHeader, err := headers.GetUserIdentity(r)
				So(err, ShouldBeNil)
				So(userIdentityHeader, ShouldEqual, userID)

				serviceAuthToken, err := headers.GetServiceAuthToken(r)
				So(err, ShouldBeNil)
				So(serviceAuthToken, ShouldEqual, serviceToken)
			})
		})

		Convey("When AddAuthHeaders is called with context that has user ID", func() {

			userID := "user@test"

			r, _ := http.NewRequest("POST", "http://localhost:21800/jobs", nil)
			ctx := SetUser(context.Background(), userID)
			AddAuthHeaders(ctx, r, "")

			Convey("Then the request has the user header set", func() {
				userIdentityHeader, err := headers.GetUserIdentity(r)
				So(err, ShouldBeNil)
				So(userIdentityHeader, ShouldEqual, userID)

				serviceAuthToken, err := headers.GetServiceAuthToken(r)
				So(err, ShouldResemble, headers.ErrHeaderNotFound)
				So(serviceAuthToken, ShouldBeBlank)
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

func TestGetRequestId(t *testing.T) {
	Convey("should return requestID if it exists in the provided context", t, func() {
		ctx := WithRequestId(context.Background(), "666")
		So(ctx.Value(ContextKey("request-id")).(string), ShouldEqual, "666")
	})

	Convey("should return empty value if requestID is not in the provided context", t, func() {
		id := GetRequestId(context.Background())
		So(id, ShouldBeBlank)
	})
}

func TestSetRequestId(t *testing.T) {
	Convey("set request id in empty context", t, func() {
		ctx := WithRequestId(context.Background(), "123")
		So(ctx.Value(ContextKey("request-id")), ShouldEqual, "123")

		Convey("overwrite context request id with new value", func() {
			newCtx := WithRequestId(ctx, "456")
			So(newCtx.Value(ContextKey("request-id")), ShouldEqual, "456")
		})
	})
}
