package healthcheck

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ONSdigital/go-ns/healthcheck/mock_healthcheck"
	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitHealthcheck(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	Convey("test Do returns status 500 when external service error found", t, func() {
		mcli := mock_healthcheck.NewMockClient(mockCtrl)
		mcli.EXPECT().Healthcheck().Return("my-external-service", errors.New("bad healthcheck - sad face"))

		MonitorExternal(mcli)

		req := httptest.NewRequest("GET", "/healthcheck", nil)
		w := httptest.NewRecorder()
		Do(w, req)

		output := w.Body.String()
		So(w.Code, ShouldEqual, http.StatusInternalServerError)
		So(output, ShouldStartWith, "{")
		So(output, ShouldContainSubstring, "unsuccessful healthcheck for my-external-service: bad healthcheck - sad face")
	})

	Convey("test Do returns status 200 when external service error found", t, func() {
		mcli := mock_healthcheck.NewMockClient(mockCtrl)
		mcli.EXPECT().Healthcheck().Return("", nil)

		MonitorExternal(mcli)

		req := httptest.NewRequest("GET", "/healthcheck", nil)
		w := httptest.NewRecorder()
		Do(w, req)

		output := w.Body.String()
		So(w.Code, ShouldEqual, http.StatusOK)
		So(output, ShouldStartWith, "{")
		So(output, ShouldContainSubstring, `"status":"OK"`)
	})
}
