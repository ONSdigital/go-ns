package mongo

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/ONSdigital/go-ns/log"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	hasSessionSleep bool
	session         *mgo.Session

	Collection = "test"
	Database   = "test"
	URI        = "localhost:27017"
)

// Mongo represents a simplistic MongoDB configuration.
type Mongo struct {
	Collection string
	Database   string
	URI        string
}

type ungraceful struct{}

func (t ungraceful) shutdown(ctx context.Context, session *mgo.Session, closedChannel chan bool) {
	time.Sleep(timeLeft + (100 * time.Millisecond))
	if ctx.Value("return") == "true" || ctx.Err() != nil {
		return
	}

	session.Close()

	closedChannel <- true
	return
}

func TestSuccessfulCloseMongoSession(t *testing.T) {
	_, err := setupSession()
	if err != nil {
		log.Info("mongo instance not available, skip tests", log.Data{"error": err})
		os.Exit(0)
	}

	if err = cleanupTestData(session.Copy()); err != nil {
		log.ErrorC("Failed to delete test data", err, nil)
	}

	Convey("Safely close mongo session", t, func() {
		if !hasSessionSleep {
			Convey("with no context deadline", func() {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				err := Close(ctx, session.Copy())

				So(err, ShouldBeNil)
			})
		}

		Convey("within context timeout (deadline)", func() {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			err := Close(ctx, session.Copy())

			So(err, ShouldBeNil)
		})

		Convey("within context deadline", func() {
			time := time.Now().Local().Add(time.Second * time.Duration(2))
			ctx, cancel := context.WithDeadline(context.Background(), time)
			defer cancel()
			err := Close(ctx, session.Copy())

			So(err, ShouldBeNil)
		})
	})

	if err = setUpTestData(session.Copy()); err != nil {
		log.ErrorC("Failed to insert test data, skipping tests", err, nil)
		os.Exit(1)
	}

	Convey("Timed out from safely closing mongo session", t, func() {
		Convey("with no context deadline", func() {
			start = ungraceful{}
			copiedSession := session.Copy()
			go func() {
				_ = queryMongo(copiedSession)
			}()
			// Sleep for half a second for mongo query to begin
			time.Sleep(500 * time.Millisecond)

			contextKey := "return"
			ctx := context.WithValue(context.Background(), contextKey, "true")
			err := Close(ctx, copiedSession)

			So(err, ShouldNotBeNil)
			So(err, ShouldResemble, errors.New("closing mongo timed out"))
			time.Sleep(500 * time.Millisecond)
		})

		Convey("with context deadline", func() {
			copiedSession := session.Copy()
			go func() {
				_ = queryMongo(copiedSession)
			}()
			// Sleep for half a second for mongo query to begin
			time.Sleep(500 * time.Millisecond)

			ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
			defer cancel()
			err := Close(ctx, copiedSession)

			So(err, ShouldNotBeNil)
			So(err, ShouldResemble, context.DeadlineExceeded)
		})
	})

	if err = cleanupTestData(session.Copy()); err != nil {
		log.ErrorC("Failed to delete test data", err, nil)
	}
}

func cleanupTestData(session *mgo.Session) error {
	defer session.Close()

	err := session.DB(Database).DropDatabase()
	if err != nil {
		return err
	}

	return nil
}

func queryMongo(session *mgo.Session) error {
	defer session.Close()

	_, err := session.DB(Database).C(Collection).Find(bson.M{"$where": "sleep(2000) || true"}).Count()
	if err != nil {
		return err
	}

	return nil
}

func setUpTestData(session *mgo.Session) error {
	defer session.Close()

	firstDoc := bson.M{
		"_id":   "1",
		"state": "first",
	}

	secondDoc := bson.M{
		"_id":   "2",
		"state": "second",
	}

	_, err := session.DB(Database).C(Collection).Upsert(bson.M{"_id": "1"}, firstDoc)
	if err != nil {
		return err
	}

	_, err = session.DB(Database).C(Collection).Upsert(bson.M{"_id": "2"}, secondDoc)
	if err != nil {
		return err
	}

	return nil
}

func setupSession() (*Mongo, error) {
	mongo := &Mongo{
		Collection: Collection,
		Database:   Database,
		URI:        URI,
	}

	if session != nil {
		return nil, errors.New("Failed to initialise mongo")
	}

	var err error

	if session, err = mgo.Dial(URI); err != nil {
		return nil, err
	}

	session.EnsureSafe(&mgo.Safe{WMode: "majority"})
	session.SetMode(mgo.Strong, true)
	return mongo, nil
}
