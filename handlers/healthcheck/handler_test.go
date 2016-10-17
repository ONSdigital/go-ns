package healthcheck

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestHandler(t *testing.T) {
	Convey("Healthcheck should return a 200 response", t, func() {
		req, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			t.Fail()
		}
		w := httptest.NewRecorder()

		Handler(w, req)
		So(w.Code, ShouldEqual, 200)
	})
}
