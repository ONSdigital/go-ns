package dataset

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"

	"github.com/ONSdigital/go-ns/clients/clientlog"
	"github.com/ONSdigital/go-ns/common"
	"github.com/ONSdigital/go-ns/rchttp"
	"github.com/pkg/errors"
)

const service = "dataset-api"

// ErrInvalidDatasetAPIResponse is returned when the dataset api does not respond
// with a valid status
type ErrInvalidDatasetAPIResponse struct {
	actualCode int
	uri        string
	body       string
}

// Error should be called by the user to print out the stringified version of the error
func (e ErrInvalidDatasetAPIResponse) Error() string {
	return fmt.Sprintf("invalid response: %d from dataset api: %s, body: %s",
		e.actualCode,
		e.uri,
		e.body,
	)
}

// Code returns the status code received from dataset api if an error is returned
func (e ErrInvalidDatasetAPIResponse) Code() int {
	return e.actualCode
}

var _ error = ErrInvalidDatasetAPIResponse{}

// Client is a dataset api client which can be used to make requests to the server
type Client struct {
	cli common.RCHTTPClienter
	url string
}

// NewAPIClient creates a new instance of Client with a given dataset api url and the relevant tokens
func NewAPIClient(datasetAPIURL, serviceToken, xDownloadServiceToken string) *Client {
	return &Client{
		cli: rchttp.ClientWithDownloadServiceToken(
			rchttp.ClientWithServiceToken(nil, serviceToken),
			xDownloadServiceToken,
		),
		url: datasetAPIURL,
	}
}

// Healthcheck calls the healthcheck endpoint on the api and alerts the caller of any errors
func (c *Client) Healthcheck() (string, error) {
	ctx := context.Background()

	resp, err := c.cli.Get(ctx, c.url+"/healthcheck")
	if err != nil {
		return service, err
	}

	if resp.StatusCode != http.StatusOK {
		return service, NewDatasetAPIResponse(resp, "/healthcheck")
	}

	return service, nil
}

// Get returns dataset level information for a given dataset id
func (c *Client) Get(ctx context.Context, id string) (m Model, err error) {
	uri := fmt.Sprintf("%s/datasets/%s", c.url, id)

	clientlog.Do(ctx, "retrieving dataset", service, uri)

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return
	}

	req = setCollectionID(ctx, req)

	resp, err := c.cli.Do(ctx, req)
	if err != nil {
		return
	}

	if resp.StatusCode != http.StatusOK {
		err = NewDatasetAPIResponse(resp, uri)
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
	if next, ok := body["next"]; ok && (common.IsCallerPresent(ctx) || common.IsFlorenceIdentityPresent(ctx)) {
		b, err = json.Marshal(next)
		if err != nil {
			return
		}
	}

	err = json.Unmarshal(b, &m)
	return
}

// GetDatasets returns the list of datasets
func (c *Client) GetDatasets(ctx context.Context) (m ModelCollection, err error) {
	uri := fmt.Sprintf("%s/datasets", c.url)

	clientlog.Do(ctx, "retrieving datasets", service, uri)

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return
	}

	req = setCollectionID(ctx, req)

	resp, err := c.cli.Do(ctx, req)
	if err != nil {
		return
	}

	if resp.StatusCode != http.StatusOK {
		err = NewDatasetAPIResponse(resp, uri)
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

	return
}

// GetEdition retrieves a single edition document from a given datasetID and edition label
func (c *Client) GetEdition(ctx context.Context, datasetID, edition string) (m Edition, err error) {
	uri := fmt.Sprintf("%s/datasets/%s/editions/%s", c.url, datasetID, edition)

	clientlog.Do(ctx, "retrieving dataset editions", service, uri)

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return
	}

	req = setCollectionID(ctx, req)

	resp, err := c.cli.Do(ctx, req)
	if err != nil {
		return
	}

	if resp.StatusCode != http.StatusOK {
		err = NewDatasetAPIResponse(resp, uri)
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

	if next, ok := body["next"]; ok && common.IsCallerPresent(ctx) {
		b, err = json.Marshal(next)
		if err != nil {
			return
		}
	}

	err = json.Unmarshal(b, &m)
	return
}

// GetEditions returns all editions for a dataset
func (c *Client) GetEditions(ctx context.Context, id string) (m []Edition, err error) {
	uri := fmt.Sprintf("%s/datasets/%s/editions", c.url, id)

	clientlog.Do(ctx, "retrieving dataset editions", service, uri)

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return
	}

	req = setCollectionID(ctx, req)

	resp, err := c.cli.Do(ctx, req)
	if err != nil {
		return
	}

	if resp.StatusCode != http.StatusOK {
		err = NewDatasetAPIResponse(resp, uri)
		return
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var body map[string]interface{}
	if err = json.Unmarshal(b, &body); err != nil {
		return nil, nil
	}

	if _, ok := body["items"].([]interface{})[0].(map[string]interface{})["next"]; ok && common.IsCallerPresent(ctx) {
		var items []map[string]interface{}
		for _, item := range body["items"].([]interface{}) {
			items = append(items, item.(map[string]interface{})["next"].(map[string]interface{}))
		}
		parentItems := make(map[string]interface{})
		parentItems["items"] = items
		b, err = json.Marshal(parentItems)
		if err != nil {
			return
		}
	}

	editions := struct {
		Items []Edition `json:"items"`
	}{}
	err = json.Unmarshal(b, &editions)
	m = editions.Items
	return
}

// GetVersions gets all versions for an edition from the dataset api
func (c *Client) GetVersions(ctx context.Context, id, edition string) (m []Version, err error) {
	uri := fmt.Sprintf("%s/datasets/%s/editions/%s/versions", c.url, id, edition)

	clientlog.Do(ctx, "retrieving dataset versions", service, uri)

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return
	}

	req = setCollectionID(ctx, req)

	resp, err := c.cli.Do(ctx, req)
	if err != nil {
		return
	}

	if resp.StatusCode != http.StatusOK {
		err = NewDatasetAPIResponse(resp, uri)
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
func (c *Client) GetVersion(ctx context.Context, id, edition, version string) (m Version, err error) {
	uri := fmt.Sprintf("%s/datasets/%s/editions/%s/versions/%s", c.url, id, edition, version)

	clientlog.Do(ctx, "retrieving dataset version", service, uri)

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return
	}

	req = setCollectionID(ctx, req)

	resp, err := c.cli.Do(ctx, req)
	if err != nil {
		return
	}

	if resp.StatusCode != http.StatusOK {
		err = NewDatasetAPIResponse(resp, uri)
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

// GetInstance returns an instance from the dataset api
func (c *Client) GetInstance(ctx context.Context, instanceID string) (m Instance, err error) {
	uri := fmt.Sprintf("%s/instances/%s", c.url, instanceID)

	clientlog.Do(ctx, "retrieving dataset version", service, uri)

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return
	}

	req = setCollectionID(ctx, req)

	resp, err := c.cli.Do(ctx, req)
	if err != nil {
		return
	}

	if resp.StatusCode != http.StatusOK {
		err = NewDatasetAPIResponse(resp, uri)
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
func (c *Client) PutVersion(ctx context.Context, datasetID, edition, version string, v Version) error {
	uri := fmt.Sprintf("%s/datasets/%s/editions/%s/versions/%s", c.url, datasetID, edition, version)
	clientlog.Do(ctx, "updating version", service, uri)

	b, err := json.Marshal(v)
	if err != nil {
		return errors.Wrap(err, "error while attempting to marshall version")
	}

	req, err := http.NewRequest(http.MethodPut, uri, bytes.NewBuffer(b))
	if err != nil {
		return errors.Wrap(err, "error while attempting to create http request")
	}

	req = setCollectionID(ctx, req)

	resp, err := c.cli.Do(ctx, req)
	if err != nil {
		return errors.Wrap(err, "http client returned error while attempting to make request")
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("incorrect http status, expected: 200, actual: %d, uri: %s", resp.StatusCode, uri)
	}
	return nil
}

// GetMetadataURL returns the URL for the metadata of a given dataset id, edition and version
func (c *Client) GetMetadataURL(id, edition, version string) string {
	return fmt.Sprintf("%s/datasets/%s/editions/%s/versions/%s/metadata", c.url, id, edition, version)
}

// GetVersionMetadata returns the metadata for a given dataset id, edition and version
func (c *Client) GetVersionMetadata(ctx context.Context, id, edition, version string) (m Metadata, err error) {
	uri := c.GetMetadataURL(id, edition, version)

	clientlog.Do(ctx, "retrieving dataset version metadata", service, uri)

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return
	}

	req = setCollectionID(ctx, req)

	resp, err := c.cli.Do(ctx, req)
	if err != nil {
		return
	}

	if resp.StatusCode != http.StatusOK {
		err = NewDatasetAPIResponse(resp, uri)
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
func (c *Client) GetDimensions(ctx context.Context, id, edition, version string) (m Dimensions, err error) {
	uri := fmt.Sprintf("%s/datasets/%s/editions/%s/versions/%s/dimensions", c.url, id, edition, version)

	clientlog.Do(ctx, "retrieving dataset version dimensions", service, uri)

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return
	}

	req = setCollectionID(ctx, req)

	resp, err := c.cli.Do(ctx, req)
	if err != nil {
		return
	}

	if resp.StatusCode != http.StatusOK {
		err = NewDatasetAPIResponse(resp, uri)
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
func (c *Client) GetOptions(ctx context.Context, id, edition, version, dimension string) (m Options, err error) {
	uri := fmt.Sprintf("%s/datasets/%s/editions/%s/versions/%s/dimensions/%s/options", c.url, id, edition, version, dimension)

	clientlog.Do(ctx, "retrieving options for dimension", service, uri)

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return
	}

	req = setCollectionID(ctx, req)

	resp, err := c.cli.Do(ctx, req)
	if err != nil {
		return
	}

	if resp.StatusCode != http.StatusOK {
		err = NewDatasetAPIResponse(resp, uri)
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

// NewDatasetAPIResponse creates an error response, optionally adding body to e when status is 404
func NewDatasetAPIResponse(resp *http.Response, uri string) (e *ErrInvalidDatasetAPIResponse) {
	e = &ErrInvalidDatasetAPIResponse{
		actualCode: resp.StatusCode,
		uri:        uri,
	}
	if resp.StatusCode == http.StatusNotFound {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			e.body = "Client failed to read DatasetAPI body"
			return
		}
		defer resp.Body.Close()
		e.body = string(b)
	}
	return
}

func setCollectionID(ctx context.Context, req *http.Request) *http.Request {

	rawKeyValue := ctx.Value(common.CollectionIDHeaderKey)

	if rawKeyValue != nil {  // avoid stringifying an empty interface
		collectionID := rawKeyValue.(string)
		req.Header.Set(common.CollectionIDHeaderKey, collectionID)
	}

	return req
}
