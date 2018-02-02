package dataset

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"

	"bytes"

	"io"
	"net/url"

	"github.com/ONSdigital/go-ns/clients/clientlog"
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/go-ns/rhttp"
	"github.com/pkg/errors"
)

const (
	service         = "dataset-api"
	authTokenHeader = "Internal-Token"
)

//go:generate moq -out dataset_mocks/mocks.go -pkg dataset_mocks . RHTTPClient

// ErrInvalidDatasetAPIResponse is returned when the dataset api does not respond
// with a valid status
type ErrInvalidDatasetAPIResponse struct {
	expectedCode int
	actualCode   int
	uri          string
}

// RHTTPClient retry http client.
type RHTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
	Get(url string) (*http.Response, error)
	Head(url string) (*http.Response, error)
	Post(url string, contentType string, body io.Reader) (*http.Response, error)
	PostForm(uri string, data url.Values) (*http.Response, error)
}

// Error should be called by the user to print out the stringified version of the error
func (e ErrInvalidDatasetAPIResponse) Error() string {
	return fmt.Sprintf("invalid response from dataset api - should be: %d, got: %d, path: %s",
		e.expectedCode,
		e.actualCode,
		e.uri,
	)
}

// Code returns the status code received from dataset api if an error is returned
func (e ErrInvalidDatasetAPIResponse) Code() int {
	return e.actualCode
}

var _ error = ErrInvalidDatasetAPIResponse{}

// Client is a dataset api client which can be used to make requests to the server
type Client struct {
	cli           RHTTPClient
	url           string
	internalToken string
}

// New creates a new instance of Client with a given filter api url
func New(datasetAPIURL string) *Client {
	return &Client{
		cli: rhttp.DefaultClient,
		url: datasetAPIURL,
	}
}

// SetInternalToken will set an internal token to use for the dataset api
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
		return service, &ErrInvalidDatasetAPIResponse{http.StatusOK, resp.StatusCode, "/healthcheck"}
	}

	return service, nil
}

// Get returns dataset level information for a given dataset id
func (c *Client) Get(id string) (m Model, err error) {
	uri := fmt.Sprintf("%s/datasets/%s", c.url, id)

	clientlog.Do("retrieving dataset", service, uri)

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return
	}
	c.setInternalTokenHeader(req)

	resp, err := c.cli.Do(req)
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

	var body map[string]interface{}
	if err = json.Unmarshal(b, &body); err != nil {
		return
	}

	// TODO: Authentication will sort this problem out for us. Currently
	// the shape of the response body is different if you are authenticated
	// so return the "next" item only
	if next, ok := body["next"]; ok && len(c.internalToken) > 0 {
		b, err = json.Marshal(next)
		if err != nil {
			return
		}
	}

	err = json.Unmarshal(b, &m)
	return
}

// GetEdition retrieves a single edition document from a given datasetID and edition label
func (c *Client) GetEdition(datasetID, edition string) (m Edition, err error) {
	uri := fmt.Sprintf("%s/datasets/%s/editions/%s", c.url, datasetID, edition)

	clientlog.Do("retrieving dataset editions", service, uri)

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return
	}
	c.setInternalTokenHeader(req)

	resp, err := c.cli.Do(req)
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

// GetEditions returns all editions for a dataset
func (c *Client) GetEditions(id string) (m []Edition, err error) {
	uri := fmt.Sprintf("%s/datasets/%s/editions", c.url, id)

	clientlog.Do("retrieving dataset editions", service, uri)

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return
	}
	c.setInternalTokenHeader(req)

	resp, err := c.cli.Do(req)
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

	editions := struct {
		Items []Edition `json:"items"`
	}{}
	err = json.Unmarshal(b, &editions)
	m = editions.Items
	return
}

// GetVersions gets all versions for an edition from the dataset api
func (c *Client) GetVersions(id, edition string) (m []Version, err error) {
	uri := fmt.Sprintf("%s/datasets/%s/editions/%s/versions", c.url, id, edition)

	clientlog.Do("retrieving dataset versions", service, uri)

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return
	}
	c.setInternalTokenHeader(req)

	resp, err := c.cli.Do(req)
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

	versions := struct {
		Items []Version `json:"items"`
	}{}

	err = json.Unmarshal(b, &versions)
	m = versions.Items
	return
}

// GetVersion gets a specific version for an edition from the dataset api
func (c *Client) GetVersion(id, edition, version string) (m Version, err error) {
	uri := fmt.Sprintf("%s/datasets/%s/editions/%s/versions/%s", c.url, id, edition, version)

	clientlog.Do("retrieving dataset version", service, uri)

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return
	}
	c.setInternalTokenHeader(req)

	resp, err := c.cli.Do(req)
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

// PutVersion update the version
func (c *Client) PutVersion(datasetID, edition, version string, v Version) error {
	uri := fmt.Sprintf("%s/datasets/%s/editions/%s/versions/%s", c.url, datasetID, edition, version)
	clientlog.Do("updating version", service, uri)

	b, err := json.Marshal(v)
	if err != nil {
		return errors.Wrap(err, "error while attempting to marshall version")
	}

	req, err := http.NewRequest(http.MethodPut, uri, bytes.NewBuffer(b))
	if err != nil {
		return errors.Wrap(err, "error while attempting to create http request")
	}

	if c.internalToken == "" {
		log.Info("no authentication provided, sending request without auth token header", nil)
	} else {
		log.Info("applying configured request authentication token header", nil)
		req.Header.Set(authTokenHeader, c.internalToken)
	}

	resp, err := c.cli.Do(req)
	if err != nil {
		return errors.Wrap(err, "http client returned error while attempting to make request")
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("incorrect http status, expected: 200, actual: %d, uri: %s", resp.StatusCode, uri)
	}
	return nil
}

// GetVersionMetadata returns the metadata for a given dataset id, edition and version
func (c *Client) GetVersionMetadata(id, edition, version string) (m Metadata, err error) {
	uri := fmt.Sprintf("%s/datasets/%s/editions/%s/versions/%s/metadata", c.url, id, edition, version)

	clientlog.Do("retrieving dataset version metadata", service, uri)

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return
	}
	c.setInternalTokenHeader(req)

	resp, err := c.cli.Do(req)
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

// GetDimensions will return a versions dimensions
func (c *Client) GetDimensions(id, edition, version string) (m Dimensions, err error) {
	uri := fmt.Sprintf("%s/datasets/%s/editions/%s/versions/%s/dimensions", c.url, id, edition, version)

	clientlog.Do("retrieving dataset version dimensions", service, uri)

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return
	}
	c.setInternalTokenHeader(req)

	resp, err := c.cli.Do(req)
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

	if err = json.Unmarshal(b, &m); err != nil {
		return
	}

	sort.Sort(m.Items)

	return
}

// GetOptions will return the options for a dimension
func (c *Client) GetOptions(id, edition, version, dimension string) (m Options, err error) {
	uri := fmt.Sprintf("%s/datasets/%s/editions/%s/versions/%s/dimensions/%s/options", c.url, id, edition, version, dimension)

	clientlog.Do("retrieving options for dimension", service, uri)

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return
	}
	c.setInternalTokenHeader(req)

	resp, err := c.cli.Do(req)
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
