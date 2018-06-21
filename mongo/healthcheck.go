package mongo

import (
	"github.com/ONSdigital/go-ns/log"
	mgo "github.com/globalsign/mgo"
)

// HealthCheckClient provides a healthcheck.Client implementation for health checking the service
type HealthCheckClient struct {
	mongo       *mgo.Session
	serviceName string
}

// NewHealthCheckClient returns a new health check client using the given service
func NewHealthCheckClient(db *mgo.Session) *HealthCheckClient {
	return &HealthCheckClient{
		mongo:       db,
		serviceName: "mongodb",
	}
}

// Healthcheck calls service to check its health status
func (m *HealthCheckClient) Healthcheck() (res string, err error) {
	s := m.mongo.Copy()
	defer s.Close()
	res = m.serviceName
	err = s.Ping()
	if err != nil {
		log.ErrorC("Ping mongo", err, nil)
	}

	return
}
