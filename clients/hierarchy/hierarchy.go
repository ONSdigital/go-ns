package hierarchy

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/ONSdigital/go-ns/clients/clientlog"
	"github.com/ONSdigital/go-ns/rhttp"
)

const service = "hierarchy-api"

// ErrInvalidHierarchyAPIResponse is returned when the hierarchy api does not respond
// with a valid status
type ErrInvalidHierarchyAPIResponse struct {
	expectedCode int
	actualCode   int
	uri          string
}

// Error should be called by the user to print out the stringified version of the error
func (e ErrInvalidHierarchyAPIResponse) Error() string {
	return fmt.Sprintf("invalid response from hierarchy api - should be: %d, got: %d, path: %s",
		e.expectedCode,
		e.actualCode,
		e.uri,
	)
}

var _ error = ErrInvalidHierarchyAPIResponse{}

// Client is a hierarchy api client which can be used to make requests to the server
type Client struct {
	cli *rhttp.Client
	url string
}

// New creates a new instance of Client with a given filter api url
func New(hierarchyAPIURL string) *Client {
	return &Client{
		cli: rhttp.DefaultClient,
		url: hierarchyAPIURL,
	}
}

// Healthcheck calls the healthcheck endpoint on the api and alerts the caller of any errors
func (c *Client) Healthcheck() (string, error) {
	resp, err := c.cli.Get(c.url + "/healthcheck")
	if err != nil {
		return service, err
	}

	if resp.StatusCode != http.StatusOK {
		return service, &ErrInvalidHierarchyAPIResponse{http.StatusOK, resp.StatusCode, "/healthcheck"}
	}

	return service, nil
}

// GetHierarchy returns the hierarchy for the requested path
func (c *Client) GetHierarchy(path string) (m Model, err error) {
	uri := fmt.Sprintf("%s/hierarchies/%s", c.url, path)

	clientlog.Do("retrieving hierarchy", service, uri)

	resp, err := c.cli.Get(uri)
	if err != nil {
		return
	}

	if resp.StatusCode != http.StatusOK {
		err = &ErrInvalidHierarchyAPIResponse{http.StatusOK, resp.StatusCode, uri}
		return
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	err = json.Unmarshal(b, &m)
	return
}
