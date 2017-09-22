package mongo

import (
	"context"
	"errors"
	"time"

	mgo "gopkg.in/mgo.v2"
)

// Mongo represents a simplistic MongoDB configuration.
type Mongo struct {
	Collection string
	Database   string
	URI        string
}

func (m *Mongo) Close(ctx context.Context, session *mgo.Session) error {
	closedChannel := make(chan bool)
	defer close(closedChannel)

	timeLeft := 1000 * time.Millisecond
	if deadline, ok := ctx.Deadline(); ok {
		// Add some time to timeLeft so case where ctx.Done in select
		// statement below gets called before time.After(timeLeft) gets called.
		// This is so the context error is returned over hardcoded error.
		timeLeft = deadline.Sub(time.Now()) + (10 * time.Millisecond)
	}

	go func() {
		// Uncomment the four line below for unit tests
		/*time.Sleep(1100 * time.Millisecond)
		if ctx.Value("return") == "true" || ctx.Err() != nil {
			return
		}*/

		session.Close()

		closedChannel <- true
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
