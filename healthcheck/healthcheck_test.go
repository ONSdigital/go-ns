package healthcheck

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ONSdigital/go-ns/healthcheck/mock_healthcheck"
	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitHealthcheck(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	Convey("given no healthcheck has yet been run, when endpoint is called", t, func() {

		req := httptest.NewRequest("GET", "/healthcheck", nil)
		w := httptest.NewRecorder()
		Do(w, req)

		Convey("then http status is 429", func() {

			So(w.Code, ShouldEqual, http.StatusTooManyRequests)
		})

		Convey("then GetState returns zero-valued lastTry time", func() {

			state, lastTry, _ := GetState()
			So(len(state), ShouldEqual, 0)
			So(lastTry.IsZero(), ShouldBeTrue)
		})
	})

	Convey("given external service errors, when MonitorExternal is called", t, func() {
		mcli := mock_healthcheck.NewMockClient(mockCtrl)
		mcli.EXPECT().Healthcheck().Return("my-external-service", errors.New("bad healthcheck - sad face"))

		startTime := time.Now()
		MonitorExternal(mcli)

		Convey("then Do returns status 500", func() {

			req := httptest.NewRequest("GET", "/healthcheck", nil)
			w := httptest.NewRecorder()
			Do(w, req)

			output := w.Body.String()
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
			So(output, ShouldStartWith, "{")
			So(output, ShouldContainSubstring, "bad healthcheck - sad face")
		})

		Convey("then GetState returns populated error map", func() {

			state, lastTry, _ := GetState()
			So(state, ShouldResemble, HealthState{"my-external-service": errors.New("bad healthcheck - sad face")})
			So(lastTry, ShouldHappenOnOrAfter, startTime)
		})
	})

	Convey("given no external service errors, when MonitorExternal is called", t, func() {
		mcli := mock_healthcheck.NewMockClient(mockCtrl)
		mcli.EXPECT().Healthcheck().Return("", nil)

		startTime := time.Now()
		MonitorExternal(mcli)

		Convey("then Do returns status 200", func() {

			req := httptest.NewRequest("GET", "/healthcheck", nil)
			w := httptest.NewRecorder()
			Do(w, req)

			output := w.Body.String()
			So(w.Code, ShouldEqual, http.StatusOK)
			So(output, ShouldStartWith, "{")
			So(output, ShouldContainSubstring, `"status":"OK"`)
		})

		Convey("then GetState returns empty error map", func() {

			state, lastTry, lastSuccess := GetState()
			So(len(state), ShouldEqual, 0)
			So(lastTry, ShouldHappenOnOrAfter, startTime)
			So(lastSuccess, ShouldHappenOnOrAfter, startTime)
		})

	})
}
