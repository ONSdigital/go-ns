package common

import (
	"context"
	"net/http"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestONSContext_NewONSContext(t *testing.T) {

	Convey("Given no pre-existing correlation-id", t, func() {

		Convey("when NewONSContext is called", func() {
			req := http.Request{}
			ctx := context.Background()

			Convey("then the ONSContext from an `http.Request` should have a correlation ID", func() {
				c := NewONSContext(req)
				So(GetContextID(c), ShouldNotBeBlank)
				// test that c is a `context`
				So(c.Done(), ShouldBeNil)
			})
			Convey("then the ONSContext from a `context` should have a correlation ID", func() {
				c := NewONSContext(ctx)
				So(GetContextID(c), ShouldNotBeBlank)
			})
			Convey("then the ONSContext from '' should have a correlation ID", func() {
				c := NewONSContext("")
				So(GetContextID(c), ShouldNotBeBlank)
			})
			Convey("then the ONSContext from an unexpected type should panic", func() {
				So(func() { NewONSContext(0) }, ShouldPanic)
			})

		})
	})

	Convey("Given a pre-existing correlation-id", t, func() {

		Convey("when NewONSContext is called", func() {
			hdr := http.Header{}
			hdr.Add("x-request-id", "ook")
			req := http.Request{Header: hdr}
			ctx := context.WithValue(context.Background(), "correlation_id", "koo")

			Convey("then the ONSContext from an `http.Request` should also have a correlation ID", func() {
				c := NewONSContext(req)
				So(GetContextID(c), ShouldEqual, "ook")
			})
			Convey("then the ONSContext from a `context` should have a correlation ID", func() {
				c := NewONSContext(ctx)
				So(GetContextID(c), ShouldEqual, "koo")
			})
			Convey("then the ONSContext from '' should have a correlation ID", func() {
				c := NewONSContext("foo")
				So(GetContextID(c), ShouldEqual, "foo")
				// test that c is a `context`
				So(c.Done(), ShouldBeNil)
			})

		})
	})
}
