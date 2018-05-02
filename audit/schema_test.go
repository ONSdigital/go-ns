package audit

import (
	"github.com/ONSdigital/go-ns/common"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

var (
	testService     = "_service"
	reqID           = "_reqID"
	params          = common.Params{"key": "value"}
	actionResult    = "_ActionResult"
	attemptedAction = "_AttemptedAction"
	testUser        = "bob"
	created         = "now"
)

func TestEventSchema(t *testing.T) {
	Convey("given a valid audit event", t, func() {

		auditEvent := Event{
			Service:         testService,
			RequestID:       reqID,
			Params:          params,
			AttemptedAction: attemptedAction,
			ActionResult:    actionResult,
			User:            testUser,
			Created:         created,
		}

		var b []byte
		var err error

		Convey("when marshall is called no error is returned", func() {
			b, err = EventSchema.Marshal(auditEvent)
			So(err, ShouldBeNil)
			So(len(b), ShouldBeGreaterThan, 0)

			Convey("when unmarshal is called no error is returned", func() {
				var actual Event
				err = EventSchema.Unmarshal(b, &actual)
				So(err, ShouldBeNil)

				Convey("and the unmarshalled value is as expected", func() {
					So(actual, ShouldResemble, auditEvent)
				})
			})
		})
	})
}
