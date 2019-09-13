package server

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/facebookgo/freeport"
	. "github.com/smartystreets/goconvey/convey"
)

func newWithPort(h http.Handler) (string, *Server) {
	port, err := freeport.Get()
	So(err, ShouldBeNil)
	So(port, ShouldBeGreaterThan, 0)

	sPort := fmt.Sprintf(":%d", port)

	return sPort, New(sPort, h)
}

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

		Convey("Default middleware should include RequestID and Log", func() {
			So(s.Middleware, ShouldContainKey, RequestIDHandlerKey)
			So(s.Middleware, ShouldContainKey, LogHandlerKey)
			So(s.MiddlewareOrder, ShouldResemble, []string{RequestIDHandlerKey, LogHandlerKey})
		})

		Convey("Default timeouts should be sensible", func() {
			So(s.ReadTimeout, ShouldEqual, time.Second*5)
			So(s.WriteTimeout, ShouldEqual, time.Second*10)
			So(s.ReadHeaderTimeout, ShouldEqual, 0)
			So(s.IdleTimeout, ShouldEqual, 0)
		})

		Convey("Handle OS signals by default", func() {
			So(s.HandleOSSignals, ShouldEqual, true)
		})

		Convey("A default shutdown context is initialised", func() {
			So(s.DefaultShutdownTimeout, ShouldEqual, 10*time.Second)
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
				s.ListenAndServeTLS("testdata/certFile", "testdata/keyFile")
			}, ShouldPanicWith, "middleware not found: foo")
		})
	})

	Convey("ListenAndServeTLS", t, func() {
		Convey("ListenAndServeTLS should set CertFile/KeyFile", func() {
			h := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {})

			sPort, s := newWithPort(h)
			go func() {
				s.ListenAndServeTLS("testdata/certFile", "testdata/keyFile")
			}()
			http.Get("http://localhost" + sPort) // ensure above is responding before we check below
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

		sPort, s := newWithPort(h)
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
