package collectionID


import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ONSdigital/go-ns/common"
	. "github.com/smartystreets/goconvey/convey"
)

var dummyHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {})

func TestHandler(t *testing.T) {

	Convey("Existing collection ID should be used if present", t, func() {
		req, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			t.Fail()
		}
		w := httptest.NewRecorder()

		req.Header.Set(common.CollectionIDHeaderKey, "test")
		So(req.Header.Get(common.CollectionIDHeaderKey), ShouldNotBeEmpty)

		handler := dummyHandler

		handler.ServeHTTP(w, req)
		So(w.Code, ShouldEqual, 200)

		header := req.Header.Get(common.CollectionIDHeaderKey)
		So(header, ShouldNotBeEmpty)
		So(header, ShouldHaveLength, 4)
		So(header, ShouldEqual, "test")
	})

}
