package avro

import (
	"reflect"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

var testSchema1 = `{ "type": "record",
 "name": "example1",
 "namespace": "correct",
 "fields": [
    {"name": "manager", "type": "string"},
    {"name": "team_name", "type": "string"},
    {"name": "ownerOfTeam", "type": "string"},
    {"name": "kind-of-sport", "type": "string"},
    {"name": "uri", "type": "string", "default": ""},
    {"name": "has_changed_name", "type": "boolean"},
    {"name": "number_of_players", "type": "int"}
 ]
}`

type testData1 struct {
	Manager         string `avro:"manager"`
	TeamName        string `avro:"team_name"`
	Owner           string `avro:"ownerOfTeam"`
	Sport           string `avro:"kind-of-sport"`
	Uri             string `avro:"uri"`
	HasChangedName  bool   `avro:"has_changed_name"`
	NumberOfPlayers int32  `avro:"number_of_players"`
}

type testData2 struct {
	Manager         string `avro:"manager"`
	TeamName        string `avro:"team_name"`
	Owner           string `avro:"ownerOfTeam"`
	Sport           string `avro:"kind-of-sport"`
	Uri             string `avro:"-"`
	HasChangedName  bool   `avro:"has_changed_name"`
	NumberOfPlayers int32  `avro:"number_of_players"`
}

type testData3 struct {
	Manager         string `avro:"manager"`
	NumberOfPlayers int64  `avro:"number_of_players"`
}

func TestUnitMarshal(t *testing.T) {
	Convey("Successfully marshal data", t, func() {
		incs := &Schema{
			Definition: testSchema1,
		}

		td1a := &testData1{
			Manager:         "Pardew, Alan",
			TeamName:        "Crystal Palace FC",
			Owner:           "Long, Martin",
			Sport:           "Football",
			HasChangedName:  false,
			NumberOfPlayers: int32(24),
		}

		bufferBytes1a, err1a := incs.Marshal(td1a)
		So(err1a, ShouldBeNil)
		So(bufferBytes1a, ShouldNotBeNil)
	})

	Convey("Successfully marshal data missing uri", t, func() {
		incs := &Schema{
			Definition: testSchema1,
		}

		td1a := &testData2{
			Manager:         "Pardew, Alan",
			TeamName:        "Crystal Palace FC",
			Owner:           "Long, Martin",
			Sport:           "Football",
			HasChangedName:  false,
			NumberOfPlayers: int32(24),
		}

		bufferBytes1a, err := incs.Marshal(td1a)
		So(err, ShouldBeNil)
		So(bufferBytes1a, ShouldNotBeNil)
	})

	Convey("Marshal should return an error unless given a pointer to a struct", t, func() {
		cs := &Schema{
			Definition: testSchema1,
		}

		td1b := "string"
		bufferBytes1b, err1b := cs.Marshal(td1b)
		So(err1b, ShouldNotBeNil)
		So(err1b, ShouldHaveSameTypeAs, ErrUnsupportedType(reflect.ValueOf(td1b).Kind()))
		So(bufferBytes1b, ShouldBeNil)
	})

	Convey("Marshal should return an error if field is of the wrong type", t, func() {
		incs := &Schema{
			Definition: testSchema1,
		}

		id := &testData3{
			Manager:         "Pardew, Alan",
			NumberOfPlayers: int64(10),
		}

		bufferBytes2, err2 := incs.Marshal(id)
		So(err2, ShouldNotBeNil)
		So(err2, ShouldEqual, ErrUnsupportedFieldType)
		So(bufferBytes2, ShouldBeNil)
	})
}

func TestUnitUnmarshal(t *testing.T) {
	Convey("Correctly unmarshal byte array", t, func() {
		message, err := createMessage(testSchema1)
		So(err, ShouldBeNil)

		cs := &Schema{
			Definition: testSchema1,
		}

		var data testData2

		err1 := cs.Unmarshal(message, &data)
		So(err1, ShouldBeNil)
		So(data.Manager, ShouldNotBeNil)
		So(data.Manager, ShouldEqual, "John Elway")
		So(data.TeamName, ShouldNotBeNil)
		So(data.TeamName, ShouldEqual, "Denver Broncos")
		So(data.Owner, ShouldNotBeNil)
		So(data.Owner, ShouldEqual, "Pat Bowlen")
		So(data.Sport, ShouldNotBeNil)
		So(data.Sport, ShouldEqual, "American Football")
		So(data.Uri, ShouldNotBeNil)
		So(data.Uri, ShouldEqual, "")
		So(data.HasChangedName, ShouldNotBeNil)
		So(data.HasChangedName, ShouldEqual, false)
	})

	Convey("Check error return for unsupported interface types", t, func() {
		message, err := createMessage(testSchema1)
		So(err, ShouldBeNil)

		cs := &Schema{
			Definition: testSchema1,
		}

		data := ""
		reflectData := reflect.ValueOf(data)

		err1 := cs.Unmarshal(message, data)
		So(err1, ShouldNotBeNil)
		So(err1, ShouldResemble, ErrUnsupportedType(reflectData.Kind()))
	})
}

func createMessage(schema string) ([]byte, error) {
	marshalSchema := &Schema{
		Definition: testSchema1,
	}

	data := &testData1{
		Manager:         "John Elway",
		TeamName:        "Denver Broncos",
		Owner:           "Pat Bowlen",
		Sport:           "American Football",
		HasChangedName:  false,
		Uri:             "http://denverbroncos.com",
		NumberOfPlayers: 11,
	}

	message, err := marshalSchema.Marshal(data)
	if err != nil {
		return nil, err
	}

	return message, nil
}
