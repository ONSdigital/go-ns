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
)

func TestClient_Healthcheck(t *testing.T) {
	uri := "/healthcheck"

	Convey("given clienter.Get returns an error", t, func() {
		expectedErr := errors.New("disciples of the watch obey")

		clienter := &rchttp.ClienterMock{
			GetFunc: func(ctx context.Context, url string) (*http.Response, error) {
				return nil, expectedErr
			},
		}

		codelistClient := &Client{
			cli: clienter,
			url: testHost,
		}

		Convey("when codelistClient.Healthcheck is called", func() {
			actual, err := codelistClient.Healthcheck()

			Convey("then the expected error is returned", func() {
				So(actual, ShouldEqual, service)
				So(err, ShouldResemble, expectedErr)
			})

			Convey("and client.Get should be called 1 time with the expected parameters", func() {
				calls := clienter.GetCalls()
				So(calls, ShouldHaveLength, 1)
				So(calls[0].URL, ShouldEqual, testHost+uri)
			})
		})
	})

	Convey("given clienter.Get returns a non 200 response status", t, func() {
		resp := httpmocks.NewResponseMock(nil, 401)

		clienter := &rchttp.ClienterMock{
			GetFunc: func(ctx context.Context, url string) (*http.Response, error) {
				return resp, nil
			},
		}

		codelistClient := &Client{
			cli: clienter,
			url: testHost,
		}

		Convey("when codelistClient.Healthcheck is called", func() {
			actual, err := codelistClient.Healthcheck()

			Convey("then the expected error is returned", func() {
				So(actual, ShouldEqual, service)
				So(err, ShouldResemble, &ErrInvalidCodelistAPIResponse{http.StatusOK, resp.StatusCode, "/healthcheck"})
			})

			Convey("and client.Get should be called 1 time with the expected parameters", func() {
				calls := clienter.GetCalls()
				So(calls, ShouldHaveLength, 1)
				So(calls[0].URL, ShouldEqual, testHost+uri)
			})
		})
	})

	Convey("given clienter.Get returns a 200 response status", t, func() {
		resp := httpmocks.NewResponseMock(nil, 200)

		clienter := &rchttp.ClienterMock{
			GetFunc: func(ctx context.Context, url string) (*http.Response, error) {
				return resp, nil
			},
		}

		codelistClient := &Client{
			cli: clienter,
			url: testHost,
		}

		Convey("when codelistClient.Healthcheck is called", func() {
			actual, err := codelistClient.Healthcheck()

			Convey("then no error is returned", func() {
				So(actual, ShouldEqual, service)
				So(err, ShouldBeNil)
			})

			Convey("and client.Get should be called 1 time with the expected parameters", func() {
				calls := clienter.GetCalls()
				So(calls, ShouldHaveLength, 1)
				So(calls[0].URL, ShouldEqual, testHost+uri)
			})
		})
	})
}

func TestClient_GetValues(t *testing.T) {

	Convey("should return expect values for 200 status response", t, func() {
		b, err := json.Marshal(testDimensionValues)
		So(err, ShouldBeNil)

		body := httpmocks.NewReadCloserMock(b, nil)
		resp := httpmocks.NewResponseMock(body, 200)

		clienter := getClienterMock(resp, nil)
		codelistClient := &Client{cli: clienter, url: testHost}

		actual, err := codelistClient.GetValues(nil, testServiceAuthToken, "999")

		So(err, ShouldBeNil)
		So(actual, ShouldResemble, testDimensionValues)

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

	uri := "/code-lists/666/codes"
	host := "localhost:8080"

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
				So(req.URL.Path, ShouldEqual, uri)
				So(req.URL.Host, ShouldEqual, host)
				So(req.Method, ShouldEqual, "GET")
				So(req.Body, ShouldBeNil)
				So(req.Header.Get(common.AuthHeaderKey), ShouldEqual, common.BearerPrefix+testServiceAuthToken)
			})
		})
	})

	Convey("given clienter.Do returns a non 200 response status", t, func() {
		expectedErr := &ErrInvalidCodelistAPIResponse{
			expectedCode: http.StatusOK,
			actualCode:   403,
			uri:          testHost + uri,
		}

		body := httpmocks.NewReadCloserMock([]byte{}, nil)
		resp := httpmocks.NewResponseMock(body, 403)
		clienter := getClienterMock(resp, nil)

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
				So(req.URL.Path, ShouldEqual, uri)
				So(req.URL.Host, ShouldEqual, host)
				So(req.Method, ShouldEqual, "GET")
				So(req.Body, ShouldBeNil)
				So(req.Header.Get(common.AuthHeaderKey), ShouldEqual, common.BearerPrefix+testServiceAuthToken)
			})

			Convey("and the response body is closed", func() {
				So(body.IsClosed, ShouldBeTrue)
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
				So(req.URL.Path, ShouldEqual, uri)
				So(req.URL.Host, ShouldEqual, host)
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
				So(req.URL.Path, ShouldEqual, uri)
				So(req.URL.Host, ShouldEqual, host)
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
		b := httpmocks.GetEntityBytes(t, testDimensionValues)
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
				So(req.URL.Path, ShouldEqual, uri)
				So(req.URL.Host, ShouldEqual, host)
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
	uri := "/code-lists"
	host := "localhost:8080"
	query := "type=geography"

	Convey("given clienter.Do returns an error", t, func() {
		expectedErr := errors.New("master master obey your master")
		clienter := getClienterMock(nil, expectedErr)

		codelistclient := &Client{
			url: testHost,
			cli: clienter,
		}

		Convey("when codelistClient.GetGeographyCodeLists is called", func() {
			actual, err := codelistclient.GetGeographyCodeLists(nil, testServiceAuthToken)

			Convey("then the expected error is returned", func() {
				So(actual, ShouldResemble, CodeListResults{})
				So(err, ShouldResemble, expectedErr)
			})

			Convey("and client.Do should be called 1 time with the expected parameters", func() {
				calls := clienter.DoCalls()
				So(calls, ShouldHaveLength, 1)

				req := calls[0].Req
				So(req.URL.Path, ShouldEqual, uri)
				So(req.URL.RawQuery, ShouldEqual, query)
				So(req.URL.Host, ShouldEqual, host)
				So(req.Method, ShouldEqual, "GET")
				So(req.Body, ShouldBeNil)
				So(req.Header.Get(common.AuthHeaderKey), ShouldEqual, common.BearerPrefix+testServiceAuthToken)
			})
		})
	})

	Convey("given clienter.Do returns a non 200 response status", t, func() {
		body := httpmocks.NewReadCloserMock([]byte{}, nil)
		resp := httpmocks.NewResponseMock(body, 500)
		clienter := getClienterMock(resp, nil)

		codelistclient := &Client{
			url: testHost,
			cli: clienter,
		}

		Convey("when codelistClient.GetGeographyCodeLists is called", func() {
			actual, err := codelistclient.GetGeographyCodeLists(nil, testServiceAuthToken)

			Convey("then the expected error is returned", func() {
				So(actual, ShouldResemble, CodeListResults{})

				expectedErr := &ErrInvalidCodelistAPIResponse{
					expectedCode: http.StatusOK,
					actualCode:   500,
					uri:          testHost + uri + "?" + query,
				}

				So(err, ShouldResemble, expectedErr)
			})

			Convey("and client.Do should be called 1 time with the expected parameters", func() {
				calls := clienter.DoCalls()
				So(calls, ShouldHaveLength, 1)

				req := calls[0].Req
				So(req.URL.Path, ShouldEqual, uri)
				So(req.URL.RawQuery, ShouldEqual, query)
				So(req.URL.Host, ShouldEqual, host)
				So(req.Method, ShouldEqual, "GET")
				So(req.Body, ShouldBeNil)
				So(req.Header.Get(common.AuthHeaderKey), ShouldEqual, common.BearerPrefix+testServiceAuthToken)
			})

			Convey("and the response body is closed", func() {
				So(body.IsClosed, ShouldBeTrue)
			})
		})
	})

	Convey("given ioutil.ReadAll returns an error", t, func() {
		expectedErr := errors.New("Peace sells, but who's buying?")
		body := httpmocks.NewReadCloserMock([]byte{}, expectedErr)
		resp := httpmocks.NewResponseMock(body, 200)
		clienter := getClienterMock(resp, nil)

		codelistclient := &Client{
			url: testHost,
			cli: clienter,
		}

		Convey("when codelistClient.GetGeographyCodeLists is called", func() {
			actual, err := codelistclient.GetGeographyCodeLists(nil, testServiceAuthToken)

			Convey("then the expected error is returned", func() {
				So(actual, ShouldResemble, CodeListResults{})
				So(err, ShouldResemble, expectedErr)
			})

			Convey("and client.Do should be called 1 time with the expected parameters", func() {
				calls := clienter.DoCalls()
				So(calls, ShouldHaveLength, 1)

				req := calls[0].Req
				So(req.URL.Path, ShouldEqual, uri)
				So(req.URL.RawQuery, ShouldEqual, query)
				So(req.URL.Host, ShouldEqual, host)
				So(req.Method, ShouldEqual, "GET")
				So(req.Body, ShouldBeNil)
				So(req.Header.Get(common.AuthHeaderKey), ShouldEqual, common.BearerPrefix+testServiceAuthToken)
			})

			Convey("and the response body is closed", func() {
				So(body.IsClosed, ShouldBeTrue)
			})
		})
	})

	Convey("given json.Unmarshal returns an error", t, func() {
		entity := []int{1, 666, 8, 16}
		b := httpmocks.GetEntityBytes(t, entity) // return bytes that are incompatible with the expected return type
		body := httpmocks.NewReadCloserMock(b, nil)
		resp := httpmocks.NewResponseMock(body, 200)
		clienter := getClienterMock(resp, nil)

		codelistclient := &Client{
			url: testHost,
			cli: clienter,
		}

		Convey("when codelistClient.GetGeographyCodeLists is called", func() {
			actual, err := codelistclient.GetGeographyCodeLists(nil, testServiceAuthToken)

			Convey("then the expected error is returned", func() {
				So(actual, ShouldResemble, CodeListResults{})
				So(err, ShouldNotBeNil)
			})

			Convey("and client.Do should be called 1 time with the expected parameters", func() {
				calls := clienter.DoCalls()
				So(calls, ShouldHaveLength, 1)

				req := calls[0].Req
				So(req.URL.Path, ShouldEqual, uri)
				So(req.URL.RawQuery, ShouldEqual, query)
				So(req.URL.Host, ShouldEqual, host)
				So(req.Method, ShouldEqual, "GET")
				So(req.Body, ShouldBeNil)
				So(req.Header.Get(common.AuthHeaderKey), ShouldEqual, common.BearerPrefix+testServiceAuthToken)
			})

			Convey("and the response body is closed", func() {
				So(body.IsClosed, ShouldBeTrue)
			})
		})
	})

	Convey("given codelistClient is successful", t, func() {
		b := httpmocks.GetEntityBytes(t, testCodeListResults)
		body := httpmocks.NewReadCloserMock(b, nil)
		resp := httpmocks.NewResponseMock(body, 200)
		clienter := getClienterMock(resp, nil)

		codelistclient := &Client{
			url: testHost,
			cli: clienter,
		}

		Convey("when codelistClient.GetGeographyCodeLists is called", func() {
			actual, err := codelistclient.GetGeographyCodeLists(nil, testServiceAuthToken)

			Convey("then the expected result is returned", func() {
				So(actual, ShouldResemble, testCodeListResults)
				So(err, ShouldBeNil)
			})

			Convey("and client.Do should be called 1 time with the expected parameters", func() {
				calls := clienter.DoCalls()
				So(calls, ShouldHaveLength, 1)

				req := calls[0].Req
				So(req.URL.Path, ShouldEqual, uri)
				So(req.URL.RawQuery, ShouldEqual, query)
				So(req.URL.Host, ShouldEqual, host)
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

func getClienterMock(resp *http.Response, err error) *rchttp.ClienterMock {
	return &rchttp.ClienterMock{
		DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
			return resp, err
		},
	}
}
