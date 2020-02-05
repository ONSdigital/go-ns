package requestID

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/headers"
	"github.com/ONSdigital/go-ns/common"
	. "github.com/smartystreets/goconvey/convey"
)

var dummyHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {})

func TestHandler(t *testing.T) {
	Convey("requestID handler should wrap another handler", t, func() {
		handler := Handler(20)
		wrapped := handler(dummyHandler)
		So(wrapped, ShouldHaveSameTypeAs, dummyHandler)
	})

	Convey("requestID should create a request ID if it doesn't exist", t, func() {
		req, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			t.Fail()
		}
		w := httptest.NewRecorder()

		requestID, err := headers.GetRequestID(req)
		So(err, ShouldResemble, headers.ErrHeaderNotFound)
		So(requestID, ShouldBeEmpty)

		handler := Handler(20)
		wrapped := handler(dummyHandler)

		wrapped.ServeHTTP(w, req)
		So(w.Code, ShouldEqual, 200)

		header, err := headers.GetRequestID(req)
		So(err, ShouldBeNil)
		So(header, ShouldNotBeEmpty)
		So(header, ShouldHaveLength, 20)
	})

	Convey("Existing request ID should be used if present", t, func() {
		req, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			t.Fail()
		}
		w := httptest.NewRecorder()
		expectedRequestID := "test"
		headers.SetRequestID(req, expectedRequestID)
		requestID, _ := headers.GetRequestID(req)
		So(requestID, ShouldEqual, expectedRequestID)

		handler := Handler(20)
		wrapped := handler(dummyHandler)

		wrapped.ServeHTTP(w, req)
		So(w.Code, ShouldEqual, 200)

		header, err := headers.GetRequestID(req)
		So(err, ShouldBeNil)
		So(header, ShouldNotBeEmpty)
		So(header, ShouldEqual, expectedRequestID)
	})

	Convey("Length of requestID should be configurable", t, func() {
		req, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			t.Fail()
		}
		w := httptest.NewRecorder()

		handler := Handler(30)
		wrapped := handler(dummyHandler)

		wrapped.ServeHTTP(w, req)
		So(w.Code, ShouldEqual, 200)

		header, err := headers.GetRequestID(req)
		So(err, ShouldBeNil)
		So(header, ShouldNotBeEmpty)
		So(header, ShouldHaveLength, 30)
	})

	Convey("generated requestIDs should be added to the request context", t, func() {
		req, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			t.Fail()
		}

		var reqCtx context.Context
		var captureContextHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			reqCtx = req.Context()
		})

		w := httptest.NewRecorder()

		handler := Handler(30)
		wrapped := handler(captureContextHandler)

		wrapped.ServeHTTP(w, req)
		So(w.Code, ShouldEqual, 200)
		So(reqCtx, ShouldNotBeNil)

		id := common.GetRequestId(reqCtx)
		So(id, ShouldNotBeEmpty)
		So(len(id), ShouldEqual, 30)
	})

	Convey("existing requestIDs should be added to the request context", t, func() {
		req, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			t.Fail()
		}
		headers.SetRequestID(req, "666")

		var reqCtx context.Context
		var captureContextHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			reqCtx = req.Context()
		})

		w := httptest.NewRecorder()

		handler := Handler(30)
		wrapped := handler(captureContextHandler)

		wrapped.ServeHTTP(w, req)
		So(w.Code, ShouldEqual, 200)
		So(reqCtx, ShouldNotBeNil)

		id := common.GetRequestId(reqCtx)
		So(id, ShouldNotBeEmpty)
		So(id, ShouldEqual, "666")
	})
}
