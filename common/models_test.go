package common

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCopy(t *testing.T) {

	Convey("Given 'Params' are copied", t, func() {
		p := Params{"name": "john", "surname": "smith"}

		copiedParams := p.Copy()
		So(copiedParams, ShouldNotEqual, p)
		So(copiedParams, ShouldResemble, p)

		Convey("When the original params are changed", func() {
			p["name"] = "dave"
			So(p, ShouldResemble, Params{"name": "dave", "surname": "smith"})

			Convey("Then copied params are unchanged", func() {
				So(copiedParams, ShouldNotResemble, p)
				So(copiedParams, ShouldResemble, Params{"name": "john", "surname": "smith"})
			})
		})
	})

	Convey("Given 'Params' are empty", t, func() {
		Convey("When copy func is called", func() {
			Convey("Then returned parameters is nil", func() {
				var noParams Params
				NoParams := noParams.Copy()
				So(NoParams, ShouldBeNil)
			})
		})
	})
}
