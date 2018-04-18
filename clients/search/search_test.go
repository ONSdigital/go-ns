package search

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/ONSdigital/go-ns/common/mock_common"
	"github.com/ONSdigital/go-ns/rchttp"
	"github.com/golang/mock/gomock"

	. "github.com/smartystreets/goconvey/convey"
)

var ctx = context.Background()

func TestSearchUnit(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	limit := 1
	offset := 1

	Convey("test New creates a valid Client instance", t, func() {
		cli := New("http://localhost:22000")
		So(cli.url, ShouldEqual, "http://localhost:22000")
		So(cli.cli, ShouldHaveSameTypeAs, rchttp.DefaultClient)
	})

	Convey("test Dimension Method", t, func() {
		searchResp, err := ioutil.ReadFile("./search_mocks/search.json")
		So(err, ShouldBeNil)

		Convey("test Dimension successfully returns a model upon a 200 response from search api", func() {
			mockCli := mock_common.NewMockRCHTTPClienter(mockCtrl)

			mockCli.EXPECT().Do(gomock.Any(), gomock.Any()).Return(
				&http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(bytes.NewReader(searchResp)),
				}, nil,
			)

			searchCli := &Client{
				cli: mockCli,
				url: "http://localhost:22000",
			}
			ctx := context.Background()
			m, err := searchCli.Dimension(ctx, "12345", "time-series", "1", "geography", "Newport", Config{Limit: &limit, Offset: &offset})
			So(err, ShouldBeNil)
			So(m.Count, ShouldEqual, 1)
			So(m.Limit, ShouldEqual, 1)
			So(m.Offset, ShouldEqual, 0)
			So(m.TotalCount, ShouldEqual, 1)
			So(m.Items, ShouldHaveLength, 1)

			item := m.Items[0]
			So(item.Code, ShouldEqual, "6789")
			So(item.DimensionOptionURL, ShouldEqual, "http://localhost:22000/datasets/12345/editions/time-series/versions/1/dimensions/geography/options/6789")
			So(item.HasData, ShouldBeTrue)
			So(item.Label, ShouldEqual, "Newport")
			So(item.Matches.Label, ShouldHaveLength, 1)
			So(item.NumberOfChildren, ShouldEqual, 3)

			label := item.Matches.Label[0]
			So(label.Start, ShouldEqual, 0)
			So(label.End, ShouldEqual, 6)
		})

		Convey("test Dimension returns error from HTTPClient if it throws an error", func() {
			mockCli := mock_common.NewMockRCHTTPClienter(mockCtrl)

			mockCli.EXPECT().Do(gomock.Any(), gomock.Any()).Return(
				nil, errors.New("client threw an error"),
			)

			searchCli := &Client{
				cli: mockCli,
				url: "http://localhost:22000",
			}

			m, err := searchCli.Dimension(ctx, "12345", "time-series", "1", "geography", "Newport", Config{Limit: &limit, Offset: &offset})
			So(err.Error(), ShouldEqual, "client threw an error")
			So(m, ShouldBeNil)
		})

		Convey("test Dimension returns error if HTTP Status code is not 200", func() {
			mockCli := mock_common.NewMockRCHTTPClienter(mockCtrl)

			mockCli.EXPECT().Do(gomock.Any(), gomock.Any()).Return(
				&http.Response{
					StatusCode: http.StatusBadRequest,
					Body:       ioutil.NopCloser(bytes.NewReader(nil)),
				}, nil,
			)

			searchCli := &Client{
				cli: mockCli,
				url: "http://localhost:22000",
			}

			m, err := searchCli.Dimension(ctx, "12345", "time-series", "1", "geography", "Newport", Config{Limit: &limit, Offset: &offset})
			So(err.Error(), ShouldEqual, "invalid response from search api - should be: 200, got: 400, path: http://localhost:22000/search/datasets/12345/editions/time-series/versions/1/dimensions/geography?limit=1&offset=1&q=Newport")
			So(err.(*ErrInvalidSearchAPIResponse).Code(), ShouldEqual, http.StatusBadRequest)
			So(m, ShouldBeNil)
		})
	})

	Convey("test Healthcheck method", t, func() {
		Convey("test Healthcheck returns no error upon a 200 response from search api", func() {
			mockCli := mock_common.NewMockRCHTTPClienter(mockCtrl)
			mockCli.EXPECT().Get(ctx, "http://localhost:22000/healthcheck").Return(
				&http.Response{
					StatusCode: http.StatusOK,
				}, nil,
			)

			searchCli := &Client{
				cli: mockCli,
				url: "http://localhost:22000",
			}

			s, err := searchCli.Healthcheck()
			So(err, ShouldBeNil)
			So(s, ShouldEqual, service)
		})

		Convey("test Healthcheck returns error from HTTPClient if it throws an error", func() {
			mockCli := mock_common.NewMockRCHTTPClienter(mockCtrl)
			mockCli.EXPECT().Get(ctx, "http://localhost:22000/healthcheck").Return(
				nil, errors.New("client threw an error"),
			)

			searchCli := &Client{
				cli: mockCli,
				url: "http://localhost:22000",
			}

			s, err := searchCli.Healthcheck()
			So(err.Error(), ShouldEqual, "client threw an error")
			So(s, ShouldEqual, service)
		})

		Convey("test Dimension returns error if HTTP Status code is not 200", func() {
			mockCli := mock_common.NewMockRCHTTPClienter(mockCtrl)
			mockCli.EXPECT().Get(ctx, "http://localhost:22000/healthcheck").Return(
				&http.Response{
					StatusCode: http.StatusInternalServerError,
				}, nil,
			)

			searchCli := &Client{
				cli: mockCli,
				url: "http://localhost:22000",
			}

			s, err := searchCli.Healthcheck()
			So(err.Error(), ShouldEqual, "invalid response from search api - should be: 200, got: 500, path: http://localhost:22000/healthcheck")
			So(err.(*ErrInvalidSearchAPIResponse).Code(), ShouldEqual, http.StatusInternalServerError)
			So(s, ShouldEqual, service)
		})
	})
}
