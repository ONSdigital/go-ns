package filter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/ONSdigital/go-ns/common/commontest"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
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

func TestClient_GetOutput(t *testing.T) {
	filterOutputID := "foo"
	filterOutputBody := `{"filter_id":"` + filterOutputID + `"}`
	Convey("When bad request is returned", t, func() {
		mockedAPI := getMockfilterAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: 400, Body: ""})
		_, err := mockedAPI.GetOutput(ctx, serviceToken, downloadServiceToken, filterOutputID)
		So(err, ShouldNotBeNil)
	})

	Convey("When server error is returned", t, func() {
		mockedAPI := getMockfilterAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: 500, Body: "qux"})
		_, err := mockedAPI.GetOutput(ctx, serviceToken, downloadServiceToken, filterOutputID)
		So(err, ShouldNotBeNil)
	})

	Convey("When a filter-instance is returned", t, func() {
		mockedAPI := getMockfilterAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: 200, Body: filterOutputBody})
		_, err := mockedAPI.GetOutput(ctx, serviceToken, downloadServiceToken, filterOutputID)
		So(err, ShouldBeNil)
	})
}

func TestClient_GetDimension(t *testing.T) {
	filterOutputID := "foo"
	name := "corge"
	dimensionBody := `{
		"url": "www.ons.gov.uk", 
		"name": "quuz", 
		"options": ["corge"]}`
	Convey("When bad request is returned", t, func() {
		mockedAPI := getMockfilterAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: 400, Body: ""})
		_, err := mockedAPI.GetDimension(ctx, serviceToken, filterOutputID, name)
		So(err, ShouldNotBeNil)
	})

	Convey("When server error is returned", t, func() {
		mockedAPI := getMockfilterAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: 500, Body: "qux"})
		_, err := mockedAPI.GetDimension(ctx, serviceToken, filterOutputID, name)
		So(err, ShouldNotBeNil)
	})

	Convey("When a dimension-instance is returned", t, func() {
		mockedAPI := getMockfilterAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: 200, Body: dimensionBody})
		_, err := mockedAPI.GetDimension(ctx, serviceToken, filterOutputID, name)
		So(err, ShouldBeNil)
	})
}

func TestClient_GetDimensions(t *testing.T) {
	filterOutputID := "foo"
	dimensionBody := `[{
		"url": "www.ons.gov.uk", 
		"name": "quuz", 
		"options": ["corge"]}]`

	Convey("When bad request is returned", t, func() {
		mockedAPI := getMockfilterAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: 400, Body: ""})
		_, err := mockedAPI.GetDimensions(ctx, serviceToken, filterOutputID)
		So(err, ShouldNotBeNil)
	})

	Convey("When server error is returned", t, func() {
		mockedAPI := getMockfilterAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: 500, Body: "qux"})
		_, err := mockedAPI.GetDimensions(ctx, serviceToken, filterOutputID)
		So(err, ShouldNotBeNil)
	})

	Convey("When a dimension-instance is returned", t, func() {
		mockedAPI := getMockfilterAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: 200, Body: dimensionBody})
		_, err := mockedAPI.GetDimensions(ctx, serviceToken, filterOutputID)
		So(err, ShouldBeNil)
	})
}

func TestClient_GetDimensionOptions(t *testing.T) {
	filterOutputID := "foo"
	dimensionBody := `[{"dimension_options_url":"quux","option": "quuz"}]`
	name := "corge"
	Convey("When bad request is returned", t, func() {
		mockedAPI := getMockfilterAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: 400, Body: ""})
		_, err := mockedAPI.GetDimensionOptions(ctx, serviceToken, filterOutputID, name)
		So(err, ShouldNotBeNil)
	})

	Convey("When server error is returned", t, func() {
		mockedAPI := getMockfilterAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: 500, Body: "qux"})
		_, err := mockedAPI.GetDimensionOptions(ctx, serviceToken, filterOutputID, name)
		So(err, ShouldNotBeNil)
	})

	Convey("When a dimension option is returned", t, func() {
		mockedAPI := getMockfilterAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: 200, Body: dimensionBody})
		_, err := mockedAPI.GetDimensionOptions(ctx, serviceToken, filterOutputID, name)
		So(err, ShouldBeNil)
	})
}

func TestClient_CreateBlueprint(t *testing.T) {
	datasetID := "foo"
	edition := "quux"
	version := "1"
	names := []string{"quuz", "corge"}

	checkResponse := func(mockRCHTTPCli *commontest.RCHTTPClienterMock, expectedFilterID string) {
		So(len(mockRCHTTPCli.DoCalls()), ShouldEqual, 1)

		actualBody, _ := ioutil.ReadAll(mockRCHTTPCli.DoCalls()[0].Req.Body)
		var actualVersion string
		json.Unmarshal(actualBody, &actualVersion)
		So(actualVersion, ShouldResemble, expectedFilterID)
	}

	Convey("Given a valid Blueprint is returned", t, func() {
		mockRCHTTPCli := &commontest.RCHTTPClienterMock{
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusCreated,
					Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"filter_id":""}`))),
				}, nil
			},
		}

		cli := Client{
			cli: mockRCHTTPCli,
			url: "http://localhost:8080",
		}

		Convey("when CreateBlueprint is called", func() {
			bp, err := cli.CreateBlueprint(ctx, serviceToken, downloadServiceToken, datasetID, edition, version, names)

			Convey("then no error is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("and rchttp client is called one time with the expected parameters", func() {
				checkResponse(mockRCHTTPCli, bp)
			})
		})
	})

	Convey("given rchttpclient.do returns an error", t, func() {
		mockErr := errors.New("foo")
		mockRCHTTPCli := &commontest.RCHTTPClienterMock{
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return nil, mockErr
			},
		}

		cli := Client{cli: mockRCHTTPCli, url: "http://localhost:8080"}

		Convey("when CreateBlueprint is called", func() {
			bp, err := cli.CreateBlueprint(ctx, serviceToken, downloadServiceToken, datasetID, edition, version, names)

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldResemble, mockErr.Error())
			})

			Convey("and rchttpclient.do is called 1 time with the expected parameters", func() {
				checkResponse(mockRCHTTPCli, bp)
			})
		})
	})

	Convey("given rchttpclient.do returns a non 200 response status", t, func() {
		url := "http://localhost:8080"
		mockInvalidStatusCodeError := ErrInvalidFilterAPIResponse{http.StatusCreated, 500, url + "/filters"}
		mockRCHTTPCli := &commontest.RCHTTPClienterMock{
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusInternalServerError,
					Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
				}, nil
			},
		}

		cli := Client{cli: mockRCHTTPCli, url: url}

		Convey("when CreateBlueprint is called", func() {
			bp, err := cli.CreateBlueprint(ctx, serviceToken, downloadServiceToken, datasetID, edition, version, names)

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldResemble, mockInvalidStatusCodeError.Error())
			})

			Convey("and rchttpclient.do is called 1 time with the expected parameters", func() {
				checkResponse(mockRCHTTPCli, bp)
			})
		})
	})
}

func TestClient_UpdateBlueprint(t *testing.T) {
	model := Model{
		FilterID:    "",
		InstanceID:  "",
		Links:       Links{},
		DatasetID:   "",
		Edition:     "",
		Version:     "",
		State:       "",
		Dimensions:  nil,
		Downloads:   nil,
		Events:      nil,
		IsPublished: false,
	}
	doSubmit := true

	checkResponse := func(mockRCHTTPCli *commontest.RCHTTPClienterMock, expectedModel Model) {
		So(len(mockRCHTTPCli.DoCalls()), ShouldEqual, 1)

		actualBody, _ := ioutil.ReadAll(mockRCHTTPCli.DoCalls()[0].Req.Body)
		var actualModel Model

		json.Unmarshal(actualBody, &actualModel)
		So(actualModel, ShouldResemble, expectedModel)
	}

	Convey("Given a valid blueprint update is given", t, func() {
		mockRCHTTPCli := &commontest.RCHTTPClienterMock{
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"filter_id":""}`))),
				}, nil
			},
		}

		cli := Client{
			cli: mockRCHTTPCli,
			url: "http://localhost:8080",
		}

		Convey("when CreateBlueprint is called", func() {
			bp, err := cli.UpdateBlueprint(ctx, serviceToken, downloadServiceToken, model, doSubmit)

			Convey("then no error is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("and rchttp client is called one time with the expected parameters", func() {
				checkResponse(mockRCHTTPCli, bp)
			})
		})
	})

	Convey("given rchttpclient.do returns an error", t, func() {
		mockErr := errors.New("foo")
		mockRCHTTPCli := &commontest.RCHTTPClienterMock{
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return nil, mockErr
			},
		}

		cli := Client{cli: mockRCHTTPCli, url: "http://localhost:8080"}

		Convey("when CreateBlueprint is called", func() {
			bp, err := cli.UpdateBlueprint(ctx, serviceToken, downloadServiceToken, model, doSubmit)

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldResemble, mockErr.Error())
			})

			Convey("and rchttpclient.do is called 1 time with the expected parameters", func() {
				checkResponse(mockRCHTTPCli, bp)
			})
		})
	})

	Convey("given rchttpclient.do returns a non 200 response status", t, func() {
		url := "http://localhost:8080"
		mockInvalidStatusCodeError := ErrInvalidFilterAPIResponse{http.StatusOK, 500, url + "/filters/?submitted=" + strconv.FormatBool(doSubmit)}
		mockRCHTTPCli := &commontest.RCHTTPClienterMock{
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusInternalServerError,
					Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
				}, nil
			},
		}

		cli := Client{cli: mockRCHTTPCli, url: url}

		Convey("when CreateBlueprint is called", func() {
			bp, err := cli.UpdateBlueprint(ctx, serviceToken, downloadServiceToken, model, doSubmit)

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldResemble, mockInvalidStatusCodeError.Error())
			})

			Convey("and rchttpclient.do is called 1 time with the expected parameters", func() {
				checkResponse(mockRCHTTPCli, bp)
			})
		})
	})

}

func TestClient_AddDimensionValue(t *testing.T) {
	filterID := "baz"
	name := "quz"

	Convey("Given a valid dimension value is added", t, func() {
		mockRCHTTPCli := &commontest.RCHTTPClienterMock{
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusCreated,
					Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"filter_id":""}`))),
				}, nil
			},
		}

		cli := Client{
			cli: mockRCHTTPCli,
			url: "http://localhost:8080",
		}

		Convey("when AddDimensionValue is called", func() {
			err := cli.AddDimensionValue(ctx, serviceToken, filterID, name, service)

			Convey("then no error is returned", func() {
				So(err, ShouldBeNil)
			})
		})
	})

	Convey("given rchttpclient.do returns an error", t, func() {
		mockErr := errors.New("foo")
		mockRCHTTPCli := &commontest.RCHTTPClienterMock{
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return nil, mockErr
			},
		}

		cli := Client{cli: mockRCHTTPCli, url: "http://localhost:8080"}

		Convey("when AddDimensionValue is called", func() {
			err := cli.AddDimensionValue(ctx, serviceToken, filterID, name, service)

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldResemble, mockErr.Error())
			})

		})
	})

	Convey("given rchttpclient.do returns a non 200 response status", t, func() {
		url := "http://localhost:8080"
		uri := url + "/filters/" + filterID + "/dimensions/" + name + "/options/filter-api"
		mockInvalidStatusCodeError := ErrInvalidFilterAPIResponse{http.StatusCreated, 500, uri}
		mockRCHTTPCli := &commontest.RCHTTPClienterMock{
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusInternalServerError,
					Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
				}, nil
			},
		}

		cli := Client{cli: mockRCHTTPCli, url: url}

		Convey("when AddDimensionValue is called", func() {
			err := cli.AddDimensionValue(ctx, serviceToken, filterID, name, service)

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldResemble, mockInvalidStatusCodeError.Error())
			})

		})
	})
}

func TestClient_RemoveDimensionValue(t *testing.T) {
	filterID := "baz"
	name := "quz"
	Convey("Given a dimension value is removed", t, func() {
		mockRCHTTPCli := &commontest.RCHTTPClienterMock{
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusNoContent,
					Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"filter_id":""}`))),
				}, nil
			},
		}

		cli := Client{
			cli: mockRCHTTPCli,
			url: "http://localhost:8080",
		}

		Convey("when RemoveDimensionValue is called", func() {
			err := cli.RemoveDimensionValue(ctx, serviceToken, filterID, name, service)

			Convey("then no error is returned", func() {
				So(err, ShouldBeNil)
			})
		})
	})

	Convey("given rchttpclient.do returns an error", t, func() {
		mockErr := errors.New("foo")
		mockRCHTTPCli := &commontest.RCHTTPClienterMock{
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return nil, mockErr
			},
		}

		cli := Client{cli: mockRCHTTPCli, url: "http://localhost:8080"}

		Convey("when RemoveDimensionValue is called", func() {
			err := cli.RemoveDimensionValue(ctx, serviceToken, filterID, name, service)

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldResemble, mockErr.Error())
			})

		})
	})

	Convey("given rchttpclient.do returns a non 200 response status", t, func() {
		url := "http://localhost:8080"
		uri := url + "/filters/" + filterID + "/dimensions/" + name + "/options/filter-api"
		mockInvalidStatusCodeError := ErrInvalidFilterAPIResponse{http.StatusNoContent, 500, uri}
		mockRCHTTPCli := &commontest.RCHTTPClienterMock{
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusInternalServerError,
					Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
				}, nil
			},
		}

		cli := Client{cli: mockRCHTTPCli, url: url}

		Convey("when RemoveDimensionValue is called", func() {
			err := cli.RemoveDimensionValue(ctx, serviceToken, filterID, name, service)

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldResemble, mockInvalidStatusCodeError.Error())
			})

		})
	})
}

func TestClient_AddDimension(t *testing.T) {
	filterID := "baz"
	name := "quz"

	Convey("Given a dimension is provided", t, func() {
		mockRCHTTPCli := &commontest.RCHTTPClienterMock{
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusCreated,
					Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"filter_id":""}`))),
				}, nil
			},
		}

		cli := Client{
			cli: mockRCHTTPCli,
			url: "http://localhost:8080",
		}

		Convey("when AddDimension is called", func() {
			err := cli.AddDimension(ctx, serviceToken, filterID, name)

			Convey("then no error is returned", func() {
				So(err, ShouldBeNil)
			})
		})
	})

	Convey("given rchttpclient.do returns an error", t, func() {
		mockErr := errors.New("foo")
		mockRCHTTPCli := &commontest.RCHTTPClienterMock{
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return nil, mockErr
			},
		}

		cli := Client{cli: mockRCHTTPCli, url: "http://localhost:8080"}

		Convey("when AddDimension is called", func() {
			err := cli.AddDimension(ctx, serviceToken, filterID, name)

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldResemble, mockErr.Error())
			})

		})
	})

	Convey("given rchttpclient.do returns a non 200 response status", t, func() {
		url := "http://localhost:8080"
		mockInvalidStatusCodeError := errors.New("invalid status from filter api")
		mockRCHTTPCli := &commontest.RCHTTPClienterMock{
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusInternalServerError,
					Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
				}, nil
			},
		}

		cli := Client{cli: mockRCHTTPCli, url: url}

		Convey("when AddDimension is called", func() {
			err := cli.AddDimension(ctx, serviceToken, filterID, name)

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldResemble, mockInvalidStatusCodeError.Error())
			})

		})
	})
}

func TestClient_GetJobState(t *testing.T) {
	filterID := "foo"
	mockJobStateBody := `{
		"jobState": "www.ons.gov.uk"}`
	Convey("When a state is returned", t, func() {

		mockedAPI := getMockfilterAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: 200, Body: mockJobStateBody})
		_, err := mockedAPI.GetJobState(ctx, serviceToken, downloadServiceToken, filterID)
		So(err, ShouldBeNil)
	})
	Convey("When bad request is returned", t, func() {
		mockedAPI := getMockfilterAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: 400, Body: ""})
		_, err := mockedAPI.GetJobState(ctx, serviceToken, downloadServiceToken, filterID)
		So(err, ShouldNotBeNil)
	})

	Convey("When server error is returned", t, func() {
		mockedAPI := getMockfilterAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: 500, Body: "qux"})
		_, err := mockedAPI.GetJobState(ctx, serviceToken, downloadServiceToken, filterID)
		So(err, ShouldNotBeNil)
	})
}

func TestClient_AddDimensionValues(t *testing.T) {
	filterID := "baz"
	name := "quz"
	options := []string{"`quuz"}

	Convey("Given a valid dimension and filter", t, func() {
		mockRCHTTPCli := &commontest.RCHTTPClienterMock{
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusCreated,
					Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"filter_id":""}`))),
				}, nil
			},
		}

		cli := Client{
			cli: mockRCHTTPCli,
			url: "http://localhost:8080",
		}

		Convey("when AddDimensionValues is called", func() {
			err := cli.AddDimensionValues(ctx, serviceToken, filterID, name, options)

			Convey("then no error is returned", func() {
				So(err, ShouldBeNil)
			})
		})
	})

	Convey("given rchttpclient.do returns an error", t, func() {
		mockErr := errors.New("foo")
		mockRCHTTPCli := &commontest.RCHTTPClienterMock{
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return nil, mockErr
			},
		}

		cli := Client{cli: mockRCHTTPCli, url: "http://localhost:8080"}

		Convey("when AddDimensionValues is called", func() {
			err := cli.AddDimensionValues(ctx, serviceToken, filterID, name, options)

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldResemble, mockErr.Error())
			})

		})
	})

	Convey("given rchttpclient.do returns a non 200 response status", t, func() {
		url := "http://localhost:8080"
		uri := url + "/filters/" + filterID + "/dimensions/" + name
		mockInvalidStatusCodeError := &ErrInvalidFilterAPIResponse{http.StatusCreated, http.StatusInternalServerError, uri}
		mockRCHTTPCli := &commontest.RCHTTPClienterMock{
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusInternalServerError,
					Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
				}, nil
			},
		}

		cli := Client{cli: mockRCHTTPCli, url: url}

		Convey("when AddDimensionValues is called", func() {
			err := cli.AddDimensionValues(ctx, serviceToken, filterID, name, options)

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldResemble, mockInvalidStatusCodeError.Error())
			})

		})
	})
}

func TestClient_GetPreview(t *testing.T) {
	filterOutputID := "foo"
	previewBody := `{"somePreview":""}`
	Convey("When bad request is returned", t, func() {
		mockedAPI := getMockfilterAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: 400, Body: ""})
		_, err := mockedAPI.GetPreview(ctx, serviceToken, downloadServiceToken, filterOutputID)
		So(err, ShouldNotBeNil)
	})

	Convey("When server error is returned", t, func() {
		mockedAPI := getMockfilterAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: 500, Body: "qux"})
		_, err := mockedAPI.GetPreview(ctx, serviceToken, downloadServiceToken, filterOutputID)
		So(err, ShouldNotBeNil)
	})

	Convey("When a preview is returned", t, func() {
		mockedAPI := getMockfilterAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: 200, Body: previewBody})
		_, err := mockedAPI.GetPreview(ctx, serviceToken, downloadServiceToken, filterOutputID)
		So(err, ShouldBeNil)
	})
}
