package mongo

import (
	"context"
	"errors"
	"flag"
	"os"
	"testing"
	"time"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/ONSdigital/go-ns/log"
	. "github.com/smartystreets/goconvey/convey"
)

var session *mgo.Session
var (
	Collection = "test"
	Database   = "test"
	URI        = "localhost:27017"
)
var hasSessionSleep bool

func init() {
	// To run all tests uncomment out lines 32 to 35 in mongo.go
	// and run go test -has.session.sleep=true ./... under the mongo package
	flag.BoolVar(&hasSessionSleep, "has.session.sleep", false, "enables tests that require close channel function to sleep before closing down mongo session")
	flag.Parse()
}

func TestSuccessfulCloseMongoSession(t *testing.T) {
	mongo, err := setupSession()
	if err != nil {
		log.Info("mongo instance not available, skip tests", log.Data{"error": err})
		os.Exit(0)
	}

	Convey("Safely close mongo session", t, func() {
		if !hasSessionSleep {
			Convey("with no context deadline", func() {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				err := mongo.Close(ctx, session.Copy())

				So(err, ShouldBeNil)
			})
		}

		Convey("within context timeout (deadline)", func() {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			err := mongo.Close(ctx, session.Copy())

			So(err, ShouldBeNil)
		})

		Convey("within context deadline", func() {
			time := time.Now().Local().Add(time.Second * time.Duration(2))
			ctx, cancel := context.WithDeadline(context.Background(), time)
			defer cancel()
			err := mongo.Close(ctx, session.Copy())

			So(err, ShouldBeNil)
		})
	})

	if !hasSessionSleep {
		log.Info("Skipping tests", nil)
	} else {
		if err = setUpTestData(session.Copy()); err != nil {
			log.ErrorC("Failed to insert test data, skipping tests", err, nil)
			os.Exit(1)
		}

		Convey("Timed out from safely closing mongo session", t, func() {
			Convey("with no context deadline", func() {
				copiedSession := session.Copy()
				go func() {
					_ = queryMongo(copiedSession)
				}()
				// Sleep for half a second for mongo query to begin
				time.Sleep(500 * time.Millisecond)

				contextKey := "return"
				ctx := context.WithValue(context.Background(), contextKey, "true")
				err := mongo.Close(ctx, copiedSession)

				So(err, ShouldNotBeNil)
				So(err, ShouldResemble, errors.New("closing mongo timed out"))
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
				err := mongo.Close(ctx, copiedSession)

				So(err, ShouldNotBeNil)
				So(err, ShouldResemble, context.DeadlineExceeded)
			})
		})

		if err = cleanupTestData(session.Copy()); err != nil {
			log.ErrorC("Failed to delete test data", err, nil)
		}
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
