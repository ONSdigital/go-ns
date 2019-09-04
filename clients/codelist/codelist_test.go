package codelist

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ONSdigital/dp-mocking/httpmocks"
	rchttp "github.com/ONSdigital/dp-rchttp"
	"github.com/ONSdigital/go-ns/common"
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"testing"
)

var (
	testAuthToken   = "666"
	testHost        = "http://localhost:8080"
	dimensionValues = DimensionValues{
		Items: []Item{
			{
				ID:    "123",
				Label: "Schwifty",
			},
		},
		NumberOfResults: 1,
	}
)

func TestClient_GetValues(t *testing.T) {

	Convey("should return expect values for 200 status response", t, func() {
		b, err := json.Marshal(dimensionValues)
		So(err, ShouldBeNil)

		body := httpmocks.NewReadCloserMock(b, nil)
		resp := httpmocks.NewResponseMock(body, 200)

		clienter := getClienterMock(resp, nil)
		codelistClient := &Client{cli: clienter, url: testHost}

		actual, err := codelistClient.GetValues(nil, testAuthToken, "999")

		So(err, ShouldBeNil)
		So(actual, ShouldResemble, dimensionValues)

		calls := clienter.DoCalls()
		So(calls, ShouldHaveLength, 1)

		req := calls[0].Req
		So(req.URL.Path, ShouldEqual, "/code-lists/999/codes")
		So(req.Method, ShouldEqual, "GET")
		So(req.Body, ShouldBeNil)
		So(req.Header.Get(common.AuthHeaderKey), ShouldEqual, common.BearerPrefix+testAuthToken)
		So(body.IsClosed, ShouldBeTrue)
	})

	Convey("should return expect error if clienter.Do returns an error", t, func() {
		expectedErr := errors.New("lets get schwifty")

		clienter := getClienterMock(nil, expectedErr)

		codelistClient := &Client{cli: clienter, url: testHost}

		actual, err := codelistClient.GetValues(nil, testAuthToken, "999")

		So(err, ShouldResemble, expectedErr)
		So(actual, ShouldResemble, DimensionValues{})

		calls := clienter.DoCalls()
		So(calls, ShouldHaveLength, 1)

		req := calls[0].Req
		So(req.URL.Path, ShouldEqual, "/code-lists/999/codes")
		So(req.Method, ShouldEqual, "GET")
		So(req.Body, ShouldBeNil)
		So(req.Header.Get(common.AuthHeaderKey), ShouldEqual, common.BearerPrefix+testAuthToken)
	})

	Convey("should return expected error for non 200 response status", t, func() {
		body := httpmocks.NewReadCloserMock([]byte{}, nil)
		resp := httpmocks.NewResponseMock(body, 500)

		clienter := getClienterMock(resp, nil)
		codelistClient := &Client{cli: clienter, url: testHost}

		expectedURI := fmt.Sprintf("%s/code-lists/%s/codes", testHost, "999")
		expectedErr := &ErrInvalidCodelistAPIResponse{http.StatusOK, 500, expectedURI}

		dimensionValues, err := codelistClient.GetValues(nil, testAuthToken, "999")

		So(err, ShouldResemble, expectedErr)
		So(dimensionValues, ShouldResemble, DimensionValues{})

		calls := clienter.DoCalls()
		So(calls, ShouldHaveLength, 1)

		req := calls[0].Req
		So(req.URL.Path, ShouldEqual, "/code-lists/999/codes")
		So(req.Method, ShouldEqual, "GET")
		So(req.Body, ShouldBeNil)
		So(req.Header.Get(common.AuthHeaderKey), ShouldEqual, common.BearerPrefix+testAuthToken)
		So(body.IsClosed, ShouldBeTrue)
	})

	Convey("should return expected error if ioutil.ReadAll returns an error", t, func() {
		expectedErr := errors.New("lets get schwifty")
		body := httpmocks.NewReadCloserMock(nil, expectedErr)
		resp := httpmocks.NewResponseMock(body, 200)
		clienter := getClienterMock(resp, nil)
		codelistClient := &Client{cli: clienter, url: testHost}

		dimensionValues, err := codelistClient.GetValues(nil, testAuthToken, "999")

		So(err, ShouldResemble, expectedErr)
		So(dimensionValues, ShouldResemble, DimensionValues{})

		calls := clienter.DoCalls()
		So(calls, ShouldHaveLength, 1)

		req := calls[0].Req
		So(req.URL.Path, ShouldEqual, "/code-lists/999/codes")
		So(req.Method, ShouldEqual, "GET")
		So(req.Body, ShouldBeNil)
		So(req.Header.Get(common.AuthHeaderKey), ShouldEqual, common.BearerPrefix+testAuthToken)
		So(body.IsClosed, ShouldBeTrue)
	})
}

func getClienterMock(resp *http.Response, err error) *rchttp.ClienterMock {
	return &rchttp.ClienterMock{
		DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
			return resp, err
		},
	}
}
