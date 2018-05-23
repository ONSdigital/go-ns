package mongo

import (
	"context"
	"errors"
	"time"

	"github.com/globalsign/mgo"
)

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
