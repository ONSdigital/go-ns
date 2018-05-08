package audit

import (
	"context"
	"fmt"
	"github.com/ONSdigital/go-ns/common"
	"github.com/ONSdigital/go-ns/handlers/requestID"
	"github.com/ONSdigital/go-ns/log"
	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

const (
	service     = "audit-test"
	auditAction = "test"
	auditResult = "success"
	user        = "some-user"
)

func TestAuditor_RecordNoUserOrService(t *testing.T) {
	Convey("given no user or service identity exists in the provided context", t, func() {
		producer := &OutboundProducerMock{}

		auditor := New(producer, service)

		// record the audit event
		err := auditor.Record(context.Background(), auditAction, auditResult, nil)

		So(err, ShouldResemble, NewAuditError("expected user or caller identity but none found", auditAction, auditResult, nil))
		So(len(producer.OutputCalls()), ShouldEqual, 0)
	})
}

func TestAuditor_RecordNoUser(t *testing.T) {
	Convey("given no user identity exists in the provided context", t, func() {
		producer := &OutboundProducerMock{}

		auditor := New(producer, service)

		// record the audit event
		ctx := common.SetCaller(context.Background(), "Lucky The Donkey")
		err := auditor.Record(ctx, auditAction, auditResult, nil)

		So(err, ShouldBeNil)
		So(len(producer.OutputCalls()), ShouldEqual, 0)
	})
}

func TestAuditor_RecordAvroMarshalError(t *testing.T) {
	Convey("given there is an error converting the audit event to into avro", t, func() {
		producer := &OutboundProducerMock{}
		auditor := New(producer, service)

		auditor.marshalToAvro = func(s interface{}) ([]byte, error) {
			return nil, errors.New("avro marshal error")
		}

		// record the audit event
		err := auditor.Record(setUpContext(), auditAction, auditResult, nil)

		expectedErr := NewAuditError("error marshalling event to arvo", auditAction, auditResult, nil)
		So(err, ShouldResemble, expectedErr)
		So(len(producer.OutputCalls()), ShouldEqual, 0)
	})
}

func TestAuditor_RecordSuccess(t *testing.T) {
	Convey("given valid parameters are provided", t, func() {
		output := make(chan []byte, 1)

		producer := &OutboundProducerMock{
			OutputFunc: func() chan []byte {
				return output
			},
		}

		auditor := New(producer, service)

		var results []byte

		// record the audit event
		err := auditor.Record(setUpContext(), auditAction, auditResult, common.Params{"ID": "12345"})
		So(err, ShouldBeNil)

		select {
		case results = <-output:
			log.Info("output received", nil)
		case <-time.After(time.Second * 5):
			log.Debug("failing test due to timeout, expected output channel to receive event but none", nil)
			t.FailNow()
		}

		So(len(producer.OutputCalls()), ShouldEqual, 1)

		var actualEvent Event
		err = EventSchema.Unmarshal(results, &actualEvent)
		if err != nil {
			log.ErrorC("avro unmarshal error", err, nil)
			t.FailNow()
		}

		So(actualEvent.RequestID, ShouldBeEmpty)
		So(actualEvent.Service, ShouldEqual, service)
		So(actualEvent.AttemptedAction, ShouldEqual, auditAction)
		So(actualEvent.ActionResult, ShouldEqual, auditResult)
		So(actualEvent.Created, ShouldNotBeEmpty)
		So(actualEvent.User, ShouldEqual, user)
		So(actualEvent.Params, ShouldResemble, common.Params{"ID": "12345"})
	})
}

func TestAuditor_RecordRequestIDInContext(t *testing.T) {
	Convey("given the context contain a requestID", t, func() {
		output := make(chan []byte, 1)

		producer := &OutboundProducerMock{
			OutputFunc: func() chan []byte {
				return output
			},
		}

		auditor := New(producer, service)

		var results []byte

		// record the audit event
		ctx := context.WithValue(setUpContext(), requestID.ContextKey, "666")
		err := auditor.Record(ctx, auditAction, auditResult, common.Params{"ID": "12345"})
		So(err, ShouldBeNil)

		select {
		case results = <-output:
			log.Info("output received", nil)
		case <-time.After(time.Second * 5):
			log.Debug("failing test due to timeout, expected output channel to receive event but none", nil)
			t.FailNow()
		}

		So(len(producer.OutputCalls()), ShouldEqual, 1)

		var actualEvent Event
		err = EventSchema.Unmarshal(results, &actualEvent)
		if err != nil {
			log.ErrorC("avro unmarshal error", err, nil)
			t.FailNow()
		}

		So(actualEvent.RequestID, ShouldEqual, "666")
		So(actualEvent.Service, ShouldEqual, service)
		So(actualEvent.AttemptedAction, ShouldEqual, auditAction)
		So(actualEvent.ActionResult, ShouldEqual, auditResult)
		So(actualEvent.Created, ShouldNotBeEmpty)
		So(actualEvent.User, ShouldEqual, user)
		So(actualEvent.Params, ShouldResemble, common.Params{"ID": "12345"})
	})
}

func TestAuditor_RecordEmptyAction(t *testing.T) {
	Convey("given Record is called with an empty action value then the expected error is returned", t, func() {
		producer := &OutboundProducerMock{}

		auditor := New(producer, service)

		err := auditor.Record(setUpContext(), "", "", nil)

		So(len(producer.OutputCalls()), ShouldEqual, 0)
		expectedErr := NewAuditError("attemptedAction required but was empty", "", "", nil)
		So(err, ShouldResemble, expectedErr)
	})
}

func TestAuditor_RecordEmptyResult(t *testing.T) {
	Convey("given Record is called with an empty result value then the expected error is returned", t, func() {
		producer := &OutboundProducerMock{}

		auditor := New(producer, service)

		err := auditor.Record(setUpContext(), auditAction, "", nil)

		So(len(producer.OutputCalls()), ShouldEqual, 0)
		expectedErr := NewAuditError("actionResult required but was empty", "test", "", nil)
		So(err, ShouldResemble, expectedErr)
	})
}

func Test_newAuditError(t *testing.T) {
	Convey("given no values are provided", t, func() {
		actual := NewAuditError("", "", "", nil)

		Convey("then an error with default values is returned", func() {
			expected := Error{
				Cause:  "",
				Action: "",
				Result: "",
				Params: nil,
			}

			So(actual, ShouldResemble, expected)
		})

		Convey("and Error() returns the expected value", func() {
			fmt.Println(actual.Error())
			So(actual.Error(), ShouldEqual, "unable to audit event, attempted action: , action result: , cause: , params: []")
		})
	})

	Convey("given valid values for all fields", t, func() {
		actual := NewAuditError("_cause", "_action", "_result", common.Params{
			"bbb": "bbb",
			"aaa": "aaa",
			"ccc": "ccc",
		})

		expected := Error{
			Cause:  "_cause",
			Action: "_action",
			Result: "_result",
			Params: common.Params{
				"aaa": "aaa",
				"bbb": "bbb",
				"ccc": "ccc",
			},
		}

		So(actual.Cause, ShouldEqual, expected.Cause)
		So(actual.Action, ShouldEqual, expected.Action)
		So(actual.Result, ShouldEqual, expected.Result)

		expectedStr := "unable to audit event, attempted action: _action, action result: _result, cause: _cause, params: [aaa:aaa, bbb:bbb, ccc:ccc]"
		So(actual.Error(), ShouldEqual, expectedStr)
	})
}

func setUpContext() context.Context {
	ctx := common.SetCaller(context.Background(), service)
	ctx = common.SetUser(ctx, user)
	return ctx
}
