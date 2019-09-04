package dataset

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/ONSdigital/dp-rchttp"
	"github.com/ONSdigital/go-ns/common"
	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"
	"io/ioutil"
	"net/http"
	"testing"
)

var ctx = context.Background()

const authToken = "auth token"

func TestClient_PutVersion(t *testing.T) {

	checkResponse := func(mockRCHTTPCli *rchttp.ClienterMock, expectedVersion Version) {
		So(len(mockRCHTTPCli.DoCalls()), ShouldEqual, 1)

		actualBody, _ := ioutil.ReadAll(mockRCHTTPCli.DoCalls()[0].Req.Body)
		var actualVersion Version
		json.Unmarshal(actualBody, &actualVersion)
		So(actualVersion, ShouldResemble, expectedVersion)
	}

	Convey("Given a valid version", t, func() {
		mockRCHTTPCli := &rchttp.ClienterMock{
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
				}, nil
			},
		}

		cli := Client{
			cli: mockRCHTTPCli,
			url: "http://localhost:8080",
		}

		Convey("when put version is called", func() {
			v := Version{ID: "666"}
			err := cli.PutVersion(ctx, "123", "2017", "1", authToken, v)

			Convey("then no error is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("and rchttp client is called one time with the expected parameters", func() {
				checkResponse(mockRCHTTPCli, v)
			})
		})
	})

	Convey("Given no auth token has been configured", t, func() {
		mockRCHTTPCli := &rchttp.ClienterMock{
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
				}, nil
			},
		}

		cli := Client{
			cli: mockRCHTTPCli,
			url: "http://localhost:8080",
		}

		Convey("when put version is called", func() {
			v := Version{ID: "666"}
			err := cli.PutVersion(ctx, "123", "2017", "1", authToken, v)

			Convey("then no error is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("and rchttp client is called one time with the expected parameters", func() {
				checkResponse(mockRCHTTPCli, v)
			})

		})
	})

	Convey("given rchttpclient.do returns an error", t, func() {
		mockErr := errors.New("spectacular explosion")
		mockRCHTTPCli := &rchttp.ClienterMock{
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return nil, mockErr
			},
		}

		cli := Client{cli: mockRCHTTPCli, url: "http://localhost:8080"}

		Convey("when put version is called", func() {
			v := Version{ID: "666"}
			err := cli.PutVersion(ctx, "123", "2017", "1", authToken, v)

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldResemble, errors.Wrap(mockErr, "http client returned error while attempting to make request").Error())
			})

			Convey("and rchttpclient.do is called 1 time with the expected parameters", func() {
				checkResponse(mockRCHTTPCli, v)
			})
		})
	})

	Convey("given rchttpclient.do returns a non 200 response status", t, func() {
		mockRCHTTPCli := &rchttp.ClienterMock{
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusInternalServerError,
					Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
				}, nil
			},
		}

		cli := Client{cli: mockRCHTTPCli, url: "http://localhost:8080"}

		Convey("when put version is called", func() {
			v := Version{ID: "666"}
			err := cli.PutVersion(ctx, "123", "2017", "1", authToken, v)

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldResemble, errors.Errorf("incorrect http status, expected: 200, actual: 500, uri: http://localhost:8080/datasets/123/editions/2017/versions/1").Error())
			})

			Convey("and rchttpclient.do is called 1 time with the expected parameters", func() {
				checkResponse(mockRCHTTPCli, v)
			})
		})
	})

}

func TestClient_IncludeCollectionID(t *testing.T) {

	Convey("Given a valid request", t, func() {
		mockRCHTTPCli := &rchttp.ClienterMock{
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return &http.Response{}, nil
			},
		}

		cli := Client{
			cli: mockRCHTTPCli,
			url: "http://localhost:8080",
		}

		Convey("when Collection-ID is present in the context", func() {
			ctx = context.WithValue(ctx, common.CollectionIDHeaderKey, "I'm a collection ID")

			Convey("and a request is made", func() {
				_, _ = cli.GetDatasets(ctx, authToken)

				Convey("then the Collection-ID is present in the request headers", func() {
					collectionIDFromRequest := mockRCHTTPCli.DoCalls()[0].Req.Header.Get(common.CollectionIDHeaderKey)
					So(collectionIDFromRequest, ShouldEqual, "I'm a collection ID")
				})
			})
		})
	})

	Convey("Given a valid request", t, func() {
		mockRCHTTPCli := &rchttp.ClienterMock{
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return &http.Response{}, nil
			},
		}

		cli := Client{
			cli: mockRCHTTPCli,
			url: "http://localhost:8080",
		}

		Convey("when Collection-ID is not present in the context", func() {
			ctx = context.Background()

			Convey("and a request is made", func() {
				_, _ = cli.GetDatasets(ctx, authToken)

				Convey("then no Collection-ID key is present in the request headers", func() {
					for k, _ := range mockRCHTTPCli.DoCalls()[0].Req.Header {
						So(k, ShouldNotEqual, "Collection-Id")
					}
				})
			})
		})
	})
}
