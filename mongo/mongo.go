package mongo

import (
	"context"
	"errors"
	"time"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// keep these in sync with Timestamps tags below
const (
	lastUpdatedKey     = "last_updated"
	uniqueTimestampKey = "unique_timestamp"
)

// keep tags in sync with above const
type Timestamps struct {
	LastUpdated     time.Time           `bson:"last_updated,omitempty"     json:"last_updated,omitempty"`
	UniqueTimestamp bson.MongoTimestamp `bson:"unique_timestamp,omitempty" json:"-"`
}

// Shutdown represents an interface to the shutdown method
type Shutdown interface {
	shutdown(ctx context.Context, session *mgo.Session, closedChannel chan bool)
}

type graceful struct{}

func (t graceful) shutdown(ctx context.Context, session *mgo.Session, closedChannel chan bool) {
	session.Close()

	closedChannel <- true
	return
}

var (
	start    Shutdown = graceful{}
	timeLeft          = 1000 * time.Millisecond
)

// Close represents mongo session closing within the context deadline
func Close(ctx context.Context, session *mgo.Session) error {
	closedChannel := make(chan bool)
	defer close(closedChannel)

	if deadline, ok := ctx.Deadline(); ok {
		// Add some time to timeLeft so case where ctx.Done in select
		// statement below gets called before time.After(timeLeft) gets called.
		// This is so the context error is returned over hardcoded error.
		timeLeft = deadline.Sub(time.Now()) + (10 * time.Millisecond)
	}

	go func() {
		start.shutdown(ctx, session, closedChannel)
		return
	}()

	select {
	case <-time.After(timeLeft):
		return errors.New("closing mongo timed out")
	case <-closedChannel:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// withCurrentDate creates or adds $currentDate to updateDoc - populates that with key:val
func withCurrentDate(updateDoc bson.M, key string, val interface{}) bson.M {
	var currentDate bson.M
	var ok bool
	if currentDate, ok = updateDoc["$currentDate"].(bson.M); !ok {
		currentDate = bson.M{}
	}
	switch v := val.(type) {
	case bool, bson.M:
		currentDate[key] = v
	default:
		panic("Cannot handle that type")
	}
	updateDoc["$currentDate"] = currentDate
	return updateDoc
}

// WithUpdates adds all timestamps to updateDoc
func WithUpdates(updateDoc bson.M) bson.M {
	return WithUniqueTimestampUpdate(WithLastUpdatedUpdate(updateDoc))
}

// WithNamespacedUpdates adds all timestamps to updateDoc
func WithNamespacedUpdates(updateDoc bson.M, prefixes []string) bson.M {
	return WithNamespacedUniqueTimestampUpdate(WithNamespacedLastUpdatedUpdate(updateDoc, prefixes), prefixes)
}

// WithLastUpdatedUpdate adds last_updated to updateDoc
func WithLastUpdatedUpdate(updateDoc bson.M) bson.M {
	return withCurrentDate(updateDoc, lastUpdatedKey, true)
}

// WithNamespacedLastUpdatedUpdate adds unique timestamp to updateDoc
func WithNamespacedLastUpdatedUpdate(updateDoc bson.M, prefixes []string) bson.M {
	newUpdateDoc := updateDoc
	for _, prefix := range prefixes {
		newUpdateDoc = withCurrentDate(newUpdateDoc, prefix+lastUpdatedKey, true)
	}
	return newUpdateDoc
}

// WithUniqueTimestampUpdate adds unique timestamp to updateDoc
func WithUniqueTimestampUpdate(updateDoc bson.M) bson.M {
	return withCurrentDate(updateDoc, uniqueTimestampKey, bson.M{"$type": "timestamp"})
}

// WithNamespacedUniqueTimestampUpdate adds unique timestamp to updateDoc
func WithNamespacedUniqueTimestampUpdate(updateDoc bson.M, prefixes []string) bson.M {
	newUpdateDoc := updateDoc
	for _, prefix := range prefixes {
		newUpdateDoc = withCurrentDate(newUpdateDoc, prefix+uniqueTimestampKey, bson.M{"$type": "timestamp"})
	}
	return newUpdateDoc
}

// WithUniqueTimestampQuery adds unique timestamp to queryDoc
func WithUniqueTimestampQuery(queryDoc bson.M, timestamp bson.MongoTimestamp) bson.M {
	queryDoc[uniqueTimestampKey] = timestamp
	return queryDoc
}

// WithNamespacedUniqueTimestampQuery adds unique timestamps to queryDoc sub-docs
func WithNamespacedUniqueTimestampQuery(queryDoc bson.M, timestamps []bson.MongoTimestamp, prefixes []string) bson.M {
	newQueryDoc := queryDoc
	for idx, prefix := range prefixes {
		newQueryDoc[prefix+uniqueTimestampKey] = timestamps[idx]
	}
	return newQueryDoc
}
