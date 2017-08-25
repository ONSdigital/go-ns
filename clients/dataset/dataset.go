package dataset

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// ErrInvalidDatasetAPIResponse is returned when the dataset api does not respond
// with a valid status
type ErrInvalidDatasetAPIResponse struct {
	expectedCode int
	actualCode   int
	uri          string
}

func (e ErrInvalidDatasetAPIResponse) Error() string {
	return fmt.Sprintf("invalid response from dataset api - should be: %d, got: %d, path: %s",
		e.expectedCode,
		e.actualCode,
		e.uri,
	)
}

var _ error = ErrInvalidDatasetAPIResponse{}

// Client is a dataset api client which can be used to make requests to the server
type Client struct {
	cli *http.Client
	url string
}

// New creates a new instance of Client with a given filter api url
func New(datasetAPIURL string) *Client {
	return &Client{
		cli: &http.Client{Timeout: 5 * time.Second},
		url: datasetAPIURL,
	}
}

// Healthcheck calls the healthcheck endpoint on the api and alerts the caller of any errors
func (c *Client) Healthcheck() (string, error) {
	resp, err := c.cli.Get(c.url + "/healthcheck")
	if err != nil {
		return "dataset-api", err
	}

	if resp.StatusCode != http.StatusOK {
		return "dataset-api", &ErrInvalidDatasetAPIResponse{http.StatusOK, resp.StatusCode, "/healthcheck"}
	}

	return "", nil
}

// GetDataset ...
func (c *Client) GetDataset(id, edition, version string) (m Model, err error) {
	uri := fmt.Sprintf("%s/datasets/%s/editions/%s/versions/%s", c.url, id, edition, version)
	resp, err := c.cli.Get(uri)
	if err != nil {
		return
	}

	if resp.StatusCode != http.StatusOK {
		err = &ErrInvalidDatasetAPIResponse{http.StatusOK, resp.StatusCode, uri}
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
