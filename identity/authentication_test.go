package identity

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"bytes"
	"io"

	nethttp "github.com/ONSdigital/dp-net/http"
	"github.com/ONSdigital/go-ns/audit"
	"github.com/ONSdigital/go-ns/audit/auditortest"
	"github.com/ONSdigital/go-ns/common"
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	testAction         = "test-action"
	testCallerIdentity = "user@ons.gov.uk"
)

func TestCheck_nilIdentity(t *testing.T) {
	Convey("Given a request with no identity provided in the request context", t, func() {
		req, err := http.NewRequest("POST", "http://localhost:21800/datasets/123/editions/2017/version/3", bytes.NewBufferString("some body content"))
		So(err, ShouldBeNil)

		vars := map[string]string{
			"id":      "123",
			"edition": "2017",
			"version": "3",
		}

		req = mux.SetURLVars(req, vars)
		responseRecorder := httptest.NewRecorder()

		handlerCalled := false
		httpHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			handlerCalled = true
		})

		Convey("When the authentication handler is called", func() {
			auditor := auditortest.New()
			Check(auditor, testAction, httpHandler)(responseRecorder, req)

			Convey("Then a 401 response is returned", func() {
				So(responseRecorder.Code, ShouldEqual, http.StatusUnauthorized)

				auditParams := common.Params{"dataset_id": "123", "edition": "2017", "version": "3"}
				auditor.AssertRecordCalls(
					auditortest.Expected{Action: testAction, Result: audit.Attempted, Params: auditParams},
					auditortest.Expected{Action: testAction, Result: audit.Unsuccessful, Params: auditParams},
				)
			})

			Convey("Then the downstream HTTP handler is not called", func() {
				So(handlerCalled, ShouldBeFalse)
			})

			Convey("Then the request body has been drained", func() {
				_, err := req.Body.Read(make([]byte, 1))
				So(err, ShouldEqual, io.EOF)
			})
		})
	})
}

func TestCheck_emptyIdentity(t *testing.T) {
	Convey("Given a request with an empty identity provided in the request context", t, func() {

		req, err := http.NewRequest("POST", "http://localhost:21800/jobs", bytes.NewBufferString("some body content"))

		ctx := context.WithValue(req.Context(), nethttp.CallerIdentityKey, "")
		req = req.WithContext(ctx)

		So(err, ShouldBeNil)
		responseRecorder := httptest.NewRecorder()

		handlerCalled := false
		httpHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			handlerCalled = true
		})

		Convey("When the authentication handler is called", func() {
			auditor := auditortest.New()
			Check(auditor, testAction, httpHandler)(responseRecorder, req)

			Convey("Then a 401 response is returned", func() {
				So(responseRecorder.Code, ShouldEqual, http.StatusUnauthorized)

				auditParams := common.Params{}
				auditor.AssertRecordCalls(
					auditortest.Expected{Action: testAction, Result: audit.Attempted, Params: auditParams},
					auditortest.Expected{Action: testAction, Result: audit.Unsuccessful, Params: auditParams},
				)
			})

			Convey("Then the downstream HTTP handler is not called", func() {
				So(handlerCalled, ShouldBeFalse)
			})

			Convey("Then the request body has been drained", func() {
				_, err := req.Body.Read(make([]byte, 1))
				So(err, ShouldEqual, io.EOF)
			})
		})
	})
}

func TestCheck_identityProvided(t *testing.T) {

	Convey("Given a request with an identity provided in the request context", t, func() {

		req, err := http.NewRequest("POST", "http://localhost:21800/jobs", bytes.NewBufferString("some body content"))

		ctx := context.WithValue(req.Context(), nethttp.CallerIdentityKey, testCallerIdentity)
		req = req.WithContext(ctx)

		So(err, ShouldBeNil)
		responseRecorder := httptest.NewRecorder()

		handlerCalled := false
		httpHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			handlerCalled = true
		})

		Convey("When the authentication handler is called", func() {
			auditor := auditortest.New()
			Check(auditor, testAction, httpHandler)(responseRecorder, req)

			Convey("Then the response is true", func() {
				So(responseRecorder.Code, ShouldEqual, http.StatusOK)

				auditParams := common.Params{"caller_identity": testCallerIdentity}
				auditor.AssertRecordCalls(
					auditortest.Expected{Action: testAction, Result: audit.Attempted, Params: auditParams},
				)
			})

			Convey("Then the downstream HTTP handler is called", func() {
				So(handlerCalled, ShouldBeTrue)
			})
		})
	})
}

func TestCheck_AuditFailure(t *testing.T) {
	Convey("Given a request with an identity provided in the request context", t, func() {

		req, err := http.NewRequest("POST", "http://localhost:21800/jobs", bytes.NewBufferString("some body content"))

		ctx := context.WithValue(req.Context(), nethttp.CallerIdentityKey, testCallerIdentity)
		req = req.WithContext(ctx)

		So(err, ShouldBeNil)
		responseRecorder := httptest.NewRecorder()

		handlerCalled := false
		httpHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			handlerCalled = true
		})

		Convey("When the authentication handler is called and fails to audit attempt successfully", func() {
			auditor := auditortest.NewErroring(testAction, audit.Attempted)
			Check(auditor, testAction, httpHandler)(responseRecorder, req)

			Convey("Then a 500 response is returned", func() {
				So(responseRecorder.Code, ShouldEqual, http.StatusInternalServerError)
				So(responseRecorder.Body.String(), ShouldContainSubstring, "internal error")

				auditParams := common.Params{"caller_identity": testCallerIdentity}
				auditor.AssertRecordCalls(
					auditortest.Expected{Action: testAction, Result: audit.Attempted, Params: auditParams},
				)
			})

			Convey("Then the downstream HTTP handler is not called", func() {
				So(handlerCalled, ShouldBeFalse)
			})

			Convey("Then the request body has been drained", func() {
				_, err := req.Body.Read(make([]byte, 1))
				So(err, ShouldEqual, io.EOF)
			})
		})
	})

	Convey("Given a request with an empty identity provided in the request context", t, func() {

		req, err := http.NewRequest("POST", "http://localhost:21800/jobs", bytes.NewBufferString("some body content"))

		ctx := context.WithValue(req.Context(), nethttp.CallerIdentityKey, "")
		req = req.WithContext(ctx)

		So(err, ShouldBeNil)
		responseRecorder := httptest.NewRecorder()

		handlerCalled := false
		httpHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			handlerCalled = true
		})

		Convey("When the authentication handler is called and fails to audit attempt successfully", func() {
			auditor := auditortest.NewErroring(testAction, audit.Unsuccessful)
			Check(auditor, testAction, httpHandler)(responseRecorder, req)

			Convey("Then a 500 response is returned", func() {
				So(responseRecorder.Code, ShouldEqual, http.StatusInternalServerError)
				So(responseRecorder.Body.String(), ShouldContainSubstring, "internal error")

				auditParams := common.Params{}
				auditor.AssertRecordCalls(
					auditortest.Expected{Action: testAction, Result: audit.Attempted, Params: auditParams},
					auditortest.Expected{Action: testAction, Result: audit.Unsuccessful, Params: auditParams},
				)
			})

			Convey("Then the downstream HTTP handler is not called", func() {
				So(handlerCalled, ShouldBeFalse)
			})

			Convey("Then the request body has been drained", func() {
				_, err := req.Body.Read(make([]byte, 1))
				So(err, ShouldEqual, io.EOF)
			})
		})
	})
}

func TestIsPresent_withIdentity(t *testing.T) {

	Convey("Given a context with an identity", t, func() {

		ctx := context.WithValue(context.Background(), nethttp.CallerIdentityKey, testCallerIdentity)

		Convey("When IsCallerPresent is called with the context", func() {

			identityIsPresent := nethttp.IsCallerPresent(ctx)

			Convey("Then the response is true", func() {
				So(identityIsPresent, ShouldBeTrue)
			})
		})
	})
}

func TestIsPresent_withNoIdentity(t *testing.T) {

	Convey("Given a context with no identity", t, func() {

		ctx := context.Background()

		Convey("When IsCallerPresent is called with the context", func() {

			identityIsPresent := nethttp.IsCallerPresent(ctx)

			Convey("Then the response is false", func() {
				So(identityIsPresent, ShouldBeFalse)
			})
		})
	})
}

func TestIsPresent_withEmptyIdentity(t *testing.T) {
	Convey("Given a context with an empty identity", t, func() {

		ctx := context.WithValue(context.Background(), nethttp.CallerIdentityKey, "")

		Convey("When IsCallerPresent is called with the context", func() {

			identityIsPresent := nethttp.IsCallerPresent(ctx)

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
			ctx = nethttp.SetUser(ctx, user)

			Convey("Then the response had the caller identity", func() {
				So(ctx.Value(nethttp.UserIdentityKey), ShouldEqual, user)
			})
		})
	})
}

func TestUser(t *testing.T) {

	Convey("Given a context with a user identity", t, func() {

		ctx := context.WithValue(context.Background(), nethttp.UserIdentityKey, "Frederico")

		Convey("When User is called with the context", func() {

			user := nethttp.User(ctx)

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

			user := nethttp.User(ctx)

			Convey("Then the response is empty", func() {
				So(user, ShouldEqual, "")
			})
		})
	})
}

func TestUser_emptyUserIdentity(t *testing.T) {

	Convey("Given a context with an empty user identity", t, func() {

		ctx := context.WithValue(context.Background(), nethttp.UserIdentityKey, "")

		Convey("When User is called with the context", func() {

			user := nethttp.User(ctx)

			Convey("Then the response is empty", func() {
				So(user, ShouldEqual, "")
			})
		})
	})
}

func TestCaller(t *testing.T) {

	Convey("Given a context with a caller identity", t, func() {

		ctx := context.WithValue(context.Background(), nethttp.CallerIdentityKey, "Frederico")

		Convey("When Caller is called with the context", func() {

			caller := nethttp.Caller(ctx)

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
			ctx = nethttp.SetCaller(ctx, caller)

			Convey("Then the response had the caller identity", func() {
				So(ctx.Value(nethttp.CallerIdentityKey), ShouldEqual, caller)
			})
		})
	})
}

func TestCaller_noCallerIdentity(t *testing.T) {

	Convey("Given a context with no caller identity", t, func() {

		ctx := context.Background()

		Convey("When Caller is called with the context", func() {

			caller := nethttp.Caller(ctx)

			Convey("Then the response is empty", func() {
				So(caller, ShouldEqual, "")
			})
		})
	})
}

func TestCaller_emptyCallerIdentity(t *testing.T) {

	Convey("Given a context with an empty caller identity", t, func() {

		ctx := context.WithValue(context.Background(), nethttp.CallerIdentityKey, "")

		Convey("When Caller is called with the context", func() {

			caller := nethttp.Caller(ctx)

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
			nethttp.AddUserHeader(r, user)

			Convey("Then the request has the user header set", func() {
				So(r.Header.Get(nethttp.UserHeaderKey), ShouldEqual, user)
			})
		})
	})
}

func TestAddServiceTokenHeader(t *testing.T) {

	Convey("Given a request", t, func() {

		r, _ := http.NewRequest("POST", "http://localhost:21800/jobs", nil)

		Convey("When AddServiceTokenHeader is called", func() {

			serviceToken := "123"
			nethttp.AddServiceTokenHeader(r, serviceToken)

			Convey("Then the request has the service token header set", func() {
				So(r.Header.Get(nethttp.AuthHeaderKey), ShouldEqual, nethttp.BearerPrefix+serviceToken)
			})
		})
	})
}
