package filter

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// ErrInvalidFilterAPIResponse is returned when the filter api does not respond
// with a valid status
type ErrInvalidFilterAPIResponse struct {
	expectedCode int
	actualCode   int
	uri          string
}

func (e ErrInvalidFilterAPIResponse) Error() string {
	return fmt.Sprintf("invalid response from filter api - should be: %d, got: %d, path: %s",
		e.expectedCode,
		e.actualCode,
		e.uri,
	)
}

var _ error = ErrInvalidFilterAPIResponse{}

// Client is a filter api client which can be used to make requests to the server
type Client struct {
	cli *http.Client
	url string
}

// New creates a new instance of Client with a given filter api url
func New(filterAPIURL string) *Client {
	return &Client{
		cli: &http.Client{Timeout: 5 * time.Second},
		url: filterAPIURL,
	}
}

// Healthcheck calls the healthcheck endpoint on the api and alerts the caller of any errors
func (c *Client) Healthcheck() (string, error) {
	resp, err := c.cli.Get(c.url + "/healthcheck")
	if err != nil {
		return "filter-api", err
	}

	if resp.StatusCode != http.StatusOK {
		return "filter-api", &ErrInvalidFilterAPIResponse{http.StatusOK, resp.StatusCode, "/healthcheck"}
	}

	return "", nil
}

// GetDimension returns information on a requested dimension name for a given filterID
func (c *Client) GetDimension(filterID, name string) (dim Dimension, err error) {
	uri := fmt.Sprintf("%s/filters/%s/dimensions/%s", c.url, filterID, name)
	resp, err := c.cli.Get(uri)
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
func (c *Client) GetDimensions(filterID string) (dims []Dimension, err error) {
	uri := fmt.Sprintf("%s/filters/%s/dimensions", c.url, filterID)
	resp, err := c.cli.Get(uri)
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
func (c *Client) GetDimensionOptions(filterID, name string) (opts []DimensionOption, err error) {
	uri := fmt.Sprintf("%s/filters/%s/dimensions/%s/options", c.url, filterID, name)
	resp, err := c.cli.Get(uri)
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

// AddDimensionValue adds a particular value to a filter job for a given filterID
// and name
func (c *Client) AddDimensionValue(filterID, name, value string) error {
	uri := fmt.Sprintf("%s/filters/%s/dimensions/%s/options/%s", c.url, filterID, name, value)
	req, err := http.NewRequest("POST", uri, nil)
	if err != nil {
		return err
	}

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
func (c *Client) RemoveDimensionValue(filterID, name, value string) error {
	uri := fmt.Sprintf("%s/filters/%s/dimensions/%s/options/%s", c.url, filterID, name, value)
	req, err := http.NewRequest("DELETE", uri, nil)
	if err != nil {
		return err
	}

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
func (c *Client) RemoveDimension(filterID, name string) (err error) {
	uri := fmt.Sprintf("%s/filters/%s/dimensions/%s", c.url, filterID, name)
	req, err := http.NewRequest("DELETE", uri, nil)
	if err != nil {
		return
	}

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
func (c *Client) AddDimension(id, name string) error {
	resp, err := http.Post(fmt.Sprintf("%s/filters/%s/dimensions/%s", c.url, id, name), "application/json", bytes.NewBufferString(`{}`))
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusCreated {
		return errors.New("invalid status from filter api")
	}

	return nil
}

// GetJobState will return the current state of the filter job
func (c *Client) GetJobState(filterID string) (m Model, err error) {
	uri := fmt.Sprintf("%s/filters/%s", c.url, filterID)
	resp, err := c.cli.Get(uri)
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
func (c *Client) AddDimensionValues(filterID, name string, options []string) error {
	uri := fmt.Sprintf("%s/filters/%s/dimensions/%s", c.url, filterID, name)

	body := struct {
		Options []string `json:"options"`
	}{
		Options: options,
	}

	b, err := json.Marshal(body)
	if err != nil {
		return err
	}

	resp, err := c.cli.Post(uri, "application/json", bytes.NewBuffer(b))
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusCreated {
		return &ErrInvalidFilterAPIResponse{http.StatusCreated, resp.StatusCode, uri}
	}

	return nil
}
