package timeout

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

var dummyHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
	time.Sleep(2 * time.Second)
})

func TestHandler(t *testing.T) {
	Convey("timeout handler should wrap another handler", t, func() {
		handler := Handler(1 * time.Second)
		wrapped := handler(dummyHandler)
		So(wrapped, ShouldHaveSameTypeAs, http.TimeoutHandler(dummyHandler, 1*time.Second, "timed out"))
	})

	Convey("timeout handler should time out if response takes too long", t, func() {
		req, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			t.Fail()
		}
		w := httptest.NewRecorder()

		So(req.Header.Get("X-Request-Id"), ShouldBeEmpty)

		handler := Handler(20)
		wrapped := handler(dummyHandler)

		wrapped.ServeHTTP(w, req)
		So(w.Code, ShouldEqual, 503)
	})

}
