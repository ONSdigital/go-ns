package elasticsearch

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"time"

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
	ErrorTimedOut               = errors.New("timeout waiting for response")
)

const unhealthy = "red"

// HealthCheckClient provides a healthcheck.Client implementation for health checking elasticsearch.
type HealthCheckClient struct {
	cli          *rhttp.Client
	path         string
	serviceName  string
	signRequests bool
	timeout      time.Duration
}

// ClusterHealth represents the response from the elasticsearch cluster health check
type ClusterHealth struct {
	Status string `json:"status"`
}

// NewHealthCheckClient returns a new elasticsearch health check client.
func NewHealthCheckClient(url string, signRequests bool, timeout time.Duration) *HealthCheckClient {

	return &HealthCheckClient{
		cli:          rhttp.DefaultClient,
		path:         url + "/_cluster/health",
		serviceName:  "elasticsearch",
		signRequests: signRequests,
		timeout:      timeout,
	}
}

type healthResult struct {
	Error error
}

// Healthcheck calls elasticsearch to check its health status.
func (elasticsearch *HealthCheckClient) Healthcheck() (string, error) {

	logData := log.Data{"url": elasticsearch.path}

	healthChan := make(chan healthResult)
	defer close(healthChan)

	ctx, cancel := context.WithTimeout(context.Background(), elasticsearch.timeout)

	go func(ctx context.Context) {
		URL, err := url.Parse(elasticsearch.path)
		if err != nil {
			log.ErrorC("failed to create url for elasticsearch healthcheck", err, logData)

			if ctx.Err() == nil {
				healthChan <- healthResult{Error: err}
			}

			return
		}

		path := URL.String()
		logData["url"] = path

		req, err := http.NewRequest("GET", path, nil)
		if err != nil {
			log.ErrorC("failed to create request for healthcheck call to elasticsearch", err, logData)

			if ctx.Err() == nil {
				healthChan <- healthResult{Error: err}
			}

			return
		}

		if elasticsearch.signRequests {
			awsauth.Sign(req)
		}

		resp, err := elasticsearch.cli.Do(req)
		if err != nil {
			log.ErrorC("failed to call elasticsearch", err, logData)

			if ctx.Err() == nil {
				healthChan <- healthResult{Error: err}
			}

			return
		}
		defer resp.Body.Close()

		logData["http_code"] = resp.StatusCode
		if resp.StatusCode < http.StatusOK || resp.StatusCode >= 300 {
			log.Error(ErrorUnexpectedStatusCode, logData)

			if ctx.Err() == nil {
				healthChan <- healthResult{Error: ErrorUnexpectedStatusCode}
			}

			return
		}

		jsonBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {

			log.ErrorC("failed to read response body from call to elastic", err, logData)
			if ctx.Err() == nil {
				healthChan <- healthResult{Error: ErrorUnexpectedStatusCode}
			}

			return
		}

		var clusterHealth ClusterHealth
		err = json.Unmarshal(jsonBody, &clusterHealth)
		if err != nil {
			log.Error(ErrorParsingBody, logData)
			if ctx.Err() == nil {
				healthChan <- healthResult{Error: ErrorParsingBody}
			}

			return
		}

		logData["cluster_health"] = clusterHealth.Status
		if clusterHealth.Status == unhealthy {
			log.Error(ErrorUnhealthyClusterStatus, logData)
			if ctx.Err() == nil {
				healthChan <- healthResult{Error: ErrorUnhealthyClusterStatus}
			}

			return
		}

		if ctx.Err() == nil {
			healthChan <- healthResult{Error: nil}
		}

		return
	}(ctx)

	var myError error
	select {
	case res := <-healthChan:
		myError = res.Error
	case <-ctx.Done():
		log.Error(ErrorTimedOut, logData)
		myError = ErrorTimedOut
	}

	cancel()
	return elasticsearch.serviceName, myError
}
