package zebedee

import (
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"encoding/json"

	"github.com/ONSdigital/go-ns/log"
	model "github.com/ONSdigital/go-ns/zebedee/model"
)

type zebedeeRequestError struct {
	err error
}

func (z zebedeeRequestError) Error() string {
	return z.err.Error()
}

var _ error = zebedeeRequestError{}

const uriParam = "uri"
const dataAPI = "/data"
const taxonomyAPI = "/taxonomy"
const breadcrumbAPI = "/parents"
const pageTypeHeader = "Ons-Page-Type"
const zebedeeGetError = "GET zebedee/data request returned an unexpected error."
const requestContextIDParam = "requestContextId"

var incorrectStatusCodeErrDesc = "Incorrect status code."
var ErrUnauthorised = errors.New("unauthorised user")

// httpClient provides only the methods of http.client that we are using allowing it to be mocked.
type httpClient interface {
	Get(url string) (resp *http.Response, err error)
	Do(req *http.Request) (*http.Response, error)
}

// Client holds the required fields to call Zebedee.
type Client struct {
	httpClient httpClient
	url        string
}

type parameter struct {
	name  string
	value string
}

// Hide read response body behind behind type to allow us to replace with stub during tests.
type responseBodyReader func(io.Reader) ([]byte, error)

var resReader responseBodyReader = ioutil.ReadAll

// CreateClient will create a new ZebedeeHTTPClient for the given url and timeout.
func CreateClient(timeout time.Duration, zebedeeURL string) *Client {
	return &Client{
		&http.Client{
			Timeout: timeout,
		},
		zebedeeURL}
}

// GetData will call Zebedee and return the data it provides in a []byte
func (zebedee *Client) GetData(uri string, requestContextID string) (data []byte, pageType string, err error) {
	var response *http.Response

	request, error := zebedee.buildGetRequest(dataAPI, requestContextID, []parameter{{name: uriParam, value: uri}})
	if err != nil {
		return data, pageType, zebedeeRequestError{err}
	}

	response, error = zebedee.httpClient.Do(request)

	if error != nil {
		return data, pageType, zebedeeRequestError{err}
	}

	if response.StatusCode != 200 {
		if response.StatusCode == 401 {
			return nil, "", errors.New("Unauthorised")
		}
		// onsErr := errorWithReqContextID(errors.New("Unexpected Response status code"), incorrectStatusCodeErrDesc, requestContextID)
		// onsErr.AddParameter("zebedeeURI", request.URL.Path)
		// onsErr.AddParameter("expectedStatusCode", 200)
		// onsErr.AddParameter("actualStatusCode", response.StatusCode)
		// onsErr.AddParameter("query", request.URL.Query().Get("uri"))
		return data, pageType, errors.New("Unexpected response code")
	}

	data, error = resReader(response.Body)
	defer response.Body.Close()

	if error != nil {
		return data, pageType, errors.New("error reading response body") //errorWithReqContextID(error, "error reading response body", requestContextID)
	}

	pageType = response.Header.Get(pageTypeHeader)
	log.Debug("Identified page type", log.Data{"page type": pageType})
	return
}

// GetTaxonomy gets the taxonomy structure of the website from Zebedee
func (zebedee *Client) GetTaxonomy(uri string, depth int, requestContextID string) ([]model.ContentNode, error) {
	var zebedeeContentNodeList []model.ContentNode
	params := []parameter{
		{name: uriParam, value: uri},
		{name: "depth", value: strconv.Itoa(depth)},
	}
	zebedeeBytes, err := zebedee.get(taxonomyAPI, requestContextID, params)

	if err != nil {
		return zebedeeContentNodeList, err
	}

	unmarshallErr := json.Unmarshal(zebedeeBytes, &zebedeeContentNodeList)
	if unmarshallErr != nil {
		return zebedeeContentNodeList, errors.New("error unmarshaling content taxonomy nodes") //errorWithReqContextID(unmarshallErr, "Error while attempting to unmarshal content taxonomy nodes.", requestContextID)
	}
	return zebedeeContentNodeList, nil
}

// GetParents gets the breadcrumb for the given url.
func (zebedee *Client) GetParents(uri string, requestContextID string) ([]model.ContentNode, error) {
	var zebedeeContentNodes []model.ContentNode
	zebedeeBytes, err := zebedee.get(breadcrumbAPI, requestContextID, []parameter{{name: uriParam, value: uri}})

	if err != nil {
		return zebedeeContentNodes, err
	}

	unmarshallErr := json.Unmarshal(zebedeeBytes, &zebedeeContentNodes)
	if unmarshallErr != nil {
		return zebedeeContentNodes, errors.New("error unmarshaling zebedee content nodes") // errorWithReqContextID(err, "error unmarshalling zebedee contentNodes", requestContextID)
	}
	return zebedeeContentNodes, nil
}

// GetTimeSeries - get timeseries data.json from Zebedee.
func (zebedee *Client) GetTimeSeries(uri string, requestContextID string) (*model.TimeseriesPage, error) {
	params := []parameter{{name: uriParam, value: uri}, {name: "series"}}
	zebedeeBytes, err := zebedee.get(dataAPI, requestContextID, params)

	if err != nil {
		return nil, err
	}

	var timeSeriesPage *model.TimeseriesPage
	unmarshalErr := json.Unmarshal(zebedeeBytes, &timeSeriesPage)
	if unmarshalErr != nil {
		return nil, errors.New("error unmarshaling timeseries page") // errorWithReqContextID(unmarshalErr, "Error unmarshalling timeseries pages json.", requestContextID)
	}
	return timeSeriesPage, nil
}

// Perform a HTTP GET request to zebedee for the specified uri & parameters.
func (zebedee *Client) get(path string, requestContextID string, params []parameter) ([]byte, error) {
	request, err := zebedee.buildGetRequest(path, requestContextID, params)
	if err != nil {
		return nil, errors.New("error creating zebedee request") // errorWithReqContextID(err, "error creating zebedee request", requestContextID)
	}

	log.Debug("Zebedee Client HTTP GET", log.Data{
		"uri":                 request.URL.Path,
		"method":              "GET",
		requestContextIDParam: requestContextID,
		"query":               request.URL.RawQuery,
	})
	response, err := zebedee.httpClient.Do(request)
	defer response.Body.Close()

	if err != nil {
		return nil, errors.New("error calling zebedee") //errorWithReqContextID(err, "error performing zebedee request", requestContextID)
	}

	if response.StatusCode != 200 {
		// onsError := errorWithReqContextID(errors.New("Unexpected Response status code"), incorrectStatusCodeErrDesc, requestContextID)
		// onsError.AddParameter("expectedStatusCode", 200)
		// onsError.AddParameter("actualStatusCode", response.StatusCode)
		// onsError.AddParameter("zebedeeURI", request.URL.Path)
		// onsError.AddParameter(requestContextIDParam, requestContextID)
		return nil, errors.New("unexpected response code")
	}

	body, err := resReader(response.Body)
	if err != nil {
		return nil, errors.New("error reading response body") // errorWithReqContextID(err, "error reading zebedee response body", requestContextID)
	}
	return body, nil
}

// buildGetRequest builds a new http GET Request using the uri and parameters provided and adds the request context Id as
// a header to the new request.
func (zebedee *Client) buildGetRequest(url string, requestContextID string, params []parameter) (*http.Request, error) {
	request, err := http.NewRequest("GET", zebedee.url+url, nil)
	if err != nil {
		return nil, err
	}

	request.Header.Add("X-Request-Id", requestContextID)

	if len(params) > 0 {
		query := request.URL.Query()
		for _, param := range params {
			query.Add(param.name, param.value)
		}
		request.URL.RawQuery = query.Encode()
	}
	return request, nil
}

func (zebedee *Client) setResponseReader(f func(io.Reader) ([]byte, error)) {
	resReader = f
}
