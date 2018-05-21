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

const (
	returnContextKey = "want_return"
	early            = "early"
)

// Mongo represents a simplistic MongoDB configuration.
type Mongo struct {
	Collection string
	Database   string
	URI        string
}

type TestModel struct {
	State           string              `bson:"state"`
	NewKey          int                 `bson:"new_key,omitempty"`
	LastUpdated     time.Time           `bson:"last_updated"`
	UniqueTimestamp bson.MongoTimestamp `bson:"unique_timestamp,omitempty"`
}

type Times struct {
	LastUpdated     time.Time           `bson:"last_updated"`
	UniqueTimestamp bson.MongoTimestamp `bson:"unique_timestamp,omitempty"`
}

type testNamespacedModel struct {
	State   string `bson:"state"`
	NewKey  int    `bson:"new_key,omitempty"`
	Currant Times  `bson:"currant",omitempty"`
	Nixed   Times  `bson:"nixed",omitempty"`
}

type ungraceful struct{}

func (t ungraceful) shutdown(ctx context.Context, session *mgo.Session, closedChannel chan bool) {
	time.Sleep(timeLeft + (100 * time.Millisecond))
	if ctx.Value(returnContextKey) == early || ctx.Err() != nil {
		return
	}

	session.Close()

	closedChannel <- true
	return
}

func TestSuccessfulCloseMongoSession(t *testing.T) {
	_, err := setupSession()
	if err != nil {
		log.Info("mongo instance not available, skip close tests", log.Data{"error": err})
		return
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
				_ = slowQueryMongo(copiedSession)
			}()
			// Sleep for half a second for mongo query to begin
			time.Sleep(500 * time.Millisecond)

			ctx := context.WithValue(context.Background(), returnContextKey, early)
			err := Close(ctx, copiedSession)

			So(err, ShouldNotBeNil)
			So(err, ShouldResemble, errors.New("closing mongo timed out"))
			time.Sleep(500 * time.Millisecond)
		})

		Convey("with context deadline", func() {
			copiedSession := session.Copy()
			go func() {
				_ = slowQueryMongo(copiedSession)
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

func TestSuccessfulMongoDates(t *testing.T) {
	session = nil
	if _, err := setupSession(); err != nil {
		log.Info("mongo instance not available, skip tests", log.Data{"error": err})
		t.FailNow()
	}

	if err := setUpTestData(session.Copy()); err != nil {
		log.ErrorC("Failed to insert test data, skipping tests", err, nil)
		t.FailNow()
	}

	Convey("WithUpdates adds both fields", t, func() {

		Convey("check data in original state", func() {

			res := TestModel{}

			err := queryMongo(session.Copy(), bson.M{"_id": "1"}, &res)
			So(err, ShouldBeNil)
			So(res, ShouldResemble, TestModel{State: "first"})

		})

		Convey("check data after plain Update", func() {

			res := TestModel{}

			err := session.DB(Database).C(Collection).Update(bson.M{"_id": "1"}, bson.M{"$set": bson.M{"new_key": 123}})
			So(err, ShouldBeNil)

			err = queryMongo(session.Copy(), bson.M{"_id": "1"}, &res)
			So(err, ShouldBeNil)
			So(res, ShouldResemble, TestModel{State: "first", NewKey: 123})

		})

		Convey("check data with Update with new dates", func() {

			testStartTime := time.Now().Truncate(time.Second)
			res := TestModel{}

			update := bson.M{"$set": bson.M{"new_key": 321}}
			updateWithTimestamps := WithUpdates(update)
			So(updateWithTimestamps, ShouldResemble, bson.M{"$currentDate": bson.M{"last_updated": true, "unique_timestamp": bson.M{"$type": "timestamp"}}, "$set": bson.M{"new_key": 321}})

			err := session.DB(Database).C(Collection).Update(bson.M{"_id": "1"}, updateWithTimestamps)
			So(err, ShouldBeNil)

			err = queryMongo(session.Copy(), bson.M{"_id": "1"}, &res)
			So(err, ShouldBeNil)
			So(res.State, ShouldEqual, "first")
			So(res.NewKey, ShouldEqual, 321)
			So(res.LastUpdated, ShouldHappenOnOrAfter, testStartTime)
			// extract time part
			So(res.UniqueTimestamp<<32, ShouldBeGreaterThanOrEqualTo, testStartTime.Unix())

		})

		Convey("check data with Update with new Namespaced dates", func() {

			// ensure this testStartTime is greater than last
			time.Sleep(1010 * time.Millisecond)
			testStartTime := time.Now().Truncate(time.Second)
			res := testNamespacedModel{}

			update := bson.M{"$set": bson.M{"new_key": 1234}}
			updateWithTimestamps := WithNamespacedUpdates(update, []string{"nixed.", "currant."})
			So(updateWithTimestamps, ShouldResemble, bson.M{
				"$currentDate": bson.M{
					"currant.last_updated":     true,
					"currant.unique_timestamp": bson.M{"$type": "timestamp"},
					"nixed.last_updated":       true,
					"nixed.unique_timestamp":   bson.M{"$type": "timestamp"},
				},
				"$set": bson.M{"new_key": 1234},
			})

			err := session.DB(Database).C(Collection).Update(bson.M{"_id": "1"}, updateWithTimestamps)
			So(err, ShouldBeNil)

			err = queryNamespacedMongo(session.Copy(), bson.M{"_id": "1"}, &res)
			So(err, ShouldBeNil)
			So(res.State, ShouldEqual, "first")
			So(res.NewKey, ShouldEqual, 1234)
			So(res.Currant.LastUpdated, ShouldHappenOnOrAfter, testStartTime)
			So(res.Nixed.LastUpdated, ShouldHappenOnOrAfter, testStartTime)
			// extract time part
			So(res.Currant.UniqueTimestamp<<32, ShouldBeGreaterThanOrEqualTo, testStartTime.Unix())
			So(res.Nixed.UniqueTimestamp<<32, ShouldBeGreaterThanOrEqualTo, testStartTime.Unix())

		})

	})

	if err := cleanupTestData(session.Copy()); err != nil {
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

func slowQueryMongo(session *mgo.Session) error {
	defer session.Close()

	_, err := session.DB(Database).C(Collection).Find(bson.M{"$where": "sleep(2000) || true"}).Count()
	if err != nil {
		return err
	}

	return nil
}

func queryMongo(session *mgo.Session, query bson.M, res *TestModel) error {
	defer session.Close()

	if err := session.DB(Database).C(Collection).Find(query).One(&res); err != nil {
		return err
	}

	return nil
}

func queryNamespacedMongo(session *mgo.Session, query bson.M, res *testNamespacedModel) error {
	defer session.Close()

	if err := session.DB(Database).C(Collection).Find(query).One(&res); err != nil {
		return err
	}

	return nil
}

func getTestData() []bson.M {
	return []bson.M{
		bson.M{
			"_id":   "1",
			"state": "first",
		},
		bson.M{
			"_id":   "2",
			"state": "second",
		},
	}
}

func setUpTestData(session *mgo.Session) error {
	defer session.Close()

	if _, err := session.DB(Database).C(Collection).Upsert(bson.M{"_id": "1"}, getTestData()[0]); err != nil {
		return err
	}

	if _, err := session.DB(Database).C(Collection).Upsert(bson.M{"_id": "2"}, getTestData()[1]); err != nil {
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
