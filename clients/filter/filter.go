package filter

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/ONSdigital/go-ns/clients/clientlog"
	"github.com/ONSdigital/go-ns/log"

	"github.com/ONSdigital/go-ns/rhttp"
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
	cli *rhttp.Client
	url string
}

func (c *Client) setInternalTokenHeader(req *http.Request, cfg ...Config) {
	if len(cfg) > 0 {
		req.Header.Set("Internal-token", cfg[0].InternalToken)
	}
}

// New creates a new instance of Client with a given filter api url
func New(filterAPIURL string) *Client {
	return &Client{
		cli: rhttp.DefaultClient,
		url: filterAPIURL,
	}
}

// Healthcheck calls the healthcheck endpoint on the api and alerts the caller of any errors
func (c *Client) Healthcheck() (string, error) {
	resp, err := c.cli.Get(c.url + "/healthcheck")
	if err != nil {
		return service, err
	}

	if resp.StatusCode != http.StatusOK {
		return service, &ErrInvalidFilterAPIResponse{http.StatusOK, resp.StatusCode, "/healthcheck"}
	}

	return service, nil
}

// GetOutput returns a filter output job for a given filter output id
func (c *Client) GetOutput(filterOutputID string, cfg ...Config) (m Model, err error) {
	uri := fmt.Sprintf("%s/filter-outputs/%s", c.url, filterOutputID)

	clientlog.Do("retrieving filter output", service, uri)

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return
	}
	c.setInternalTokenHeader(req, cfg...)

	resp, err := c.cli.Do(req)
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
func (c *Client) GetDimension(filterID, name string, cfg ...Config) (dim Dimension, err error) {
	uri := fmt.Sprintf("%s/filters/%s/dimensions/%s", c.url, filterID, name)

	clientlog.Do("retrieving dimension information", service, uri)

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return
	}
	c.setInternalTokenHeader(req, cfg...)

	resp, err := c.cli.Do(req)
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
func (c *Client) GetDimensions(filterID string, cfg ...Config) (dims []Dimension, err error) {
	uri := fmt.Sprintf("%s/filters/%s/dimensions", c.url, filterID)

	clientlog.Do("retrieving all dimensions for given filter job", service, uri)

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return
	}
	c.setInternalTokenHeader(req, cfg...)

	resp, err := c.cli.Do(req)
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
func (c *Client) GetDimensionOptions(filterID, name string, cfg ...Config) (opts []DimensionOption, err error) {
	uri := fmt.Sprintf("%s/filters/%s/dimensions/%s/options", c.url, filterID, name)

	clientlog.Do("retrieving selected dimension options for filter job", service, uri)

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return
	}
	c.setInternalTokenHeader(req, cfg...)

	resp, err := c.cli.Do(req)
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
func (c *Client) CreateBlueprint(instanceID string, names []string, cfg ...Config) (string, error) {
	fj := Model{InstanceID: instanceID}

	var dimensions []ModelDimension
	for _, name := range names {
		dimensions = append(dimensions, ModelDimension{Name: name})
	}

	fj.Dimensions = dimensions

	b, err := json.Marshal(fj)
	if err != nil {
		return "", err
	}

	uri := c.url + "/filters"
	clientlog.Do("attemping to create filter blueprint", service, uri, log.Data{
		"method":     "POST",
		"instanceID": instanceID,
	})

	req, err := http.NewRequest("POST", uri, bytes.NewBuffer(b))
	if err != nil {
		return "", err
	}
	c.setInternalTokenHeader(req, cfg...)

	resp, err := c.cli.Do(req)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusCreated {
		return "", errors.New("invalid status from filter api")
	}
	defer resp.Body.Close()

	b, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if err = json.Unmarshal(b, &fj); err != nil {
		return "", err
	}

	return fj.FilterID, nil
}

// UpdateBlueprint will update a blueprint with a given filter model
func (c *Client) UpdateBlueprint(m Model, doSubmit bool, cfg ...Config) (mdl Model, err error) {
	b, err := json.Marshal(m)
	if err != nil {
		return
	}

	uri := fmt.Sprintf("%s/filters/%s", c.url, m.FilterID)

	if doSubmit {
		uri = uri + "?submitted=true"
	}

	clientlog.Do("updating filter job", service, uri, log.Data{
		"method": "PUT",
		"body":   string(b),
	})

	req, err := http.NewRequest("PUT", uri, bytes.NewBuffer(b))
	if err != nil {
		return
	}
	c.setInternalTokenHeader(req, cfg...)

	resp, err := c.cli.Do(req)
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
func (c *Client) AddDimensionValue(filterID, name, value string, cfg ...Config) error {
	uri := fmt.Sprintf("%s/filters/%s/dimensions/%s/options/%s", c.url, filterID, name, value)

	clientlog.Do("adding dimension option to filter job", service, uri, log.Data{
		"method": "POST",
		"value":  value,
	})

	req, err := http.NewRequest("POST", uri, nil)
	if err != nil {
		return err
	}
	c.setInternalTokenHeader(req, cfg...)

	resp, err := c.cli.Do(req)
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
func (c *Client) RemoveDimensionValue(filterID, name, value string, cfg ...Config) error {
	uri := fmt.Sprintf("%s/filters/%s/dimensions/%s/options/%s", c.url, filterID, name, value)
	req, err := http.NewRequest("DELETE", uri, nil)
	if err != nil {
		return err
	}
	c.setInternalTokenHeader(req, cfg...)

	clientlog.Do("removing dimension option from filter job", service, uri, log.Data{
		"method": "DELETE",
		"value":  value,
	})

	resp, err := c.cli.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return &ErrInvalidFilterAPIResponse{http.StatusOK, resp.StatusCode, uri}
	}
	return nil
}

// RemoveDimension removes a given dimension from a filter job
func (c *Client) RemoveDimension(filterID, name string, cfg ...Config) (err error) {
	uri := fmt.Sprintf("%s/filters/%s/dimensions/%s", c.url, filterID, name)

	clientlog.Do("removing dimension from filter job", service, uri, log.Data{
		"method":    "DELETE",
		"dimension": "name",
	})

	req, err := http.NewRequest("DELETE", uri, nil)
	if err != nil {
		return
	}
	c.setInternalTokenHeader(req, cfg...)

	resp, err := c.cli.Do(req)
	if err != nil {
		return
	}

	if resp.StatusCode != http.StatusOK {
		err = &ErrInvalidFilterAPIResponse{http.StatusOK, resp.StatusCode, uri}
		return
	}

	return
}

// AddDimension adds a new dimension to a filter job
func (c *Client) AddDimension(id, name string, cfg ...Config) error {
	uri := fmt.Sprintf("%s/filters/%s/dimensions/%s", c.url, id, name)
	clientlog.Do("adding dimension to filter job", service, uri, log.Data{
		"method":    "POST",
		"dimension": name,
	})

	req, err := http.NewRequest("POST", uri, bytes.NewBufferString(`{}`))
	if err != nil {
		return err
	}
	c.setInternalTokenHeader(req, cfg...)

	resp, err := c.cli.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusCreated {
		return errors.New("invalid status from filter api")
	}

	return nil
}

// GetJobState will return the current state of the filter job
func (c *Client) GetJobState(filterID string, cfg ...Config) (m Model, err error) {
	uri := fmt.Sprintf("%s/filters/%s", c.url, filterID)

	clientlog.Do("retrieving filter job state", service, uri)

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return
	}
	c.setInternalTokenHeader(req, cfg...)

	resp, err := c.cli.Do(req)
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
func (c *Client) AddDimensionValues(filterID, name string, options []string, cfg ...Config) error {
	uri := fmt.Sprintf("%s/filters/%s/dimensions/%s", c.url, filterID, name)

	clientlog.Do("adding multiple dimension values to filter job", service, uri, log.Data{
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
	c.setInternalTokenHeader(req, cfg...)

	resp, err := c.cli.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusCreated {
		return &ErrInvalidFilterAPIResponse{http.StatusCreated, resp.StatusCode, uri}
	}

	return nil
}

// GetPreview attempts to retrieve a preview for a given filterOutputID
func (c *Client) GetPreview(filterOutputID string, cfg ...Config) (p Preview, err error) {
	uri := fmt.Sprintf("%s/filter-outputs/%s/preview", c.url, filterOutputID)

	clientlog.Do("retrieving preview for filter output job", service, uri, log.Data{
		"method":   "GET",
		"filterID": filterOutputID,
	})

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return
	}
	c.setInternalTokenHeader(req, cfg...)

	resp, err := c.cli.Do(req)
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
