package healthcheck

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitHealthcheck(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	Convey("test Do returns status 500 when external service error found", t, func() {
		mcli := NewMockClient(mockCtrl)
		mcli.EXPECT().Healthcheck().Return(errors.New("bad healthcheck - sad face"))
		mcli.EXPECT().Name().Return("my-external-service")

		MonitorExternal(mcli)

		req := httptest.NewRequest("GET", "/healthcheck", nil)
		w := httptest.NewRecorder()

		output := captureOutput(func() {
			Do(w, req)
		})

		So(w.Code, ShouldEqual, http.StatusInternalServerError)
		So(output, ShouldContainSubstring, "unsuccessful healthcheck for my-external-service: bad healthcheck - sad face")
	})

	Convey("test Do returns status 200 when external service error found", t, func() {
		mcli := NewMockClient(mockCtrl)
		mcli.EXPECT().Healthcheck().Return(nil)

		MonitorExternal(mcli)

		req := httptest.NewRequest("GET", "/healthcheck", nil)
		w := httptest.NewRecorder()

		Do(w, req)
		So(w.Code, ShouldEqual, http.StatusOK)
	})
}

func captureOutput(f func()) string {
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	outC := make(chan string)

	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	w.Close()
	os.Stdout = stdout
	out := <-outC
	return out
}
