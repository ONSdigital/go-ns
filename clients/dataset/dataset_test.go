package dataset

import (
	"bytes"
	"encoding/json"
	"github.com/ONSdigital/go-ns/clients/dataset/dataset_mocks"
	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestClient_PutVersion(t *testing.T) {

	checkResponse := func(mockRHTTPCli *dataset_mocks.RHTTPClientMock, expectedVersion Version) {
		So(len(mockRHTTPCli.DoCalls()), ShouldEqual, 1)

		actualBody, _ := ioutil.ReadAll(mockRHTTPCli.DoCalls()[0].Req.Body)
		var actualVersion Version
		json.Unmarshal(actualBody, &actualVersion)
		So(actualVersion, ShouldResemble, expectedVersion)
	}

	Convey("Given a valid version", t, func() {
		mockRHTTPCli := &dataset_mocks.RHTTPClientMock{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
				}, nil
			},
		}

		cli := Client{
			internalToken: "1234",
			cli:           mockRHTTPCli,
			url:           "http://localhost:8080",
		}

		Convey("when put version is called", func() {
			v := Version{ID: "666"}
			err := cli.PutVersion("123", "2017", "1", v)

			Convey("then no error is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("and rhttp client is called one time with the expected parameters", func() {
				checkResponse(mockRHTTPCli, v)
			})

			Convey("and the configured auth token is set on the outbound request", func() {
				So(mockRHTTPCli.DoCalls()[0].Req.Header.Get(authTokenHeader), ShouldEqual, "1234")
			})
		})
	})

	Convey("Given no auth token has been configured", t, func() {
		mockRHTTPCli := &dataset_mocks.RHTTPClientMock{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
				}, nil
			},
		}

		cli := Client{
			cli: mockRHTTPCli,
			url: "http://localhost:8080",
		}

		Convey("when put version is called", func() {
			v := Version{ID: "666"}
			err := cli.PutVersion("123", "2017", "1", v)

			Convey("then no error is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("and rhttp client is called one time with the expected parameters", func() {
				checkResponse(mockRHTTPCli, v)
			})

			Convey("and the outbound request is sent without any auth tokent", func() {
				So(mockRHTTPCli.DoCalls()[0].Req.Header.Get(authTokenHeader), ShouldEqual, "")
			})
		})
	})

	Convey("given rhttpclient.do returns an error", t, func() {
		mockErr := errors.New("spectacular explosion")
		mockRHTTPCli := &dataset_mocks.RHTTPClientMock{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				return nil, mockErr
			},
		}

		cli := Client{cli: mockRHTTPCli, url: "http://localhost:8080"}

		Convey("when put version is called", func() {
			v := Version{ID: "666"}
			err := cli.PutVersion("123", "2017", "1", v)

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldResemble, errors.Wrap(mockErr, "http client returned error while attempting to make request").Error())
			})

			Convey("and rhttpclient.do is called 1 time with the expected parameters", func() {
				checkResponse(mockRHTTPCli, v)
			})
		})
	})

	Convey("given rhttpclient.do returns a non 200 response status", t, func() {
		mockRHTTPCli := &dataset_mocks.RHTTPClientMock{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusInternalServerError,
					Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
				}, nil
			},
		}

		cli := Client{cli: mockRHTTPCli, url: "http://localhost:8080"}

		Convey("when put version is called", func() {
			v := Version{ID: "666"}
			err := cli.PutVersion("123", "2017", "1", v)

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldResemble, errors.Errorf("incorrect http status, expected: 200, actual: 500, uri: http://localhost:8080/datasets/123/editions/2017/versions/1").Error())
			})

			Convey("and rhttpclient.do is called 1 time with the expected parameters", func() {
				checkResponse(mockRHTTPCli, v)
			})
		})
	})

}
