package request_test

import (
	"bytes"
	"github.com/ONSdigital/go-ns/request"
	. "github.com/smartystreets/goconvey/convey"
	"io"
	"net/http"
	"testing"
)

func TestDrainBody_WithRequestBody(t *testing.T) {

	Convey("Given a request with a body", t, func() {

		body := bytes.NewBufferString("some body content")
		r, err := http.NewRequest("GET", "/some/url", body)
		So(err, ShouldBeNil)

		Convey("When the DrainBody function is called", func() {

			request.DrainBody(r)

			Convey("The all bytes have been read from the body", func() {
				_, err = r.Body.Read(make([]byte, 1))
				So(err, ShouldEqual, io.EOF)
			})
		})
	})
}

func TestDrainBody_WithoutRequestBody(t *testing.T) {

	Convey("Given a request without a nil body", t, func() {

		r, err := http.NewRequest("GET", "/some/url", nil)
		So(err, ShouldBeNil)

		Convey("When the DrainBody function is called, there is no panic", func() {

			request.DrainBody(r)
		})
	})
}
