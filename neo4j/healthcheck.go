package neo4j

import (
	"github.com/ONSdigital/go-ns/healthcheck"
	"github.com/ONSdigital/go-ns/log"
	bolt "github.com/ONSdigital/golang-neo4j-bolt-driver"
)

// ensure the Neo4jClient satisfies the Client interface.
var _ healthcheck.Client = (*HealthCheckClient)(nil)

const pingStmt = "MATCH (i) RETURN i LIMIT 1"

// HealthCheckClient provides a healthcheck.Client implementation for health checking neo4j.
type HealthCheckClient struct {
	dbPool      bolt.ClosableDriverPool
	serviceName string
}

// NewHealthCheckClient returns a new neo4j health check client using the given connection pool.
func NewHealthCheckClient(dbPool bolt.ClosableDriverPool) *HealthCheckClient {

	return &HealthCheckClient{
		dbPool:      dbPool,
		serviceName: "neo4j",
	}
}

// Healthcheck calls neo4j to check its health status.
func (neo4j *HealthCheckClient) Healthcheck() (string, error) {

	logData := log.Data{"statement": pingStmt}
	conn, err := neo4j.dbPool.OpenPool()
	if err != nil {
		log.ErrorC("neo4j healthcheck open pool", err, logData)
		return neo4j.serviceName, err
	}
	defer conn.Close()

	rows, err := conn.QueryNeo(pingStmt, nil)
	if err != nil {
		log.ErrorC("neo4j healthcheck query", err, logData)
		return neo4j.serviceName, err
	}
	defer rows.Close()

	return neo4j.serviceName, nil
}
