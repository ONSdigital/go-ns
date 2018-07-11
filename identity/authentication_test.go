package identity

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ONSdigital/go-ns/audit"
	"github.com/ONSdigital/go-ns/audit/auditortest"
	"github.com/ONSdigital/go-ns/common"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	testAction         = "test-action"
	testCallerIdentity = "user@ons.gov.uk"
)

func TestCheck_nilIdentity(t *testing.T) {
	Convey("Given a request with no identity provided in the request context", t, func() {

		req, err := http.NewRequest("POST", "http://localhost:21800/datasets/123/editions/2017", nil)
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

				auditParams := common.Params{"dataset_id": "123", "edition": "2017"}
				auditor.AssertRecordCalls(
					auditortest.Expected{Action: testAction, Result: audit.Attempted, Params: auditParams},
					auditortest.Expected{Action: testAction, Result: audit.Unsuccessful, Params: auditParams},
				)
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
		})
	})
}

func TestCheck_identityProvided(t *testing.T) {

	Convey("Given a request with an identity provided in the request context", t, func() {

		req, err := http.NewRequest("POST", "http://localhost:21800/jobs", nil)

		ctx := context.WithValue(req.Context(), common.CallerIdentityKey, testCallerIdentity)
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

		req, err := http.NewRequest("POST", "http://localhost:21800/jobs", nil)

		ctx := context.WithValue(req.Context(), common.CallerIdentityKey, testCallerIdentity)
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
		})
	})

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
		})
	})
}

func TestIsPresent_withIdentity(t *testing.T) {

	Convey("Given a context with an identity", t, func() {

		ctx := context.WithValue(context.Background(), common.CallerIdentityKey, testCallerIdentity)

		Convey("When IsCallerPresent is called with the context", func() {

			identityIsPresent := common.IsCallerPresent(ctx)

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

			identityIsPresent := common.IsCallerPresent(ctx)

			Convey("Then the response is false", func() {
				So(identityIsPresent, ShouldBeFalse)
			})
		})
	})
}

func TestIsPresent_withEmptyIdentity(t *testing.T) {
	Convey("Given a context with an empty identity", t, func() {

		ctx := context.WithValue(context.Background(), common.CallerIdentityKey, "")

		Convey("When IsCallerPresent is called with the context", func() {

			identityIsPresent := common.IsCallerPresent(ctx)

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
