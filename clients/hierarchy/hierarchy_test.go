package hierarchy

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ONSdigital/dp-rchttp"
	. "github.com/smartystreets/goconvey/convey"
)

// client with no retries, no backoff
var client = &rchttp.Client{HTTPClient: &http.Client{}}
var ctx = context.Background()

type MockedHTTPResponse struct {
	StatusCode int
	Body       string
}

func getMockHierarchyAPI(expectRequest http.Request, mockedHTTPResponse MockedHTTPResponse) *Client {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != expectRequest.Method {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("unexpected HTTP method used"))
			return
		}
		w.WriteHeader(mockedHTTPResponse.StatusCode)
		fmt.Fprintln(w, mockedHTTPResponse.Body)
	}))
	return New(ts.URL)
}

func TestClient_GetRoot(t *testing.T) {
	instanceID := "foo"
	name := "bar"
	model := `{"label":"","has_data":true}`
	Convey("When bad request is returned", t, func() {
		mockedAPI := getMockHierarchyAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: 400, Body: ""})
		_, err := mockedAPI.GetRoot(ctx, instanceID, name)
		So(err, ShouldNotBeNil)
	})

	Convey("When server error is returned", t, func() {
		mockedAPI := getMockHierarchyAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: 500, Body: "qux"})
		_, err := mockedAPI.GetRoot(ctx, instanceID, name)
		So(err, ShouldNotBeNil)
	})
	Convey("When a filter-instance is returned", t, func() {
		mockedAPI := getMockHierarchyAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: 200, Body: model})
		_, err := mockedAPI.GetRoot(ctx, instanceID, name)
		So(err, ShouldBeNil)
	})

}

func TestClient_GetChild(t *testing.T) {
	instanceID := "foo"
	name := "bar"
	code := "baz"
	model := `{"label":"","has_data":true}`
	Convey("When bad request is returned", t, func() {
		mockedAPI := getMockHierarchyAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: 400, Body: ""})
		_, err := mockedAPI.GetChild(ctx, instanceID, name, code)
		So(err, ShouldNotBeNil)
	})

	Convey("When server error is returned", t, func() {
		mockedAPI := getMockHierarchyAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: 500, Body: "qux"})
		_, err := mockedAPI.GetChild(ctx, instanceID, name, code)
		So(err, ShouldNotBeNil)
	})

	Convey("When a filter-instance is returned", t, func() {
		mockedAPI := getMockHierarchyAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: 200, Body: model})
		_, err := mockedAPI.GetChild(ctx, instanceID, name, code)
		So(err, ShouldBeNil)
	})
}
func TestClient_GetHierarchy(t *testing.T) {
	path := "/hierarchies/foo/bar"
	model := `{"label":"","has_data":true}`

	Convey("When bad request is returned", t, func() {
		mockedAPI := getMockHierarchyAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: 400, Body: ""})
		_, err := mockedAPI.getHierarchy(path, ctx)
		So(err, ShouldNotBeNil)
	})

	Convey("When server error is returned", t, func() {
		mockedAPI := getMockHierarchyAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: 500, Body: "qux"})
		_, err := mockedAPI.getHierarchy(path, ctx)
		So(err, ShouldNotBeNil)
	})

	Convey("When a filter-instance is returned", t, func() {
		mockedAPI := getMockHierarchyAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: 200, Body: model})
		_, err := mockedAPI.getHierarchy(path, ctx)
		So(err, ShouldBeNil)
	})

}
