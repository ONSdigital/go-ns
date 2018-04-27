package audit

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/ONSdigital/go-ns/common"
	"github.com/ONSdigital/go-ns/handlers/requestID"
	"github.com/ONSdigital/go-ns/identity"
	"github.com/ONSdigital/go-ns/log"
	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	namespace   = "audit-test"
	auditAction = "test"
	auditResult = "success"
	user        = "some-user"
	service     = "some-service"
)

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
	Convey("given valid parameters are provided", t, func() {
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

		So(actualEvent.RequestID, ShouldBeEmpty)
		So(actualEvent.Namespace, ShouldEqual, namespace)
		So(actualEvent.AttemptedAction, ShouldEqual, auditAction)
		So(actualEvent.Result, ShouldEqual, auditResult)
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

		auditor := New(producer, namespace)

		var results []byte

		// record the audit event
		ctx := context.WithValue(setUpContext(), requestID.ContextKey, "666")
		err := auditor.Record(ctx, auditAction, auditResult, common.Params{"ID": "12345"})

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

		So(actualEvent.RequestID, ShouldEqual, "666")
		So(actualEvent.Namespace, ShouldEqual, namespace)
		So(actualEvent.AttemptedAction, ShouldEqual, auditAction)
		So(actualEvent.Result, ShouldEqual, auditResult)
		So(actualEvent.Created, ShouldNotBeEmpty)
		So(actualEvent.User, ShouldEqual, user)
		So(actualEvent.Params, ShouldResemble, common.Params{"ID": "12345"})
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

func Test_newAuditError(t *testing.T) {
	Convey("given no values are provided", t, func() {
		actual := newAuditError("", "", "", nil)

		Convey("then an error with default values is returned", func() {
			expected := Error{
				Cause:  "nil",
				Action: "nil",
				Result: "nil",
				Params: common.Params{},
			}

			So(actual, ShouldResemble, expected)
		})

		Convey("and Error() returns the expected value", func() {
			fmt.Println(actual.Error())
			So(actual.Error(), ShouldEqual, "unable to audit event, action: nil, result: nil, cause: nil, params: map[]")
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
			Params: common.Params{
				"bbb": "bbb",
				"aaa": "aaa",
				"ccc": "ccc",
			},
		}

		So(actual.Cause, ShouldEqual, expected.Cause)
		So(actual.Action, ShouldEqual, expected.Action)
		So(actual.Result, ShouldEqual, expected.Result)

		// verify that the parameters are as expected
		So(actual.Params, ShouldResemble, expected.Params)

		expectedStr := "unable to audit event, action: _action, result: _result, cause: _cause, params: map["
		So(actual.Error(), ShouldStartWith, expectedStr)

		actualErrorAsMap, err := extractMapFromString(actual.Error())
		So(err, ShouldBeNil)
		So(len(actualErrorAsMap), ShouldEqual, len(expected.Params))
		So(actualErrorAsMap, ShouldResemble, expected.Params)
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

func extractMapFromString(str string) (params common.Params, err error) {
	var indexPos int
	if indexPos = strings.Index(str, "map["); indexPos == -1 {
		return params, errors.New("No map[] start")
	}
	str = str[indexPos+4:] // str is now: `aaa:aaa bbb:bbb ccc:ccc]` - now remove the trailing `]`
	if indexPos = strings.LastIndex(str, "]"); indexPos == -1 {
		return params, errors.New("No map[] end")
	}
	params = make(common.Params)
	for _, keyValString := range strings.Split(str[:indexPos], " ") {
		// keyValString is: `aaa:aaa` - now split into (key,val) and put into params
		keyValArray := strings.SplitN(keyValString, ":", 2)
		params[keyValArray[0]] = keyValArray[1]
	}
	return
}
