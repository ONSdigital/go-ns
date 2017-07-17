package validator

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitRules(t *testing.T) {
	Convey("test minLength", t, func() {
		Convey("test minLength successfully validates string length", func() {
			err := minLength("hello", 3.0)
			So(err, ShouldBeNil)
		})

		Convey("test minLength throws error if first param not string", func() {
			err := minLength(12, 15.0)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "first param to minLength must be string")
		})

		Convey("test minLength throws error if second param not number", func() {
			err := minLength("hello", "world")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "second param to minLength must be number")
		})

		Convey("test minLength throws FieldValidationErr if condition not met", func() {
			err := minLength("hello", 6.0)
			So(err, ShouldNotBeNil)
			So(err, ShouldHaveSameTypeAs, FieldValidationErr{})
			So(err.Error(), ShouldEqual, "value: hello, must be at least 6 characters")
		})
	})

	Convey("test maxLength", t, func() {
		Convey("test maxLength successfully validates string length", func() {
			err := maxLength("hello", 10.0)
			So(err, ShouldBeNil)
		})

		Convey("test maxLength throws error if first param not string", func() {
			err := maxLength(12, 15.0)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "first param to maxLength must be string")
		})

		Convey("test maxLength throws error if second param not number", func() {
			err := maxLength("hello", "world")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "second param to maxLength must be number")
		})

		Convey("test maxLength throws FieldValidationErr if condition not met", func() {
			err := maxLength("hello", 4.0)
			So(err, ShouldNotBeNil)
			So(err, ShouldHaveSameTypeAs, FieldValidationErr{})
			So(err.Error(), ShouldEqual, "value: hello, must be at most 4 characters")
		})
	})

	Convey("test email", t, func() {
		Convey("test email successfully validates email string", func() {
			err := email("matt@gmail.com")
			So(err, ShouldBeNil)
		})

		Convey("test email throws error if parameter not string", func() {
			err := email(2)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "first parameter to email must be a string")
		})

		Convey("test email throws FieldValidationErr if condition not met", func() {
			err := email("helloworld")
			So(err, ShouldNotBeNil)
			So(err, ShouldHaveSameTypeAs, FieldValidationErr{})
			So(err.Error(), ShouldEqual, "email: helloworld is not a valid email address")
		})
	})

	Convey("test notEmpty", t, func() {
		Convey("test notEmpty successfully validates slice", func() {
			err := notEmpty([]string{"hello"})
			So(err, ShouldBeNil)
		})

		Convey("test notEmpty throws error if first parameter is not a slice", func() {
			err := notEmpty("hello")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "first param to notEmpty must be a slice or array")
		})

		Convey("test notEmpty throws FieldValidationErr if condition not met", func() {
			err := notEmpty([]string{})
			So(err, ShouldNotBeNil)
			So(err, ShouldHaveSameTypeAs, FieldValidationErr{})
			So(err.Error(), ShouldEqual, "slice must not be empty")
		})
	})

}
