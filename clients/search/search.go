package search

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ONSdigital/go-ns/clients/clientlog"
	"github.com/ONSdigital/go-ns/rhttp"
)

const (
	service = "search-api"
	limit   = 20
	offset  = 0
)

// HTTPClient provides an interface for methods on an HTTP Client
type HTTPClient interface {
	Get(url string) (*http.Response, error)
	Do(req *http.Request) (*http.Response, error)
}

// ErrInvalidSearchAPIResponse is returned when the search api does not respond
// with a valid status
type ErrInvalidSearchAPIResponse struct {
	expectedCode int
	actualCode   int
	uri          string
}

// Error should be called by the user to print out the stringified version of the error
func (e ErrInvalidSearchAPIResponse) Error() string {
	return fmt.Sprintf("invalid response from search api - should be: %d, got: %d, path: %s",
		e.expectedCode,
		e.actualCode,
		e.uri,
	)
}

// Code returns the status code received from search api if an error is returned
func (e ErrInvalidSearchAPIResponse) Code() int {
	return e.actualCode
}

var _ error = ErrInvalidSearchAPIResponse{}

// Client is a search api client that can be used to make requests to the server
type Client struct {
	cli           HTTPClient
	url           string
	internalToken string
}

// New creates a new instance of Client with a given search api url
func New(searchAPIURL string) *Client {
	return &Client{
		cli: rhttp.DefaultClient,
		url: searchAPIURL,
	}
}

// SetInternalToken will set an internal token to use for the search api
func (c *Client) SetInternalToken(token string) {
	c.internalToken = token
}

func (c *Client) setInternalTokenHeader(req *http.Request) {
	if len(c.internalToken) > 0 {
		req.Header.Set("Internal-token", c.internalToken)
	}
}

// Healthcheck calls the healthcheck endpoint on the api and alerts the caller of any errors
func (c *Client) Healthcheck() (string, error) {
	resp, err := c.cli.Get(c.url + "/healthcheck")
	if err != nil {
		return service, err
	}

	if resp.StatusCode != http.StatusOK {
		return service, &ErrInvalidSearchAPIResponse{http.StatusOK, resp.StatusCode, c.url + "/healthcheck"}
	}

	return service, nil
}

// Dimension allows the searching of a dimension for a specific dimension option, optionally
// pass in offset and limit parameters respectively
func (c *Client) Dimension(datasetID, edition, version, name, query string, params ...int) (m *Model, err error) {
	offset := offset
	limit := limit

	if len(params) > 0 {
		offset = params[0]
	}
	if len(params) > 1 {
		limit = params[1]
	}

	uri := fmt.Sprintf("%s/search/datasets/%s/editions/%s/versions/%s/dimensions/%s?q=%s&limit=%d&offset=%d",
		c.url,
		datasetID,
		edition,
		version,
		name,
		query,
		limit,
		offset,
	)

	clientlog.Do("searching for dataset dimension option", service, uri)

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return
	}
	c.setInternalTokenHeader(req)

	resp, err := c.cli.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, &ErrInvalidSearchAPIResponse{http.StatusOK, resp.StatusCode, uri}
	}

	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&m)

	return
}
