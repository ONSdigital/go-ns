package importapi

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

func getMockImportAPI(expectRequest http.Request, mockedHTTPResponse MockedHTTPResponse) *Client {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != expectRequest.Method {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("unexpected HTTP method used"))
			return
		}
		w.WriteHeader(mockedHTTPResponse.StatusCode)
		fmt.Fprintln(w, mockedHTTPResponse.Body)
	}))
	return NewAPIClient(client, ts.URL)
}

func TestGetImportJob(t *testing.T) {
	jobID := "jid1"
	jobJSON := `{"id":"` + jobID + `","links":{"instances":[{"id":"iid1","href":"iid1link"}]}}`
	jobMultiInstJSON := `{"id":"` + jobID + `","links":{"instances":[{"id":"iid1","href":"iid1link"},{"id":"iid2","href":"iid2link"}]}}`

	Convey("When no import-job is returned", t, func() {
		mockedAPI := getMockImportAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: 404, Body: ""})
		job, isFatal, err := mockedAPI.GetImportJob(ctx, jobID)
		So(err, ShouldBeNil)
		So(job, ShouldResemble, ImportJob{})
		So(isFatal, ShouldBeFalse)
	})

	Convey("When bad json is returned", t, func() {
		mockedAPI := getMockImportAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: 200, Body: "oops"})
		_, isFatal, err := mockedAPI.GetImportJob(ctx, jobID)
		So(err, ShouldNotBeNil)
		So(isFatal, ShouldBeTrue)
	})

	Convey("When server error is returned", t, func() {
		mockedAPI := getMockImportAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: 500, Body: "[]"})
		_, isFatal, err := mockedAPI.GetImportJob(ctx, jobID)
		So(err, ShouldNotBeNil)
		So(isFatal, ShouldBeFalse)
	})

	Convey("When a single-instance import-job is returned", t, func() {
		mockedAPI := getMockImportAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: 200, Body: jobJSON})
		job, isFatal, err := mockedAPI.GetImportJob(ctx, jobID)
		So(err, ShouldBeNil)
		So(job, ShouldResemble, ImportJob{JobID: jobID, Links: LinkMap{Instances: []InstanceLink{InstanceLink{ID: "iid1", Link: "iid1link"}}}})
		So(isFatal, ShouldBeFalse)
	})

	Convey("When a multiple-instance import-job is returned", t, func() {
		mockedAPI := getMockImportAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: 200, Body: jobMultiInstJSON})
		job, isFatal, err := mockedAPI.GetImportJob(ctx, jobID)
		So(err, ShouldBeNil)
		So(job, ShouldResemble, ImportJob{
			JobID: jobID,
			Links: LinkMap{
				Instances: []InstanceLink{
					InstanceLink{ID: "iid1", Link: "iid1link"},
					InstanceLink{ID: "iid2", Link: "iid2link"},
				},
			},
		})
		So(isFatal, ShouldBeFalse)
	})
}

func TestUpdateImportJobState(t *testing.T) {
	jobID := "jid0"
	Convey("When bad request is returned", t, func() {
		mockedAPI := getMockImportAPI(http.Request{Method: "PUT"}, MockedHTTPResponse{StatusCode: 400, Body: ""})
		err := mockedAPI.UpdateImportJobState(ctx, jobID, "newState")
		So(err, ShouldNotBeNil)
	})

	Convey("When server error is returned", t, func() {
		mockedAPI := getMockImportAPI(http.Request{Method: "PUT"}, MockedHTTPResponse{StatusCode: 500, Body: "dnm"})
		err := mockedAPI.UpdateImportJobState(ctx, jobID, "newState")
		So(err, ShouldNotBeNil)
	})

	Convey("When a single import-instance is returned", t, func() {
		mockedAPI := getMockImportAPI(http.Request{Method: "PUT"}, MockedHTTPResponse{StatusCode: 200, Body: ""})
		err := mockedAPI.UpdateImportJobState(ctx, jobID, "newState")
		So(err, ShouldBeNil)
	})
}
