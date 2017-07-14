package validator

import (
	"bytes"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

type Form struct {
	Search string `schema:"search"`
	Filter string
}

type Food struct {
	Pizza string `schema:"pizza"`
}

func TestUnitValidate(t *testing.T) {
	Convey("test validator sucessfully validates test data", t, func() {
		fv, err := New("testdata/rules.json")
		So(err, ShouldBeNil)

		data := url.Values{}
		data.Set("search", "matt@gmail.com")

		req, err := http.NewRequest("POST", "", bytes.NewBufferString(data.Encode()))
		So(err, ShouldBeNil)

		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

		frm := Form{}
		err = fv.Validate(req, &frm)
		So(err, ShouldBeNil)
	})

	Convey("test validator sucessfully validates a custom rule", t, func() {
		fv, err := New("testdata/custom_rule.json")
		So(err, ShouldBeNil)

		data := url.Values{}
		data.Set("pizza", "food")

		RulesList["food"] = func(vars ...interface{}) error {
			var f string
			var ok bool

			if f, ok = vars[0].(string); !ok {
				return errors.New("first parameter to food must be a string")
			}

			if f != "food" {
				return FieldValidationErr{errors.New("value must equal food")}
			}
			return nil
		}

		req, err := http.NewRequest("POST", "", bytes.NewBufferString(data.Encode()))
		So(err, ShouldBeNil)

		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

		food := Food{}
		err = fv.Validate(req, &food)
		So(err, ShouldBeNil)
	})

	Convey("test validator returns error if request missing form data", t, func() {
		fv, err := New("testdata/rules.json")
		So(err, ShouldBeNil)

		req, err := http.NewRequest("POST", "", nil)
		So(err, ShouldBeNil)

		err = fv.Validate(req, nil)
		So(err, ShouldNotBeNil)
	})

	Convey("test validator returns error if file does not contain valid json", t, func() {
		fv, err := New("testdata/junk.txt")
		So(err, ShouldBeNil)

		data := url.Values{}
		data.Set("search", "matt@gmail.com")

		req, err := http.NewRequest("POST", "", bytes.NewBufferString(data.Encode()))
		So(err, ShouldBeNil)

		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

		frm := Form{}
		err = fv.Validate(req, &frm)
		So(err, ShouldNotBeNil)
	})

	Convey("test validator returns error if json file rule has no corresponding func", t, func() {
		fv, err := New("testdata/rules-missing-func.json")
		So(err, ShouldBeNil)

		data := url.Values{}
		data.Set("search", "matt@gmail.com")

		req, err := http.NewRequest("POST", "", bytes.NewBufferString(data.Encode()))
		So(err, ShouldBeNil)

		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

		frm := Form{}
		err = fv.Validate(req, &frm)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "rule name: can_haz_cheeseburger, missing corresponding validation function")
	})

	Convey("test validator returns error if invalid parameter sent to min_length", t, func() {
		fv, err := New("testdata/invalid-min-length.json")
		So(err, ShouldBeNil)

		data := url.Values{}
		data.Set("search", "matt@gmail.com")

		req, err := http.NewRequest("POST", "", bytes.NewBufferString(data.Encode()))
		So(err, ShouldBeNil)

		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

		frm := Form{}
		err = fv.Validate(req, &frm)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "second param to minLength must be number")
	})

	Convey("test validator returns ErrFormValidationFailed if form validation falise", t, func() {
		fv, err := New("testdata/rules.json")
		So(err, ShouldBeNil)

		data := url.Values{}
		data.Set("search", "no")

		req, err := http.NewRequest("POST", "", bytes.NewBufferString(data.Encode()))
		So(err, ShouldBeNil)

		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

		frm := Form{}
		err = fv.Validate(req, &frm)
		So(err, ShouldEqual, ErrFormValidationFailed)

		fieldErrs := fv.GetFieldErrors()
		So(fieldErrs["search"][0].Error(), ShouldEqual, "value: no, must be at least 5 characters")
		So(fieldErrs["search"][1].Error(), ShouldEqual, "email: no is not a valid email address")
	})
}
