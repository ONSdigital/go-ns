package filter

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/ONSdigital/go-ns/clients/clientlog"
	"github.com/ONSdigital/go-ns/common"
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/go-ns/rchttp"
)

const service = "filter-api"

// ErrInvalidFilterAPIResponse is returned when the filter api does not respond
// with a valid status
type ErrInvalidFilterAPIResponse struct {
	expectedCode int
	actualCode   int
	uri          string
}

// Config contains any configuration required to send requests to the filter api
type Config struct {
	InternalToken string
	FlorenceToken string
}

// Error should be called by the user to print out the stringified version of the error
func (e ErrInvalidFilterAPIResponse) Error() string {
	return fmt.Sprintf("invalid response from filter api - should be: %d, got: %d, path: %s",
		e.expectedCode,
		e.actualCode,
		e.uri,
	)
}

// Code returns the status code received from filter api if an error is returned
func (e ErrInvalidFilterAPIResponse) Code() int {
	return e.actualCode
}

var _ error = ErrInvalidFilterAPIResponse{}

// Client is a filter api client which can be used to make requests to the server
type Client struct {
	cli common.RCHTTPClienter
	url string
}

// New creates a new instance of Client with a given filter api url
func New(filterAPIURL, serviceToken, xDownloadServiceToken string) *Client {
	return &Client{
		cli: rchttp.ClientWithDownloadServiceToken(
			rchttp.ClientWithServiceToken(nil, serviceToken),
			xDownloadServiceToken,
		),
		url: filterAPIURL,
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
		return service, &ErrInvalidFilterAPIResponse{http.StatusOK, resp.StatusCode, "/healthcheck"}
	}

	return service, nil
}

// GetOutput returns a filter output job for a given filter output id
func (c *Client) GetOutput(ctx context.Context, filterOutputID string) (m Model, err error) {
	uri := fmt.Sprintf("%s/filter-outputs/%s", c.url, filterOutputID)

	clientlog.Do(ctx, "retrieving filter output", service, uri)

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return
	}

	resp, err := c.cli.Do(ctx, req)
	if err != nil {
		return
	}

	if resp.StatusCode != http.StatusOK {
		err = &ErrInvalidFilterAPIResponse{http.StatusOK, resp.StatusCode, uri}
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

// GetDimension returns information on a requested dimension name for a given filterID
func (c *Client) GetDimension(ctx context.Context, filterID, name string) (dim Dimension, err error) {
	uri := fmt.Sprintf("%s/filters/%s/dimensions/%s", c.url, filterID, name)

	clientlog.Do(ctx, "retrieving dimension information", service, uri)

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return
	}

	resp, err := c.cli.Do(ctx, req)
	if err != nil {
		return
	}

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode != http.StatusNoContent {
			err = &ErrInvalidFilterAPIResponse{http.StatusOK, resp.StatusCode, uri}
		}
		return
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if err = json.Unmarshal(b, &dim); err != nil {
		return
	}

	return
}

// GetDimensions will return the dimensions associated with the provided filter id
func (c *Client) GetDimensions(ctx context.Context, filterID string) (dims []Dimension, err error) {
	uri := fmt.Sprintf("%s/filters/%s/dimensions", c.url, filterID)

	clientlog.Do(ctx, "retrieving all dimensions for given filter job", service, uri)

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return
	}

	resp, err := c.cli.Do(ctx, req)
	if err != nil {
		return
	}

	if resp.StatusCode != http.StatusOK {
		err = &ErrInvalidFilterAPIResponse{http.StatusOK, resp.StatusCode, uri}
		return
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	err = json.Unmarshal(b, &dims)
	return
}

// GetDimensionOptions retrieves a list of the dimension options
func (c *Client) GetDimensionOptions(ctx context.Context, filterID, name string) (opts []DimensionOption, err error) {
	uri := fmt.Sprintf("%s/filters/%s/dimensions/%s/options", c.url, filterID, name)

	clientlog.Do(ctx, "retrieving selected dimension options for filter job", service, uri)

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return
	}

	resp, err := c.cli.Do(ctx, req)
	if err != nil {
		return
	}

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode != http.StatusNoContent {
			err = &ErrInvalidFilterAPIResponse{http.StatusOK, resp.StatusCode, uri}
		}
		return
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	err = json.Unmarshal(b, &opts)
	return
}

// CreateBlueprint creates a filter blueprint and returns the associated filterID
func (c *Client) CreateBlueprint(ctx context.Context, datasetID, edition, version string, names []string) (string, error) {
	ver, err := strconv.Atoi(version)
	if err != nil {
		return "", err
	}

	cb := CreateBlueprint{Dataset: Dataset{DatasetID: datasetID, Edition: edition, Version: ver}}

	var dimensions []ModelDimension
	for _, name := range names {
		dimensions = append(dimensions, ModelDimension{Name: name})
	}

	cb.Dimensions = dimensions

	b, err := json.Marshal(cb)
	if err != nil {
		return "", err
	}

	uri := c.url + "/filters"
	clientlog.Do(ctx, "attemping to create filter blueprint", service, uri, log.Data{
		"method":    "POST",
		"datasetID": datasetID,
		"edition":   edition,
		"version":   version,
	})

	req, err := http.NewRequest("POST", uri, bytes.NewBuffer(b))
	if err != nil {
		return "", err
	}

	resp, err := c.cli.Do(ctx, req)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusCreated {
		return "", ErrInvalidFilterAPIResponse{http.StatusCreated, resp.StatusCode, uri}
	}
	defer resp.Body.Close()

	b, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if err = json.Unmarshal(b, &cb); err != nil {
		return "", err
	}

	return cb.FilterID, nil
}

// UpdateBlueprint will update a blueprint with a given filter model
func (c *Client) UpdateBlueprint(ctx context.Context, m Model, doSubmit bool) (mdl Model, err error) {
	b, err := json.Marshal(m)
	if err != nil {
		return
	}

	uri := fmt.Sprintf("%s/filters/%s", c.url, m.FilterID)

	if doSubmit {
		uri = uri + "?submitted=true"
	}

	clientlog.Do(ctx, "updating filter job", service, uri, log.Data{
		"method": "PUT",
		"body":   string(b),
	})

	req, err := http.NewRequest("PUT", uri, bytes.NewBuffer(b))
	if err != nil {
		return
	}

	resp, err := c.cli.Do(ctx, req)
	if err != nil {
		return
	}
	if resp.StatusCode != http.StatusOK {
		return m, ErrInvalidFilterAPIResponse{http.StatusOK, resp.StatusCode, uri}
	}

	defer resp.Body.Close()
	b, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	if err = json.Unmarshal(b, &m); err != nil {
		return
	}

	return m, nil
}

// AddDimensionValue adds a particular value to a filter job for a given filterID
// and name
func (c *Client) AddDimensionValue(ctx context.Context, filterID, name, value string) error {
	uri := fmt.Sprintf("%s/filters/%s/dimensions/%s/options/%s", c.url, filterID, name, value)

	clientlog.Do(ctx, "adding dimension option to filter job", service, uri, log.Data{
		"method": "POST",
		"value":  value,
	})

	req, err := http.NewRequest("POST", uri, nil)
	if err != nil {
		return err
	}

	resp, err := c.cli.Do(ctx, req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusCreated {
		return &ErrInvalidFilterAPIResponse{http.StatusCreated, resp.StatusCode, uri}
	}
	return nil
}

// RemoveDimensionValue removes a particular value to a filter job for a given filterID
// and name
func (c *Client) RemoveDimensionValue(ctx context.Context, filterID, name, value string) error {
	uri := fmt.Sprintf("%s/filters/%s/dimensions/%s/options/%s", c.url, filterID, name, value)
	req, err := http.NewRequest("DELETE", uri, nil)
	if err != nil {
		return err
	}

	clientlog.Do(ctx, "removing dimension option from filter job", service, uri, log.Data{
		"method": "DELETE",
		"value":  value,
	})

	resp, err := c.cli.Do(ctx, req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusNoContent {
		return &ErrInvalidFilterAPIResponse{http.StatusNoContent, resp.StatusCode, uri}
	}
	return nil
}

// RemoveDimension removes a given dimension from a filter job
func (c *Client) RemoveDimension(ctx context.Context, filterID, name string) (err error) {
	uri := fmt.Sprintf("%s/filters/%s/dimensions/%s", c.url, filterID, name)

	clientlog.Do(ctx, "removing dimension from filter job", service, uri, log.Data{
		"method":    "DELETE",
		"dimension": "name",
	})

	req, err := http.NewRequest("DELETE", uri, nil)
	if err != nil {
		return
	}

	resp, err := c.cli.Do(ctx, req)
	if err != nil {
		return
	}

	if resp.StatusCode != http.StatusNoContent {
		err = &ErrInvalidFilterAPIResponse{http.StatusNoContent, resp.StatusCode, uri}
		return
	}

	return
}

// AddDimension adds a new dimension to a filter job
func (c *Client) AddDimension(ctx context.Context, id, name string) error {
	uri := fmt.Sprintf("%s/filters/%s/dimensions/%s", c.url, id, name)
	clientlog.Do(ctx, "adding dimension to filter job", service, uri, log.Data{
		"method":    "POST",
		"dimension": name,
	})

	req, err := http.NewRequest("POST", uri, bytes.NewBufferString(`{}`))
	if err != nil {
		return err
	}

	resp, err := c.cli.Do(ctx, req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusCreated {
		return errors.New("invalid status from filter api")
	}

	return nil
}

// GetJobState will return the current state of the filter job
func (c *Client) GetJobState(ctx context.Context, filterID string) (m Model, err error) {
	uri := fmt.Sprintf("%s/filters/%s", c.url, filterID)

	clientlog.Do(ctx, "retrieving filter job state", service, uri)

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return
	}

	resp, err := c.cli.Do(ctx, req)
	if err != nil {
		return
	}

	if resp.StatusCode != http.StatusOK {
		err = &ErrInvalidFilterAPIResponse{http.StatusOK, resp.StatusCode, uri}
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

// AddDimensionValues adds many options to a filter job dimension
func (c *Client) AddDimensionValues(ctx context.Context, filterID, name string, options []string) error {
	uri := fmt.Sprintf("%s/filters/%s/dimensions/%s", c.url, filterID, name)

	clientlog.Do(ctx, "adding multiple dimension values to filter job", service, uri, log.Data{
		"method":  "POST",
		"options": options,
	})

	body := struct {
		Options []string `json:"options"`
	}{
		Options: options,
	}

	b, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", uri, bytes.NewBuffer(b))
	if err != nil {
		return err
	}

	resp, err := c.cli.Do(ctx, req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusCreated {
		return &ErrInvalidFilterAPIResponse{http.StatusCreated, resp.StatusCode, uri}
	}

	return nil
}

// GetPreview attempts to retrieve a preview for a given filterOutputID
func (c *Client) GetPreview(ctx context.Context, filterOutputID string) (p Preview, err error) {
	uri := fmt.Sprintf("%s/filter-outputs/%s/preview", c.url, filterOutputID)

	clientlog.Do(ctx, "retrieving preview for filter output job", service, uri, log.Data{
		"method":   "GET",
		"filterID": filterOutputID,
	})

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return
	}

	resp, err := c.cli.Do(ctx, req)
	if err != nil {
		return
	}

	if resp.StatusCode != http.StatusOK {
		return p, &ErrInvalidFilterAPIResponse{http.StatusOK, resp.StatusCode, uri}
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	err = json.Unmarshal(b, &p)
	return
}
