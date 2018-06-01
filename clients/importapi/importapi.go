package importapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/ONSdigital/go-ns/common"
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/go-ns/rchttp"
)

const service = "import-api"

// Clienter defines that interface for a client of the ImportAPI
type Clienter interface {
	GetImportJob(ctx context.Context, importJobID string) (ImportJob, bool, error)
	UpdateImportJobState(ctx context.Context, jobID string, newState string) error
}

// Client is an import api client which can be used to make requests to the API
type Client struct {
	client common.RCHTTPClienter
	url    string
}

// NewAPIClient creates a new API Client with initial rchttp client (optional), given url, auth token
func NewAPIClient(client common.RCHTTPClienter, apiURL, serviceToken string) *Client {
	return &Client{
		client: rchttp.ClientWithServiceToken(client, serviceToken),
		url:    apiURL,
	}
}

// ErrInvalidAPIResponse is returned when the api does not respond with a valid status
type ErrInvalidAPIResponse struct {
	actualCode int
	uri        string
	body       string
}

// Error should be called by the user to print out the stringified version of the error
func (e ErrInvalidAPIResponse) Error() string {
	return fmt.Sprintf(
		"invalid response: %d from %s: %s, body: %s",
		e.actualCode,
		service,
		e.uri,
		e.body,
	)
}

// Code returns the status code received from the api if an error is returned
func (e ErrInvalidAPIResponse) Code() int {
	return e.actualCode
}

var _ error = ErrInvalidAPIResponse{}

// Healthcheck calls the healthcheck endpoint on the api and alerts the caller of any errors
func (c *Client) Healthcheck() (string, error) {
	ctx := context.Background()

	resp, err := c.client.Get(ctx, c.url+"/healthcheck")
	if err != nil {
		return service, err
	}

	if resp.StatusCode != http.StatusOK {
		return service, NewAPIResponse(resp, "/healthcheck")
	}

	return service, nil
}

// ImportJob comes from the Import API and links an import job to its (other) instances
type ImportJob struct {
	JobID string  `json:"id"`
	Links LinkMap `json:"links,omitempty"`
}

// LinkMap is an array of instance links associated with am import job
type LinkMap struct {
	Instances []InstanceLink `json:"instances"`
}

// InstanceLink identifies an (instance or import-job) by id and url (from Import API)
type InstanceLink struct {
	ID   string `json:"id"`
	Link string `json:"href"`
}

// GetImportJob asks the Import API for the details for an Import job
func (api *Client) GetImportJob(ctx context.Context, importJobID string) (ImportJob, bool, error) {
	var importJob ImportJob
	path := api.url + "/jobs/" + importJobID

	jsonBody, httpCode, err := api.getJSON(ctx, path, 0, nil)
	if httpCode == http.StatusNotFound {
		return importJob, false, nil
	}
	logData := log.Data{
		"path":        path,
		"importJobID": importJobID,
		"httpCode":    httpCode,
		"jsonBody":    string(jsonBody),
	}
	var isFatal bool
	if err == nil && httpCode != http.StatusOK {
		if httpCode < http.StatusInternalServerError {
			isFatal = true
		}
		err = errors.New("Bad response while getting import job")
	} else {
		isFatal = true
	}
	if err != nil {
		log.ErrorC("GetImportJob", err, logData)
		return importJob, isFatal, err
	}

	if err := json.Unmarshal(jsonBody, &importJob); err != nil {
		log.ErrorC("GetImportJob unmarshal", err, logData)
		return ImportJob{}, true, err
	}

	return importJob, false, nil
}

// UpdateImportJobState tells the Import API that the state has changed of an Import job
func (api *Client) UpdateImportJobState(ctx context.Context, jobID string, newState string) error {
	path := api.url + "/jobs/" + jobID
	jsonUpload := []byte(`{"state":"` + newState + `"}`)

	jsonResult, httpCode, err := api.putJSON(ctx, path, 0, jsonUpload)
	logData := log.Data{
		"path":        path,
		"importJobID": jobID,
		"jsonUpload":  jsonUpload,
		"httpCode":    httpCode,
		"jsonResult":  jsonResult,
	}
	if err == nil && httpCode != http.StatusOK {
		err = errors.New("Bad HTTP response")
	}
	if err != nil {
		log.ErrorC("UpdateImportJobState", err, logData)
		return err
	}
	return nil
}

func (api *Client) getJSON(ctx context.Context, path string, attempts int, vars url.Values) ([]byte, int, error) {
	return callJSONAPI(ctx, api.client, "GET", path, vars)
}

func (api *Client) putJSON(ctx context.Context, path string, attempts int, payload []byte) ([]byte, int, error) {
	return callJSONAPI(ctx, api.client, "PUT", path, payload)
}

func callJSONAPI(ctx context.Context, client common.RCHTTPClienter, method, path string, payload interface{}) ([]byte, int, error) {

	logData := log.Data{"url": path, "method": method}

	URL, err := url.Parse(path)
	if err != nil {
		log.ErrorC("Failed to create url for API call", err, logData)
		return nil, 0, err
	}
	path = URL.String()
	logData["url"] = path

	var req *http.Request

	if payload != nil && method != "GET" {
		req, err = http.NewRequest(method, path, bytes.NewReader(payload.([]byte)))
		req.Header.Add("Content-type", "application/json")
		logData["payload"] = string(payload.([]byte))
	} else {
		req, err = http.NewRequest(method, path, nil)

		if payload != nil && method == "GET" {
			req.URL.RawQuery = payload.(url.Values).Encode()
			logData["payload"] = payload.(url.Values)
		}
	}
	// check above req had no errors
	if err != nil {
		log.ErrorC("Failed to create request for API", err, logData)
		return nil, 0, err
	}

	resp, err := client.Do(ctx, req)
	if err != nil {
		log.ErrorC("Failed to action API", err, logData)
		return nil, 0, err
	}

	logData["httpCode"] = resp.StatusCode
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= 300 {
		log.Debug("unexpected status code from API", logData)
	}

	jsonBody, err := getBody(resp)
	if err != nil {
		log.ErrorC("Failed to read body from API", err, logData)
		return nil, resp.StatusCode, err
	}
	return jsonBody, resp.StatusCode, nil
}

// NewAPIResponse creates an error response, optionally adding body to e when status is 404
func NewAPIResponse(resp *http.Response, uri string) (e *ErrInvalidAPIResponse) {
	e = &ErrInvalidAPIResponse{
		actualCode: resp.StatusCode,
		uri:        uri,
	}
	if resp.StatusCode == http.StatusNotFound {
		body, err := getBody(resp)
		if err != nil {
			e.body = "Client failed to read response body"
			return
		}
		e.body = string(body)
	}
	return
}

func getBody(resp *http.Response) ([]byte, error) {
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err = resp.Body.Close(); err != nil {
		log.ErrorC("closing body", err, nil)
		return nil, err
	}
	return b, nil
}
