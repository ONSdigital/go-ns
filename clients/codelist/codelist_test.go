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
	testServiceAuthToken = "666"
	testHost             = "http://localhost:8080"
	dimensionValues      = DimensionValues{
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

		actual, err := codelistClient.GetValues(nil, testServiceAuthToken, "999")

		So(err, ShouldBeNil)
		So(actual, ShouldResemble, dimensionValues)

		calls := clienter.DoCalls()
		So(calls, ShouldHaveLength, 1)

		req := calls[0].Req
		So(req.URL.Path, ShouldEqual, "/code-lists/999/codes")
		So(req.Method, ShouldEqual, "GET")
		So(req.Body, ShouldBeNil)
		So(req.Header.Get(common.AuthHeaderKey), ShouldEqual, common.BearerPrefix+testServiceAuthToken)
		So(body.IsClosed, ShouldBeTrue)
	})

	Convey("should return expect error if clienter.Do returns an error", t, func() {
		expectedErr := errors.New("lets get schwifty")

		clienter := getClienterMock(nil, expectedErr)

		codelistClient := &Client{cli: clienter, url: testHost}

		actual, err := codelistClient.GetValues(nil, testServiceAuthToken, "999")

		So(err, ShouldResemble, expectedErr)
		So(actual, ShouldResemble, DimensionValues{})

		calls := clienter.DoCalls()
		So(calls, ShouldHaveLength, 1)

		req := calls[0].Req
		So(req.URL.Path, ShouldEqual, "/code-lists/999/codes")
		So(req.Method, ShouldEqual, "GET")
		So(req.Body, ShouldBeNil)
		So(req.Header.Get(common.AuthHeaderKey), ShouldEqual, common.BearerPrefix+testServiceAuthToken)
	})

	Convey("should return expected error for non 200 response status", t, func() {
		body := httpmocks.NewReadCloserMock([]byte{}, nil)
		resp := httpmocks.NewResponseMock(body, 500)

		clienter := getClienterMock(resp, nil)
		codelistClient := &Client{cli: clienter, url: testHost}

		expectedURI := fmt.Sprintf("%s/code-lists/%s/codes", testHost, "999")
		expectedErr := &ErrInvalidCodelistAPIResponse{http.StatusOK, 500, expectedURI}

		dimensionValues, err := codelistClient.GetValues(nil, testServiceAuthToken, "999")

		So(err, ShouldResemble, expectedErr)
		So(dimensionValues, ShouldResemble, DimensionValues{})

		calls := clienter.DoCalls()
		So(calls, ShouldHaveLength, 1)

		req := calls[0].Req
		So(req.URL.Path, ShouldEqual, "/code-lists/999/codes")
		So(req.Method, ShouldEqual, "GET")
		So(req.Body, ShouldBeNil)
		So(req.Header.Get(common.AuthHeaderKey), ShouldEqual, common.BearerPrefix+testServiceAuthToken)
		So(body.IsClosed, ShouldBeTrue)
	})

	Convey("should return expected error if ioutil.ReadAll returns an error", t, func() {
		expectedErr := errors.New("lets get schwifty")
		body := httpmocks.NewReadCloserMock(nil, expectedErr)
		resp := httpmocks.NewResponseMock(body, 200)
		clienter := getClienterMock(resp, nil)
		codelistClient := &Client{cli: clienter, url: testHost}

		dimensionValues, err := codelistClient.GetValues(nil, testServiceAuthToken, "999")

		So(err, ShouldResemble, expectedErr)
		So(dimensionValues, ShouldResemble, DimensionValues{})

		calls := clienter.DoCalls()
		So(calls, ShouldHaveLength, 1)

		req := calls[0].Req
		So(req.URL.Path, ShouldEqual, "/code-lists/999/codes")
		So(req.Method, ShouldEqual, "GET")
		So(req.Body, ShouldBeNil)
		So(req.Header.Get(common.AuthHeaderKey), ShouldEqual, common.BearerPrefix+testServiceAuthToken)
		So(body.IsClosed, ShouldBeTrue)
	})
}

func TestClient_GetIDNameMap(t *testing.T) {

	Convey("give client.Do returns an error", t, func() {
		expectedErr := errors.New("bork")

		clienter := getClienterMock(nil, expectedErr)

		codelistclient := &Client{url: testHost, cli: clienter}

		Convey("when codelistClient.GetIDNameMap is called", func() {
			actual, err := codelistclient.GetIDNameMap(nil, "666", testServiceAuthToken)

			Convey("then the expected error is returned", func() {
				So(actual, ShouldBeNil)
				So(err, ShouldResemble, expectedErr)
			})

			Convey("and client.Do should be called 1 time with the expected parameters", func() {
				calls := clienter.DoCalls()
				So(calls, ShouldHaveLength, 1)

				req := calls[0].Req
				So(req.URL.Path, ShouldEqual, "/code-lists/666/codes")
				So(req.URL.Host, ShouldEqual, "localhost:8080")
				So(req.Method, ShouldEqual, "GET")
				So(req.Body, ShouldBeNil)
				So(req.Header.Get(common.AuthHeaderKey), ShouldEqual, common.BearerPrefix+testServiceAuthToken)
			})
		})
	})

	Convey("given ioutil.ReadAll returns an error", t, func() {
		expectedErr := errors.New("I wander out where you can't see Inside my shell, I wait and bleed")
		body := httpmocks.NewReadCloserMock(nil, expectedErr)
		resp := httpmocks.NewResponseMock(body, 200)
		clienter := getClienterMock(resp, nil)

		codelistclient := &Client{url: testHost, cli: clienter}

		Convey("when codelistClient.GetIDNameMap is called", func() {
			actual, err := codelistclient.GetIDNameMap(nil, "666", testServiceAuthToken)

			Convey("then the expected error is returned", func() {
				So(actual, ShouldBeNil)
				So(err, ShouldEqual, expectedErr)
			})

			Convey("and client.Do should be called 1 time with the expected parameters", func() {
				calls := clienter.DoCalls()
				So(calls, ShouldHaveLength, 1)

				req := calls[0].Req
				So(req.URL.Path, ShouldEqual, "/code-lists/666/codes")
				So(req.URL.Host, ShouldEqual, "localhost:8080")
				So(req.Method, ShouldEqual, "GET")
				So(req.Body, ShouldBeNil)
				So(req.Header.Get(common.AuthHeaderKey), ShouldEqual, common.BearerPrefix+testServiceAuthToken)
			})

			Convey("and the response body is closed", func() {
				So(body.IsClosed, ShouldBeTrue)
			})
		})
	})

	Convey("given unmarshalling the response body returns error", t, func() {
		// return bytes incompatible with the expected return type
		b := httpmocks.GetEntityBytes(t, []int{1, 2, 3, 4, 5, 6})

		body := httpmocks.NewReadCloserMock(b, nil)
		resp := httpmocks.NewResponseMock(body, 200)

		clienter := getClienterMock(resp, nil)

		codelistclient := &Client{url: testHost, cli: clienter}

		Convey("when codelistClient.GetIDNameMap is called", func() {
			actual, err := codelistclient.GetIDNameMap(nil, "666", testServiceAuthToken)

			Convey("then the expected error is returned", func() {
				So(actual, ShouldBeNil)
				So(err, ShouldNotBeNil)
			})

			Convey("and client.Do should be called 1 time with the expected parameters", func() {
				calls := clienter.DoCalls()
				So(calls, ShouldHaveLength, 1)

				req := calls[0].Req
				So(req.URL.Path, ShouldEqual, "/code-lists/666/codes")
				So(req.URL.Host, ShouldEqual, "localhost:8080")
				So(req.Method, ShouldEqual, "GET")
				So(req.Body, ShouldBeNil)
				So(req.Header.Get(common.AuthHeaderKey), ShouldEqual, common.BearerPrefix+testServiceAuthToken)
			})

			Convey("and the response body is closed", func() {
				So(body.IsClosed, ShouldBeTrue)
			})
		})
	})

	Convey("given a successful http response is returned", t, func() {
		b := httpmocks.GetEntityBytes(t, dimensionValues)
		body := httpmocks.NewReadCloserMock(b, nil)
		resp := httpmocks.NewResponseMock(body, 200)
		clienter := getClienterMock(resp, nil)

		codelistclient := &Client{url: testHost, cli: clienter}

		Convey("when codelistClient.GetIDNameMap is called", func() {
			actual, err := codelistclient.GetIDNameMap(nil, "666", testServiceAuthToken)

			Convey("then the expected ID Name map is returned", func() {
				expected := map[string]string{"123": "Schwifty"}
				So(actual, ShouldResemble, expected)
				So(err, ShouldBeNil)
			})

			Convey("and client.Do should be called 1 time with the expected parameters", func() {
				calls := clienter.DoCalls()
				So(calls, ShouldHaveLength, 1)

				req := calls[0].Req
				So(req.URL.Path, ShouldEqual, "/code-lists/666/codes")
				So(req.URL.Host, ShouldEqual, "localhost:8080")
				So(req.Method, ShouldEqual, "GET")
				So(req.Body, ShouldBeNil)
				So(req.Header.Get(common.AuthHeaderKey), ShouldEqual, common.BearerPrefix+testServiceAuthToken)
			})

			Convey("and the response body is closed", func() {
				So(body.IsClosed, ShouldBeTrue)
			})
		})
	})
}

func TestClient_GetGeographyCodeLists(t *testing.T) {
	Convey("given clienter.Do returns an error", t, func() {
		expectedErr := errors.New("Master! Master! Obey your Master!")
		clienter := getClienterMock(nil, expectedErr)

		codelistclient := &Client{
			url: testHost,
			cli: clienter,
		}

		Convey("when codelistClient.GetGeographyCodeLists is called", func() {
			codelistclient.GetGeographyCodeLists(nil, testServiceAuthToken)
		})
	})
}

func getClienterMock(resp *http.Response, err error) *rchttp.ClienterMock {
	return &rchttp.ClienterMock{
		DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
			return resp, err
		},
	}
}
