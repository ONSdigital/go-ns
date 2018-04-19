package elasticsearch

import (
	"encoding/json"
	"errors"
	"io/ioutil"

	"github.com/ONSdigital/go-ns/healthcheck"
	"github.com/ONSdigital/go-ns/rhttp"

	"net/http"
	"net/url"

	"github.com/ONSdigital/go-ns/log"
	"github.com/smartystreets/go-aws-auth"
)

// ensure the elasticsearchClient satisfies the Client interface.
var _ healthcheck.Client = (*HealthCheckClient)(nil)

// List of errors
var (
	ErrorUnexpectedStatusCode   = errors.New("unexpected status code from api")
	ErrorParsingBody            = errors.New("error parsing cluster health response body")
	ErrorUnhealthyClusterStatus = errors.New("error cluster health red")
)

const unhealthy = "red"

// HealthCheckClient provides a healthcheck.Client implementation for health checking elasticsearch.
type HealthCheckClient struct {
	cli          *rhttp.Client
	path         string
	serviceName  string
	signRequests bool
}

// ClusterHealth represents the response from the elasticsearch cluster health check
type ClusterHealth struct {
	Status string `json:"status"`
}

// NewHealthCheckClient returns a new elasticsearch health check client.
func NewHealthCheckClient(url string, signRequests bool) *HealthCheckClient {

	return &HealthCheckClient{
		cli:          rhttp.DefaultClient,
		path:         url + "/_cluster/health",
		serviceName:  "elasticsearch",
		signRequests: signRequests,
	}
}

// Healthcheck calls elasticsearch to check its health status.
func (elasticsearch *HealthCheckClient) Healthcheck() (string, error) {

	logData := log.Data{"url": elasticsearch.path}

	URL, err := url.Parse(elasticsearch.path)
	if err != nil {
		log.ErrorC("failed to create url for elasticsearch healthcheck", err, logData)
		return elasticsearch.serviceName, err
	}

	path := URL.String()
	logData["url"] = path

	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		log.ErrorC("failed to create request for healthcheck call to elasticsearch", err, logData)
		return elasticsearch.serviceName, err
	}

	if elasticsearch.signRequests {
		awsauth.Sign(req)
	}

	resp, err := elasticsearch.cli.Do(req)
	if err != nil {
		log.ErrorC("failed to call elasticsearch", err, logData)
		return elasticsearch.serviceName, err
	}
	defer resp.Body.Close()

	logData["http_code"] = resp.StatusCode
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= 300 {
		log.Error(ErrorUnexpectedStatusCode, logData)
		return elasticsearch.serviceName, ErrorUnexpectedStatusCode
	}

	jsonBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.ErrorC("failed to read response body from call to elastic", err, logData)
		return elasticsearch.serviceName, ErrorUnexpectedStatusCode
	}

	var clusterHealth ClusterHealth
	err = json.Unmarshal(jsonBody, &clusterHealth)
	if err != nil {
		log.Error(ErrorParsingBody, logData)
		return elasticsearch.serviceName, ErrorParsingBody
	}

	logData["cluster_health"] = clusterHealth.Status
	if clusterHealth.Status == unhealthy {
		log.Error(ErrorUnhealthyClusterStatus, logData)
		return elasticsearch.serviceName, ErrorUnhealthyClusterStatus
	}

	return elasticsearch.serviceName, nil
}
