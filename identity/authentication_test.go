package identity

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"context"
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
		})
	})
}

func TestCheck_emptyIdentity(t *testing.T) {
	Convey("Given a request with an empty identity provided in the request context", t, func() {

		req, err := http.NewRequest("POST", "http://localhost:21800/jobs", nil)

		ctx := context.WithValue(req.Context(), callerIdentityKey, "")
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
		})
	})
}


func TestCheck_identityProvided(t *testing.T) {

	Convey("Given a request with an identity provided in the request context", t, func() {

		req, err := http.NewRequest("POST", "http://localhost:21800/jobs", nil)

		ctx := context.WithValue(req.Context(), callerIdentityKey, "user@ons.gov.uk")
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
		})
	})
}

func TestIsPresent_withIdentity(t *testing.T) {

	Convey("Given a context with an identity", t, func() {

		ctx := context.WithValue(context.Background(), callerIdentityKey, "user@ons.gov.uk")

		Convey("When IsPresent is called with the context", func() {

			identityIsPresent := IsPresent(ctx)

			Convey("Then a 404 response is returned", func() {
				So(identityIsPresent, ShouldBeTrue)
			})
		})
	})
}

func TestIsPresent_withNoIdentity(t *testing.T) {

	Convey("Given a context with no identity", t, func() {

		ctx := context.Background()

		Convey("When IsPresent is called with the context", func() {

			identityIsPresent := IsPresent(ctx)

			Convey("Then the response is false", func() {
				So(identityIsPresent, ShouldBeFalse)
			})
		})
	})
}

func TestIsPresent_withEmptyIdentity(t *testing.T) {
	Convey("Given a context with an empty identity", t, func() {

		ctx := context.WithValue(context.Background(), callerIdentityKey, "")

		Convey("When IsPresent is called with the context", func() {

			identityIsPresent := IsPresent(ctx)

			Convey("Then the response is false", func() {
				So(identityIsPresent, ShouldBeFalse)
			})
		})
	})
}

func TestUser(t *testing.T) {

	Convey("Given a context with a user identity", t, func() {

		ctx := context.WithValue(context.Background(), userIdentityKey, "Frederico")

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

		ctx := context.WithValue(context.Background(), userIdentityKey, "")

		Convey("When User is called with the context", func() {

			user := User(ctx)

			Convey("Then the response is empty", func() {
				So(user, ShouldEqual, "")
			})
		})
	})
}


func TestCaller(t *testing.T) {

	Convey("Given a context with a caller identity", t, func() {

		ctx := context.WithValue(context.Background(), callerIdentityKey, "Frederico")

		Convey("When Caller is called with the context", func() {

			caller := Caller(ctx)

			Convey("Then the response had the caller identity", func() {
				So(caller, ShouldEqual, "Frederico")
			})
		})
	})
}

func TestCaller_noCallerIdentity(t *testing.T) {

	Convey("Given a context with no caller identity", t, func() {

		ctx := context.Background()

		Convey("When Caller is called with the context", func() {

			caller := Caller(ctx)

			Convey("Then the response is empty", func() {
				So(caller, ShouldEqual, "")
			})
		})
	})
}

func TestCaller_emptyCallerIdentity(t *testing.T) {

	Convey("Given a context with an empty caller identity", t, func() {

		ctx := context.WithValue(context.Background(), callerIdentityKey, "")

		Convey("When Caller is called with the context", func() {

			caller := Caller(ctx)

			Convey("Then the response is empty", func() {
				So(caller, ShouldEqual, "")
			})
		})
	})
}