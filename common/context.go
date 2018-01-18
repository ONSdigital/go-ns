package common

import (
	"context"
	"net/http"

	uuid "github.com/satori/go.uuid"
)

const (
	cIdKey    = "correlation_id"
	cIdHeader = "x-request-id"
)

// ONSContext wraps a context which must contain the key: cIdKey
type ONSContext context.Context

// NewONSContext returns a context with a correlation ID - it may be created from
// a context, an http request or a string ("" to create a new string).
// Any of these arguments may already be populated with a correlation ID, in which case,
// that is re-used.
func NewONSContext(c interface{}) ONSContext {
	switch ctr := c.(type) {
	case context.Context:
		if ctr == nil || ctr.Value(cIdKey) == nil || ctr.Value(cIdKey).(string) == "" {
			onsCtx := context.WithValue(ctr, cIdKey, newID()).(ONSContext)
			return onsCtx
		}
		return ctr.(ONSContext)
	case http.Request:
		correlationIds := ctr.Header.Get(cIdHeader)
		if correlationIds == "" {
			correlationIds = newID()
		}
		return context.WithValue(ctr.Context(), cIdKey, correlationIds).(ONSContext)
	case string:
		if ctr == "" {
			ctr = newID()
		}
		return context.WithValue(context.TODO(), cIdKey, ctr).(ONSContext)
	default:
		panic("bad type for NewONSContext")
	}
	return nil
}

// GetContextID returns the argument's correlation ID
func GetContextID(c ONSContext) string {
	return c.Value(cIdKey).(string)
}

func newID() string {
	return uuid.NewV4().String()
}
