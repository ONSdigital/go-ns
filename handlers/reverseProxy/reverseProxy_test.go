package reverseProxy

import (
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"net/http/httputil"
	"net/url"
	"testing"
)

func TestDirectorFunc(t *testing.T) {
	proxyURL, _ := url.Parse("https://www.ons.gov.uk")
	Convey("Create proxy", t, func() {
		reverseProxy := Create(proxyURL, nil)

		So(reverseProxy, ShouldNotBeNil)
		So(reverseProxy, ShouldImplement, (*http.Handler)(nil))

		req, _ := http.NewRequest(`GET`, `https://cy.ons.gov.uk`, nil)
		So(func() { reverseProxy.(*httputil.ReverseProxy).Director(req) }, ShouldNotPanic)
		So(req.URL.Host, ShouldEqual, `www.ons.gov.uk`)
	})

	Convey("Create proxy with director func", t, func() {
		var directorCalled bool
		reverseProxy := Create(proxyURL, func(req *http.Request) {
			directorCalled = true
			req.URL.Host = `host`
		})

		So(reverseProxy, ShouldNotBeNil)
		So(reverseProxy, ShouldImplement, (*http.Handler)(nil))

		req, _ := http.NewRequest(`GET`, `https://cy.ons.gov.uk`, nil)
		So(func() { reverseProxy.(*httputil.ReverseProxy).Director(req) }, ShouldNotPanic)
		So(req.URL.Host, ShouldEqual, `host`)

		So(directorCalled, ShouldBeTrue)
	})
}
