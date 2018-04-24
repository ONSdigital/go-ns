package audit

import (
	"context"
	"fmt"
	"github.com/ONSdigital/go-ns/common"
	"github.com/ONSdigital/go-ns/identity"
	"github.com/ONSdigital/go-ns/log"
	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const (
	namespace   = "audit-test"
	auditAction = "test"
	auditResult = "success"
	user        = "some-user"
	service     = "some-service"
)

type HandlerMock struct {
	invocations []*http.Request
	handleFunc  func(http.ResponseWriter, *http.Request)
}

func (h *HandlerMock) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.invocations = append(h.invocations, r)
}

func TestAuditor_RecordNoUser(t *testing.T) {
	Convey("given no user identity exists in the provided context", t, func() {
		producer := &OutboundProducerMock{}

		auditor := New(producer, namespace)

		// record the audit event
		err := auditor.Record(context.Background(), auditAction, auditResult, nil)

		So(err, ShouldBeNil)
		So(len(producer.OutputCalls()), ShouldEqual, 0)
	})
}

func TestAuditor_RecordNoEventInContext(t *testing.T) {
	Convey("given no audit event exists in the provided context", t, func() {
		producer := &OutboundProducerMock{}

		auditor := New(producer, namespace)

		// record the audit event
		err := auditor.Record(identity.SetUser(context.Background(), "Egon Spengler"), auditAction, auditResult, nil)

		expectedErr := newAuditError("no event found in context.Context", auditAction, auditResult, nil)
		So(err, ShouldResemble, expectedErr)
		So(len(producer.OutputCalls()), ShouldEqual, 0)
	})
}

func TestAuditor_RecordIncorrectEventFormat(t *testing.T) {
	Convey("given context event is not the expected format", t, func() {
		producer := &OutboundProducerMock{}

		auditor := New(producer, namespace)

		ctx := context.WithValue(context.Background(), contextKey("audit"), "this is not an audit event")
		ctx = identity.SetUser(ctx, "Scrooge McDuck")

		// record the audit event
		err := auditor.Record(ctx, auditAction, auditResult, nil)

		expectedErr := newAuditError("context.Context audit event was not in the expected format", auditAction, auditResult, nil)
		So(err, ShouldResemble, expectedErr)
		So(len(producer.OutputCalls()), ShouldEqual, 0)
	})
}

func TestAuditor_RecordAvroMarshalError(t *testing.T) {
	Convey("given there is an error converting the audit event to into avro", t, func() {
		producer := &OutboundProducerMock{}
		auditor := New(producer, namespace)

		auditor.marshalToAvro = func(s interface{}) ([]byte, error) {
			return nil, errors.New("avro marshal error")
		}

		// record the audit event
		err := auditor.Record(setUpContext(), auditAction, auditResult, nil)

		expectedErr := newAuditError("error marshalling event to arvo", auditAction, auditResult, nil)
		So(err, ShouldResemble, expectedErr)
		So(len(producer.OutputCalls()), ShouldEqual, 0)
	})
}

func TestAuditor_RecordSuccess(t *testing.T) {
	Convey("given a valid audit event exists in context", t, func() {
		output := make(chan []byte, 1)

		producer := &OutboundProducerMock{
			OutputFunc: func() chan []byte {
				return output
			},
		}

		auditor := New(producer, namespace)

		var results []byte

		// record the audit event
		err := auditor.Record(setUpContext(), auditAction, auditResult, common.Params{"ID": "12345"})

		select {
		case results = <-output:
			log.Info("output received", nil)
		case <-time.After(time.Second * 5):
			log.Debug("failing test due to timeout, expected output channel to receive event but none", nil)
			t.FailNow()
		}

		So(err, ShouldBeNil)
		So(len(producer.OutputCalls()), ShouldEqual, 1)

		var actualEvent Event
		err = EventSchema.Unmarshal(results, &actualEvent)
		if err != nil {
			log.ErrorC("avro unmarshal error", err, nil)
			t.FailNow()
		}

		So(actualEvent.Namespace, ShouldEqual, namespace)
		So(actualEvent.AttemptedAction, ShouldEqual, auditAction)
		So(actualEvent.Result, ShouldEqual, auditResult)
		So(actualEvent.Created, ShouldNotBeEmpty)
		So(actualEvent.Service, ShouldBeEmpty)
		So(actualEvent.User, ShouldEqual, user)
		So(actualEvent.Params, ShouldResemble, []keyValuePair{{"ID", "12345"}})
	})
}

func TestAuditor_RecordEmptyAction(t *testing.T) {
	Convey("given Record is called with an empty action value then the expected error is returned", t, func() {
		producer := &OutboundProducerMock{}

		auditor := New(producer, namespace)

		err := auditor.Record(setUpContext(), "", "", nil)

		So(len(producer.OutputCalls()), ShouldEqual, 0)
		expectedErr := newAuditError("attempted action required but was empty", "nil", "", nil)
		So(err, ShouldResemble, expectedErr)
	})
}

func TestAuditor_RecordEmptyResult(t *testing.T) {
	Convey("given Record is called with an empty result value then the expected error is returned", t, func() {
		producer := &OutboundProducerMock{}

		auditor := New(producer, namespace)

		err := auditor.Record(setUpContext(), auditAction, "", nil)

		So(len(producer.OutputCalls()), ShouldEqual, 0)
		expectedErr := newAuditError("result required but was empty", "test", "", nil)
		So(err, ShouldResemble, expectedErr)
	})
}

func TestAuditor_GetEvent(t *testing.T) {
	Convey("given no event exists in the provided context", t, func() {
		auditor := New(nil, namespace)
		event, err := auditor.GetEvent(context.Background())
		So(event, ShouldBeNil)
		So(err.Error(), ShouldEqual, "no event found in context.Context")
	})

	Convey("given the context event is not the correct type", t, func() {
		auditor := New(nil, namespace)

		event, err := auditor.GetEvent(context.WithValue(context.Background(), contextKey("audit"), "I AM NOT AN EVENT"))
		So(event, ShouldBeNil)
		So(err.Error(), ShouldEqual, "context.Context audit event was not in the expected format")
	})

	Convey("given the context contains a valid event", t, func() {
		auditor := New(nil, namespace)

		event, err := auditor.GetEvent(setUpContext())
		So(event, ShouldResemble,
			&Event{
				Namespace: "audit-test",
				User:      user,
				//				Service:   service,
			})
		So(err, ShouldBeNil)
	})
}

func TestAuditor_Interceptor(t *testing.T) {
	auditor := New(nil, namespace)

	nextHandler := &HandlerMock{invocations: make([]*http.Request, 0)}

	r := httptest.NewRequest(http.MethodGet, "/bob", nil)
	r.WithContext(context.Background())
	r.Header.Add(requestIDHeader, "666")

	Convey("given a valid request the expected base audit event is added to the request context", t, func() {
		interceptHandlerFunc := auditor.Interceptor()
		interceptHandlerFunc(nextHandler).ServeHTTP(nil, r)

		So(len(nextHandler.invocations), ShouldEqual, 1)

		ctxObj := nextHandler.invocations[0].Context().Value(contextKey("audit"))
		So(ctxObj, ShouldNotBeNil)

		auditEvent, ok := ctxObj.(Event)
		if !ok {
			log.Debug("failing test, expected audit.Event but was not", nil)
			t.FailNow()
		}

		So(auditEvent.RequestID, ShouldEqual, "666")
		So(auditEvent.Namespace, ShouldEqual, namespace)
		So(auditEvent.User, ShouldEqual, "")
	})
}

func Test_newAuditError(t *testing.T) {
	Convey("given no values are provided", t, func() {
		actual := newAuditError("", "", "", nil)

		Convey("then an error with default values is returned", func() {
			expected := Error{
				Cause:  "nil",
				Action: "nil",
				Result: "nil",
				Params: []keyValuePair{},
			}

			So(actual, ShouldResemble, expected)
		})

		Convey("and Error() returns the expected value", func() {
			fmt.Println(actual.Error())
			So(actual.Error(), ShouldEqual, "unable to audit event, action: nil, result: nil, cause: nil, params: []")
		})
	})

	Convey("given valid values for all fields", t, func() {
		actual := newAuditError("_cause", "_action", "_result", common.Params{
			"bbb": "bbb",
			"aaa": "aaa",
			"ccc": "ccc",
		})

		expected := Error{
			Cause:  "_cause",
			Action: "_action",
			Result: "_result",
			Params: []keyValuePair{
				{"bbb", "bbb"},
				{"aaa", "aaa"},
				{"ccc", "ccc"},
			},
		}

		So(actual.Cause, ShouldEqual, expected.Cause)
		So(actual.Action, ShouldEqual, expected.Action)
		So(actual.Result, ShouldEqual, expected.Result)

		// verify that the parameters are in the expected order
		So(actual.Params[0], ShouldResemble, keyValuePair{"aaa", "aaa"})
		So(actual.Params[1], ShouldResemble, keyValuePair{"bbb", "bbb"})
		So(actual.Params[2], ShouldResemble, keyValuePair{"ccc", "ccc"})

		expectedStr := "unable to audit event, action: _action, result: _result, cause: _cause, params: [{Key:aaa Value:aaa} {Key:bbb Value:bbb} {Key:ccc Value:ccc}]"
		So(actual.Error(), ShouldEqual, expectedStr)
	})
}

func setUpContext() context.Context {
	ctx := context.WithValue(context.Background(), contextKey("audit"), Event{
		Namespace: namespace,
		User:      user,
	})
	ctx = identity.SetCaller(ctx, service)
	ctx = identity.SetUser(ctx, user)
	return ctx
}
