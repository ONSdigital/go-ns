package filter

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ONSdigital/dp-rchttp"
	. "github.com/smartystreets/goconvey/convey"
)

const serviceToken = "bar"
const downloadServiceToken = "baz"

// client with no retries, no backoff
var client = &rchttp.Client{HTTPClient: &http.Client{}}
var ctx = context.Background()

type MockedHTTPResponse struct {
	StatusCode int
	Body       string
}

func getMockfilterAPI(expectRequest http.Request, mockedHTTPResponse MockedHTTPResponse) *Client {
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


func TestGetOutput(t *testing.T) {
	filterOutputID := "foo"
	filterOutputBody := `{"id":"` + filterOutputID + `","quux":"quuz"}`
	Convey("When bad request is returned", t, func() {
		mockedAPI := getMockfilterAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: 400, Body: ""})
		_, err := mockedAPI.GetOutput(ctx, filterOutputID, serviceToken, downloadServiceToken)
		So(err, ShouldNotBeNil)
	})

	Convey("When server error is returned", t, func() {
		mockedAPI := getMockfilterAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: 500, Body: "qux"})
		_, err := mockedAPI.GetOutput(ctx, filterOutputID, serviceToken, downloadServiceToken)
		So(err, ShouldNotBeNil)
	})

	Convey("When a filter-instance is returned", t, func() {
		mockedAPI := getMockfilterAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: 200, Body: filterOutputBody})
		_, err := mockedAPI.GetOutput(ctx, filterOutputID, serviceToken, downloadServiceToken)
		So(err, ShouldBeNil)
	})
}

func TestGetDimension(t *testing.T) {
	filterOutputID := "foo"
	name := "corge"
	dimensionBody := `{"id":"` + filterOutputID + `","quux":"quuz"}`
	Convey("When bad request is returned", t, func() {
		mockedAPI := getMockfilterAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: 400, Body: ""})
		_, err := mockedAPI.GetDimension(ctx, filterOutputID, name, serviceToken, downloadServiceToken)
		So(err, ShouldNotBeNil)
	})

	Convey("When server error is returned", t, func() {
		mockedAPI := getMockfilterAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: 500, Body: "qux"})
		_, err := mockedAPI.GetDimension(ctx, filterOutputID, name, serviceToken, downloadServiceToken)
		So(err, ShouldNotBeNil)
	})

	Convey("When a dimension-instance is returned", t, func() {
		mockedAPI := getMockfilterAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: 200, Body: dimensionBody})
		_, err := mockedAPI.GetDimension(ctx, filterOutputID, name, serviceToken, downloadServiceToken)
		So(err, ShouldBeNil)
	})
}
