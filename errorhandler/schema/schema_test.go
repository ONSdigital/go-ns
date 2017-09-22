package errorschema

import (
	"testing"

	"github.com/ONSdigital/dp-import-reporter/handler"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSchemaSuccessful(t *testing.T) {
	Convey("Marshall and then unmarshall a event", t, func() {
		e := handler.EventReport{
			InstanceID: "a4695fee-f0a2-49c4-b136-e3ca8dd40476",
			EventType:  "error",
			EventMsg:   "Broken on something.",
		}
		avroBytes, aErr := ReportedEventSchema.Marshal(e)
		So(aErr, ShouldBeNil)
		var event handler.EventReport
		err := ReportedEventSchema.Unmarshal(avroBytes, &event)
		So(err, ShouldBeNil)
		So(event.InstanceID, ShouldEqual, e.InstanceID)
		ReportedEventSchema.Marshal(&e)
	})
}
