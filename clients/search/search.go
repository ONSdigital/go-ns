package search

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/ONSdigital/go-ns/clients/clientlog"
	"github.com/ONSdigital/go-ns/common"
	"github.com/ONSdigital/go-ns/rchttp"
)

const (
	service       = "search-api"
	defaultLimit  = 20
	defaultOffset = 0
)

// Config represents configuration required to conduct a search request
type Config struct {
	Limit         *int
	Offset        *int
	InternalToken string
	FlorenceToken string
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
	cli common.RCHTTPClienter
	url string
}

// New creates a new instance of Client with a given search api url
func New(searchAPIURL string) *Client {
	return &Client{
		cli: rchttp.NewClient(),
		url: searchAPIURL,
	}
}

// Healthcheck calls the healthcheck endpoint on the api and alerts the caller of any errors
func (c *Client) Healthcheck() (string, error) {
	resp, err := c.cli.Get(context.Background(), c.url+"/healthcheck")
	if err != nil {
		return service, err
	}

	if resp.StatusCode != http.StatusOK {
		return service, &ErrInvalidSearchAPIResponse{http.StatusOK, resp.StatusCode, c.url + "/healthcheck"}
	}

	return service, nil
}

// Dimension allows the searching of a dimension for a specific dimension option, optionally
// pass in configuration parameters as an additional field. This can include a request specific
// internal token
func (c *Client) Dimension(ctx context.Context, datasetID, edition, version, name, query string, params ...Config) (m *Model, err error) {
	offset := defaultOffset
	limit := defaultLimit

	if len(params) > 0 {
		if params[0].Offset != nil {
			offset = *params[0].Offset
		}
		if params[0].Limit != nil {
			limit = *params[0].Limit
		}
	}

	uri := fmt.Sprintf("%s/search/datasets/%s/editions/%s/versions/%s/dimensions/%s?",
		c.url,
		datasetID,
		edition,
		version,
		name,
	)

	v := url.Values{}
	v.Add("q", query)
	v.Add("limit", strconv.Itoa(limit))
	v.Add("offset", strconv.Itoa(offset))

	uri = uri + v.Encode()

	clientlog.Do(ctx, "searching for dataset dimension option", service, uri)

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return
	}

	if len(params) > 0 {
		if len(params[0].InternalToken) > 0 {
			req.Header.Set(common.DeprecatedAuthHeader, params[0].InternalToken)
		}
		if len(params[0].FlorenceToken) > 0 {
			req.Header.Set(common.FlorenceHeaderKey, params[0].FlorenceToken)
		}
	}

	resp, err := c.cli.Do(ctx, req)
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
