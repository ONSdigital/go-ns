package zebedee

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ONSdigital/go-ns/common"
	"github.com/ONSdigital/go-ns/zebedee/model"
	. "github.com/smartystreets/goconvey/convey"
)

const requestContextID = "1234"
const protocol = "http://"
const zebedeeURI = "zebedeeUri"
const baseZebedeeURL = protocol + zebedeeURI

// Set these with the values you want to be returned for your test case.
var responseStub *http.Response
var errorStub error
var dataStub []byte
var pageTypeStub string
var onsErrorStub *common.ONSError
var responseBodyReadErrStub error
var responseBodyBytesStub []byte

// Struct replaces client in target code - allowing you to return stub data.
type testClient struct{}

func (*testClient) Get(url string) (resp *http.Response, err error) {
	recorder := httptest.NewRecorder()
	recorder.Header().Add("ONS-Page-Type", HomePage)
	return recorder.Result(), nil
}

// Stub Implementation of the Do func - set returned vars with the values required by your test case.
func (*testClient) Do(req *http.Request) (*http.Response, error) {
	return responseStub, errorStub
}

func ReadBodyMock(io.Reader) ([]byte, error) {
	return responseBodyBytesStub, responseBodyReadErrStub
}

func TestGetData(t *testing.T) {
	// create stub http client for test
	testHTTPClient := &testClient{}

	// inject it into an instance of zebedeeHTTPClient
	zebedeeClient := Client{testHTTPClient, baseZebedeeURL}

	Convey("Should return empty data, page type and correct error if zebedee.get data fails.", t, func() {

		// Set stub data to return for this test case.
		errorStub = errors.New("Zebedee get data error")
		dataStub = make([]byte, 0)
		pageTypeStub = ""
		onsErrorStub = common.NewONSError(errorStub, zebedeeGetError)
		onsErrorStub.AddParameter(requestContextIDParam, requestContextID)
		data, pageType, err := zebedeeClient.GetData("/", requestContextID)

		ShouldEqual(data, dataStub)
		ShouldEqual(pageType, pageTypeStub)
		ShouldEqual(err, onsErrorStub)
	})

	Convey("Should return empty data & page type and appropriate error if Zebedee returns an unexpected status code.", t, func() {

		// Set stub data to return for this test case.
		onsErrorStub = &common.ONSError{RootError: errors.New("Unexpected Response status code")}
		onsErrorStub = common.NewONSError(errors.New("Unexpected Response status code"), incorrectStatusCodeErrDesc)
		onsErrorStub.AddParameter("zebedeeURI", "/data")
		onsErrorStub.AddParameter("query", "/")
		onsErrorStub.AddParameter("expectedStatusCode", 200)
		onsErrorStub.AddParameter("actualStatusCode", 500)
		onsErrorStub.AddParameter(requestContextIDParam, requestContextID)
		errorStub = nil
		responseStub = &http.Response{StatusCode: 500}
		dataStub = make([]byte, 0)
		pageTypeStub = ""

		// Run test
		data, pageType, err := zebedeeClient.GetData("/", requestContextID)

		// assert results.
		So(err, ShouldResemble, onsErrorStub)
		So(0, ShouldEqual, bytes.Compare(data, dataStub))
		So(pageType, ShouldEqual, pageTypeStub)
	})

	Convey("Should return empty data & pageType & appropriate error if there is an error reading the response body.", t, func() {

		zebedeeClient.setResponseReader(ReadBodyMock)

		dataStub = []byte("")
		rootErr := errors.New("it broked")
		pageTypeStub = ""
		onsErrorStub = common.NewONSError(rootErr, "error reading response body")
		onsErrorStub.AddParameter(requestContextIDParam, requestContextID)

		responseStub = &http.Response{StatusCode: 200}
		responseStub.Header = make(map[string][]string, 0)
		responseStub.Header.Set(pageTypeHeader, HomePage)
		responseStub.Body = ioutil.NopCloser(bytes.NewBufferString(""))
		responseBodyReadErrStub = rootErr
		responseBodyBytesStub = []byte("")

		data, pageType, err := zebedeeClient.GetData("/", requestContextID)

		So(err, ShouldResemble, onsErrorStub)
		So(data, ShouldResemble, dataStub)
		So(pageType, ShouldEqual, pageTypeStub)
	})

	Convey("Should return expected data, pageType for successful calls.", t, func() {

		zebedeeClient.setResponseReader(ReadBodyMock)

		body := "I am Success!"

		dataStub = []byte(body)
		pageTypeStub = HomePage
		onsErrorStub = nil

		responseStub = &http.Response{StatusCode: 200}
		responseStub.Header = make(map[string][]string, 0)
		responseStub.Header.Set(pageTypeHeader, HomePage)
		responseStub.Body = ioutil.NopCloser(bytes.NewBufferString(body))
		responseBodyReadErrStub = nil
		responseBodyBytesStub = []byte(body)

		data, pageType, err := zebedeeClient.GetData("/", requestContextID)

		So(err, ShouldResemble, onsErrorStub)
		So(data, ShouldResemble, dataStub)
		So(pageType, ShouldEqual, pageTypeStub)
	})
}

func TestBuildRequest(t *testing.T) {
	// create stub http client for test
	testHTTPClient := &testClient{}
	zebedeeClient := Client{testHTTPClient, zebedeeURI}

	Convey("Should build the expected request for the given parameters.", t, func() {
		uriParameter := "/someURL"
		params := []parameter{{"name1", "value1"}}
		actual, err := zebedeeClient.buildGetRequest(uriParameter, requestContextID, params)

		So(err, ShouldBeEmpty)
		So(actual.URL.Path, ShouldEqual, zebedeeURI+uriParameter)
		So(actual.Method, ShouldEqual, "GET")
		So(actual.Header.Get("X-Request-Id"), ShouldResemble, requestContextID)
	})
}

func TestGetParents(t *testing.T) {
	testHTTPClient := &testClient{}
	zebedeeClient := Client{testHTTPClient, zebedeeURI}

	Convey("Should return parents for 200 response status & valid response body.", t, func() {
		// Set a mock for reading the response body.
		zebedeeClient.setResponseReader(ReadBodyMock)
		// Get some mock parents data.
		zebedeeParents := getZebedeeParents()

		// Set up the stub data for this test case.
		zebedeeBytes, _ := json.Marshal(zebedeeParents)

		responseBody := string(zebedeeBytes)
		pageTypeStub = HomePage
		onsErrorStub = nil
		responseStub = &http.Response{StatusCode: 200}
		responseStub.Header = make(map[string][]string, 0)
		responseStub.Header.Set(pageTypeHeader, HomePage)
		responseStub.Body = ioutil.NopCloser(bytes.NewBufferString(string(responseBody)))
		responseBodyReadErrStub = nil
		responseBodyBytesStub = zebedeeBytes

		result, err := zebedeeClient.GetParents("/", requestContextID)
		So(result, ShouldResemble, zebedeeParents)
		So(len(result), ShouldEqual, len(zebedeeParents))
		So(err, ShouldBeNil)
	})

	Convey("Should error if response status is not 200.", t, func() {
		// Set a mock for reading the response body.
		zebedeeClient.setResponseReader(ReadBodyMock)

		responseBody := "I am an error."
		var expectedParents []model.ContentNode

		onsErrorStub = errorWithReqContextID(errors.New("Unexpected Response status code"), incorrectStatusCodeErrDesc, requestContextID)
		onsErrorStub.AddParameter("expectedStatusCode", 200)
		onsErrorStub.AddParameter("actualStatusCode", 500)
		onsErrorStub.AddParameter("zebedeeURI", zebedeeURI+breadcrumbAPI)
		onsErrorStub.AddParameter(requestContextIDParam, requestContextID)

		pageTypeStub = HomePage
		responseStub = &http.Response{StatusCode: 500}
		responseStub.Header = make(map[string][]string, 0)
		responseStub.Header.Set(pageTypeHeader, HomePage)
		responseStub.Body = ioutil.NopCloser(bytes.NewBufferString(responseBody))
		responseBodyReadErrStub = nil
		responseBodyBytesStub = []byte(responseBody)

		result, err := zebedeeClient.GetParents("/", requestContextID)
		So(result, ShouldResemble, expectedParents)
		So(err, ShouldResemble, onsErrorStub)
	})

	Convey("Should error if response body is invalid.", t, func() {
		// Set a mock for reading the response body.
		zebedeeClient.setResponseReader(ReadBodyMock)

		responseBody := "I am an error."
		var expectedParents []model.ContentNode

		onsErrorStub = errorWithReqContextID(errors.New(""), "error unmarshalling zebedee contentNodes", requestContextID)
		pageTypeStub = HomePage
		responseStub = &http.Response{StatusCode: 200}
		responseStub.Header = make(map[string][]string, 0)
		responseStub.Header.Set(pageTypeHeader, HomePage)
		responseStub.Body = ioutil.NopCloser(bytes.NewBufferString(responseBody))
		responseBodyReadErrStub = nil
		responseBodyBytesStub = []byte(responseBody)

		result, err := zebedeeClient.GetParents("/", requestContextID)
		So(result, ShouldResemble, expectedParents)
		So(err.Parameters, ShouldResemble, onsErrorStub.Parameters)
	})
}

func getZebedeeParents() []model.ContentNode {
	child := model.ContentNode{URI: "/one"}
	parent := model.ContentNode{URI: "/", Children: []model.ContentNode{child}}
	return []model.ContentNode{parent}
}
