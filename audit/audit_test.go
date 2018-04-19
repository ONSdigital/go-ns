package audit

import (
	"context"
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
)

type HandlerMock struct {
	invocations []*http.Request
	handleFunc  func(http.ResponseWriter, *http.Request)
}

func (h *HandlerMock) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.invocations = append(h.invocations, r)
}

func TestAuditor_Record(t *testing.T) {
	Convey("given no audit event exists in the provided context", t, func() {
		producer := &OutboundProducerMock{}

		auditor := New(producer, namespace, "test")

		// record the audit event
		err := auditor.Record(context.Background(), auditAction, auditResult, nil)

		expectedErr := auditError("no event found in context.Context", auditAction, nil)
		So(err.Error(), ShouldEqual, expectedErr.Error())
		So(len(producer.OutputCalls()), ShouldEqual, 0)
	})

	Convey("given there is an error converting the audit event to into avro", t, func() {
		producer := &OutboundProducerMock{}
		auditor := New(producer, namespace, "test")

		auditor.marshalToAvro = func(s interface{}) ([]byte, error) {
			return nil, errors.New("avro marshal error")
		}

		// record the audit event
		err := auditor.Record(setUpContext(), auditAction, auditResult, nil)

		expectedErr := auditError("error marshalling event to arvo", auditAction, nil)
		So(err.Error(), ShouldEqual, expectedErr.Error())
		So(len(producer.OutputCalls()), ShouldEqual, 0)
	})

	Convey("given context event is not the expected format", t, func() {
		producer := &OutboundProducerMock{}

		auditor := New(producer, namespace, "test")

		// record the audit event
		err := auditor.Record(context.WithValue(context.Background(), contextKey("audit"), "this is not an audit event"), auditAction, auditResult, nil)

		expectedErr := auditError("context.Context audit event was not in the expected format", auditAction, nil)
		So(err.Error(), ShouldEqual, expectedErr.Error())
		So(len(producer.OutputCalls()), ShouldEqual, 0)
	})

	Convey("given a valid audit event exists in context", t, func() {
		output := make(chan []byte, 1)

		producer := &OutboundProducerMock{
			OutputFunc: func() chan []byte {
				return output
			},
		}

		auditor := New(producer, namespace, "test")

		var results []byte

		// record the audit event
		err := auditor.Record(setUpContext(), auditAction, auditResult, Params{"ID": "12345"})

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

		So(actualEvent.Namespace, ShouldEqual, "audit-test")
		So(actualEvent.AttemptedAction, ShouldEqual, "test")
		So(actualEvent.Result, ShouldEqual, "success")
		So(actualEvent.Created, ShouldNotBeEmpty)
		So(actualEvent.Service, ShouldEqual, "some-service")
		So(actualEvent.User, ShouldEqual, "some-user")
		So(actualEvent.Params, ShouldResemble, []keyValuePair{{"ID", "12345"}})
	})

	Convey("given Record is called with an empty action value then the expected error is returned", t, func() {
		producer := &OutboundProducerMock{}

		auditor := New(producer, namespace, "test")

		err := auditor.Record(nil, "", "", nil)

		So(len(producer.OutputCalls()), ShouldEqual, 0)
		expectedErr := auditError("attempted action is required field", "nil", nil)
		So(err.Error(), ShouldEqual, expectedErr.Error())
	})

	Convey("given Record is called with an empty result value then the expected error is returned", t, func() {
		producer := &OutboundProducerMock{}

		auditor := New(producer, namespace, "test")

		err := auditor.Record(nil, auditAction, "", nil)

		So(len(producer.OutputCalls()), ShouldEqual, 0)
		expectedErr := auditError("result is required field", "test", nil)
		So(err.Error(), ShouldEqual, expectedErr.Error())
	})
}

func TestAuditor_GetEvent(t *testing.T) {
	Convey("given a no event exists in the provided context", t, func() {
		auditor := New(nil, namespace, "test")
		event, err := auditor.GetEvent(context.Background())
		So(event, ShouldBeNil)
		So(err.Error(), ShouldEqual, "no event found in context.Context")
	})

	Convey("given a the context event is not the correct type", t, func() {
		auditor := New(nil, namespace, "test")

		event, err := auditor.GetEvent(context.WithValue(context.Background(), contextKey("audit"), "I AM NOT AN EVENT"))
		So(event, ShouldBeNil)
		So(err.Error(), ShouldEqual, "context.Context audit event was not in the expected format")
	})

	Convey("given a the context contains a valid event", t, func() {
		auditor := New(nil, namespace, "test")

		event, err := auditor.GetEvent(setUpContext())
		So(event, ShouldResemble, &Event{Namespace: "audit-test"})
		So(err, ShouldBeNil)
	})
}

func TestAuditor_Interceptor(t *testing.T) {
	auditor := New(nil, "test", "")

	nextHandler := &HandlerMock{invocations: make([]*http.Request, 0)}

	r := httptest.NewRequest(http.MethodGet, "/bob", nil)
	r.WithContext(context.Background())
	r.Header.Add(requestIDHeader, "666")
	r.Header.Add(fromHeader, "me")

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
		So(auditEvent.Namespace, ShouldEqual, "test")
		So(auditEvent.User, ShouldEqual, "me")
	})
}

func setUpContext() context.Context {
	ctx := context.WithValue(context.Background(), contextKey("audit"), Event{Namespace: "audit-test"})
	ctx = identity.SetCaller(ctx, "some-service")
	ctx = identity.SetUser(ctx, "some-user")
	return ctx
}
