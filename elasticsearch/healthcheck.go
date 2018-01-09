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
	cli         *rhttp.Client
	path        string
	serviceName string
}

// ClusterHealth represents the response from the elasticsearch cluster health check
type ClusterHealth struct {
	CLusterName                 string `json:"cluster_name"`
	Status                      string `json:"status"`
	TimedOut                    bool   `json:"timed_out"`
	NumberOfNodes               int64  `json:"number_of_nodes"`
	NumberOfDataNodes           int64  `json:"number_of_data_nodes"`
	ActivePrimaryShards         int64  `json:"active_primary_shards"`
	ActiveShards                int64  `json:"active_shards"`
	RelocatingShards            int64  `json:"relocating_shards"`
	InitialisingShards          int64  `json:"initializing_shards"`
	UnassignedShards            int64  `json:"unassigned_shards"`
	DelayedUnassignedShards     int64  `json:"delayed_unassigned_shards"`
	NumberOfPendingTasks        int64  `json:"number_of_pending_tasks"`
	NumberOfInFlightFetch       int64  `json:"number_of_in_flight_fetch"`
	TaskMaxWaitingInQueueMillis int64  `json:"task_max_waiting_in_queue_millis"`
}

// NewHealthCheckClient returns a new elasticsearch health check client.
func NewHealthCheckClient(url string) *HealthCheckClient {

	return &HealthCheckClient{
		cli:         rhttp.DefaultClient,
		path:        url + "/_cluster/health",
		serviceName: "elasticsearch",
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
	logData["URL"] = path

	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		log.ErrorC("failed to create request for healthcheck call to elasticsearch", err, logData)
		return elasticsearch.serviceName, err
	}

	resp, err := elasticsearch.cli.Do(req)
	if err != nil {
		log.ErrorC("Failed to call elasticsearch", err, logData)
		return elasticsearch.serviceName, err
	}
	defer resp.Body.Close()

	logData["httpCode"] = resp.StatusCode
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= 300 {
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
		return elasticsearch.serviceName, ErrorParsingBody
	}

	if clusterHealth.Status == unhealthy {
		return elasticsearch.serviceName, ErrorUnhealthyClusterStatus
	}

	return elasticsearch.serviceName, nil
}
