package healthcheck

import (
	"github.com/ONSdigital/go-ns/rhttp"
	"net/http"
	"fmt"
)

// healthcheckClient is an implementation of Client that can be used to call the healthcheck endpoint of any service
type healthcheckClient struct {
	Client  HttpClient
	Url     string
	Service string
}

// HttpClient has a get method. Implemented by http.Client etc.
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

// NewClient creates a new Client for a service with the given name and healthcheck endpoint (url).
// it uses rhttp.Client to make requests to the given url, reporting on any response status code != 200.
func NewClient(service string, url string) *healthcheckClient {
	return &healthcheckClient{
		Client:  rhttp.DefaultClient,
		Url:     url,
		Service: service,
	}
}

// Healthcheck calls the endpoint url and alerts the caller of any errors
func (c *healthcheckClient) Healthcheck() (string, error) {
	resp, err := c.Client.Get(c.Url)
	if err != nil {
		return c.Service, err
	}

	if resp.StatusCode != http.StatusOK {
		return c.Service, &errorResponse{c.Service, http.StatusOK, resp.StatusCode, c.Url}
	}

	return c.Service, nil
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
