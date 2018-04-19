package hierarchy

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/ONSdigital/go-ns/clients/clientlog"
	"github.com/ONSdigital/go-ns/log"
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

// Code returns the status code received from hierarchy api if an error is returned
func (e ErrInvalidHierarchyAPIResponse) Code() int {
	return e.actualCode
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

// GetRoot returns the root hierarchy response from the hierarchy API
func (c *Client) GetRoot(instanceID, name string) (m Model, err error) {
	path := fmt.Sprintf("/hierarchies/%s/%s", instanceID, name)

	clientlog.Do("retrieving hierarchy", service, path, log.Data{
		"method":      "GET",
		"instance_id": instanceID,
		"dimension":   name,
	})

	return c.getHierarchy(path)
}

// GetChild returns a child of a given hierarchy and code
func (c *Client) GetChild(instanceID, name, code string) (m Model, err error) {
	path := fmt.Sprintf("/hierarchies/%s/%s/%s", instanceID, name, code)

	clientlog.Do("retrieving hierarchy", service, path, log.Data{
		"method":      "GET",
		"instance_id": instanceID,
		"dimension":   name,
		"code":        code,
	})

	return c.getHierarchy(path)
}

func (c *Client) getHierarchy(path string) (m Model, err error) {
	resp, err := c.cli.Get(c.url + path)
	if err != nil {
		return
	}

	if resp.StatusCode != http.StatusOK {
		return m, &ErrInvalidHierarchyAPIResponse{http.StatusOK, resp.StatusCode, path}
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	err = json.Unmarshal(b, &m)
	return
}
