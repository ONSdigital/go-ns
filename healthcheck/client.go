package healthcheck

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ONSdigital/go-ns/common"
	"github.com/ONSdigital/go-ns/rchttp"
)

// healthcheckClient is an implementation of Client that can be used to call the healthcheck endpoint of any service
type healthcheckClient struct {
	client  common.RCHTTPClienter
	url     string
	service string
}

//go:generate moq -out mock_healthcheck/mock_httpclient.go -pkg mock_healthcheck . HttpClient

// httpClient has a get method. Implemented by http.Client etc.
type HttpClient interface {
	Get(url string) (*http.Response, error)
}

// errorResponse contains details of a failed healthcheck
type errorResponse struct {
	service      string
	expectedCode int
	actualCode   int
	uri          string
}

// NewClient creates a new Client for a service with the given name, healthcheck endpoint (url) and HttpClient,
// reporting on any response status code != 200.
func NewClient(service string, url string, client common.RCHTTPClienter) *healthcheckClient {
	return &healthcheckClient{
		client:  client,
		url:     url,
		service: service,
	}
}

// NewDefaultClient creates a new Client for a service with the given name and healthcheck endpoint (url).
// it uses rchttp.NewClient() to make requests to the given url, reporting on any response status code != 200.
func NewDefaultClient(service string, url string) *healthcheckClient {
	return NewClient(service, url, rchttp.NewClient())
}

// Healthcheck calls the endpoint url and alerts the caller of any errors
func (c *healthcheckClient) Healthcheck() (string, error) {
	resp, err := c.client.Get(context.Background(), c.url)
	if err != nil {
		return c.service, err
	}

	if resp.StatusCode != http.StatusOK {
		return c.service, &errorResponse{c.service, http.StatusOK, resp.StatusCode, c.url}
	}

	return c.service, nil
}

// Error should be called by the user to print out the stringified version of the error
func (e errorResponse) Error() string {
	return fmt.Sprintf("invalid response from %s - should be: %d, got: %d, path: %s",
		e.service,
		e.expectedCode,
		e.actualCode,
		e.uri,
	)
}
