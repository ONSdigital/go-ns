package server

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/facebookgo/freeport"
	. "github.com/smartystreets/goconvey/convey"
)

func TestNew(t *testing.T) {
	Convey("New should return a new server with sensible defaults", t, func() {
		h := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {})
		s := New(":0", h)

		So(s, ShouldNotBeNil)
		So(s.Handler, ShouldEqual, h)
		So(s.Alice, ShouldBeNil)
		So(s.Addr, ShouldEqual, ":0")
		So(s.MaxHeaderBytes, ShouldEqual, 0)

		Convey("TLS should not be configured by default", func() {
			So(s.CertFile, ShouldBeEmpty)
			So(s.KeyFile, ShouldBeEmpty)
		})

		Convey("Default middleware should include Timeout, RequestID and Log", func() {
			So(s.Middleware, ShouldContainKey, "Timeout")
			So(s.Middleware, ShouldContainKey, "RequestID")
			So(s.Middleware, ShouldContainKey, "Log")
			So(s.MiddlewareOrder, ShouldResemble, []string{"RequestID", "Log", "Timeout"})
		})

		Convey("Default timeouts should be sensible", func() {
			So(s.ReadTimeout, ShouldEqual, time.Second*5)
			So(s.WriteTimeout, ShouldEqual, time.Second*10)
			So(s.ReadHeaderTimeout, ShouldEqual, 0)
			So(s.IdleTimeout, ShouldEqual, 0)
		})
	})

	Convey("prep should prepare the server correctly", t, func() {
		Convey("prep should create a valid Server instance", func() {
			h := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {})
			s := New(":0", h)

			s.prep()
			So(s.Server.Addr, ShouldEqual, ":0")
		})

		Convey("invalid middleware should panic", func() {
			h := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {})
			s := New(":0", h)

			s.MiddlewareOrder = []string{"foo"}

			So(func() {
				s.prep()
			}, ShouldPanicWith, "middleware not found: foo")
		})

		Convey("ListenAndServe with invalid middleware should panic", func() {
			h := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {})
			s := New(":0", h)

			s.MiddlewareOrder = []string{"foo"}

			So(func() {
				s.ListenAndServe()
			}, ShouldPanicWith, "middleware not found: foo")
		})

		Convey("ListenAndServeTLS with invalid middleware should panic", func() {
			h := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {})
			s := New(":0", h)

			s.MiddlewareOrder = []string{"foo"}

			So(func() {
				s.ListenAndServeTLS("", "")
			}, ShouldPanicWith, "middleware not found: foo")
		})
	})

	Convey("ListenAndServeTLS", t, func() {
		Convey("ListenAndServeTLS should set CertFile/KeyFile", func() {
			h := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {})
			s := New(":0", h)

			go func() {
				s.ListenAndServeTLS("testdata/certFile", "testdata/keyFile")
			}()
			time.Sleep(time.Millisecond * 20)
			So(s.CertFile, ShouldEqual, "testdata/certFile")
			So(s.KeyFile, ShouldEqual, "testdata/keyFile")
		})

		Convey("ListenAndServeTLS with only CertFile should panic", func() {
			h := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {})
			s := New(":0", h)

			So(func() {
				s.ListenAndServeTLS("certFile", "")
			}, ShouldPanicWith, "either CertFile/KeyFile must be blank, or both provided")
		})

		Convey("ListenAndServeTLS with only KeyFile should panic", func() {
			h := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {})
			s := New(":0", h)

			So(func() {
				s.ListenAndServeTLS("", "keyFile")
			}, ShouldPanicWith, "either CertFile/KeyFile must be blank, or both provided")
		})
	})

	Convey("ListenAndServe starts a working HTTP server", t, func() {
		h := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {})

		port, err := freeport.Get()
		So(err, ShouldBeNil)
		So(port, ShouldBeGreaterThan, 0)

		sPort := fmt.Sprintf(":%d", port)
		s := New(sPort, h)

		go func() {
			s.ListenAndServe()
		}()
		time.Sleep(time.Millisecond * 20)

		res, err := http.Get("http://localhost" + sPort)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
		res.Body.Close()
		So(res.StatusCode, ShouldEqual, 200)
	})

}
