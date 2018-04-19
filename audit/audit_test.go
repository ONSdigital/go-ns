package audit

import (
	"context"
	"encoding/json"
	"github.com/ONSdigital/go-ns/audit/audit_test"
	"github.com/ONSdigital/go-ns/identity"
	"github.com/ONSdigital/go-ns/log"
	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

func TestAuditor_Record(t *testing.T) {
	Convey("given no audit event exists in the provided context", t, func() {
		producer := &audit_test.OutboundProducerMock{}

		auditor := New("test", producer, "test")

		// record the audit event
		err := auditor.Record(context.Background(), "test", "success", nil)

		So(err.Error(), ShouldEqual, noEventInContextError.Error())
		So(len(producer.OutputCalls()), ShouldEqual, 0)
	})

	Convey("given there is an error converting the audit event to into avro", t, func() {
		producer := &audit_test.OutboundProducerMock{}
		auditor := New("test", producer, "test")

		auditor.marshalToAvro = func(s interface{}) ([]byte, error) {
			return nil, errors.New("avro marshal error")
		}

		// record the audit event
		err := auditor.Record(setUpContext(t), "test", "success", nil)

		So(err.Error(), ShouldEqual, "avro marshal error")
		So(len(producer.OutputCalls()), ShouldEqual, 0)
	})

	Convey("given there is an error unmarshalling the event from context", t, func() {
		producer := &audit_test.OutboundProducerMock{}

		auditor := New("test", producer, "test")
		auditor.unmarshalJSON = func(data []byte, v interface{}) error {
			return errors.New("cannot unmarshal audit event")
		}

		// record the audit event
		err := auditor.Record(setUpContext(t), "test", "success", nil)

		So(err.Error(), ShouldEqual, "cannot unmarshal audit event")
		So(len(producer.OutputCalls()), ShouldEqual, 0)
	})

	Convey("given a valid audit event exists in context", t, func() {
		output := make(chan []byte, 1)

		producer := &audit_test.OutboundProducerMock{
			OutputFunc: func() chan []byte {
				return output
			},
		}

		auditor := New("test", producer, "test")
		auditor.unmarshalJSON = json.Unmarshal

		var results []byte

		// record the audit event
		err := auditor.Record(setUpContext(t), "test", "success", Params{"ID": "12345"})

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
}

func setUpContext(t *testing.T) context.Context {
	e := Event{Namespace: "audit-test"}
	b, err := json.Marshal(e)
	if err != nil {
		log.ErrorC("failing test due to json marshal error", err, nil)
		t.FailNow()
	}
	ctx := context.WithValue(context.Background(), contextKey("audit"), string(b))
	ctx = identity.SetCaller(ctx, "some-service")
	ctx = identity.SetUser(ctx, "some-user")
	return ctx
}
