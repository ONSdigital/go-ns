package avro

import (
	"reflect"
	"testing"

	"github.com/go-avro/avro"
	. "github.com/smartystreets/goconvey/convey"
)

type stringMap map[string]string

func TestUnitMarshal(t *testing.T) {
	Convey("Nested objects", t, func() {
		schema := &Schema{
			Definition: nestedObjectSchema,
		}

		data := &NestedTestData{
			Team: "Tottenham",
			Footballer: FootballerName{
				Surname:  "Kane",
				Forename: "Harry",
				AKA:      map[string]string{"Hurricane": "positive"},
			},
			AKA: map[string]string{
				"Spurs":          "team name",
				"The Lilywhites": "another team name",
			},
			Silverware: map[string]string{"FA Cup": "1900-01"},
			Stats:      int32(10),
		}

		bufferBytes, err := schema.Marshal(data)
		So(err, ShouldBeNil)
		So(bufferBytes, ShouldNotBeNil)

		var footballMessage NestedTestData
		exampleEvent := &Schema{
			Definition: nestedObjectSchema,
		}

		err = exampleEvent.Unmarshal(bufferBytes, &footballMessage)
		So(err, ShouldBeNil)
		So(footballMessage.Team, ShouldEqual, "Tottenham")
		So(footballMessage.Footballer.Surname, ShouldEqual, "Kane")
		So(footballMessage.Footballer.Forename, ShouldEqual, "Harry")
		So(footballMessage.Footballer.AKA, ShouldResemble, map[string]string{"Hurricane": "positive"})
		So(footballMessage.AKA, ShouldResemble, map[string]string{"Spurs": "team name", "The Lilywhites": "another team name"})
		So(footballMessage.Silverware, ShouldResemble, map[string]string{"FA Cup": "1900-01"})
		So(footballMessage.Stats, ShouldEqual, int32(10))
	})

	Convey("Nested object empty", t, func() {
		schema := &Schema{
			Definition: nestedObjectSchema,
		}

		data := &NestedTestData{
			Team:       "Tottenham",
			Footballer: FootballerName{},
		}

		bufferBytes, err := schema.Marshal(data)
		So(err, ShouldBeNil)
		So(bufferBytes, ShouldNotBeNil)

		var footballMessage NestedTestData
		exampleEvent := &Schema{
			Definition: nestedObjectSchema,
		}

		err = exampleEvent.Unmarshal(bufferBytes, &footballMessage)
		So(err, ShouldBeNil)
		So(footballMessage.Team, ShouldEqual, "Tottenham")
		So(footballMessage.Footballer, ShouldResemble, FootballerName{})
		So(footballMessage.Stats, ShouldEqual, 0)
		// Note: AKA was empty, so remains empty (no "null" default)
		So(footballMessage.AKA, ShouldNotBeNil)
		So(footballMessage.AKA, ShouldResemble, map[string]string{})
		// Note: Silverware was empty, but has a default of nil
		So(footballMessage.Silverware, ShouldBeNil)
	})

	Convey("Successfully marshal data", t, func() {
		schema := &Schema{
			Definition: testSchema,
		}

		data := &testData{
			Manager:         "Pardew, Alan",
			TeamName:        "Crystal Palace FC",
			Owner:           "Long, Martin",
			Sport:           "Football",
			URI:             "http://8080/cpfc.com",
			HasChangedName:  false,
			NumberOfPlayers: int32(24),
			PayPerWeek:      int64(539457394875390485),
		}

		bufferBytes, err := schema.Marshal(data)
		So(err, ShouldBeNil)
		So(bufferBytes, ShouldNotBeNil)
	})

	Convey("Successfully marshal data missing uri", t, func() {
		schema := &Schema{
			Definition: testSchema,
		}

		data := &testData1{
			Manager:         "Pardew, Alan",
			HasChangedName:  false,
			NumberOfPlayers: int32(24),
			URI:             "http://8080/cpfc.com",
		}

		bufferBytes, err := schema.Marshal(data)
		So(err, ShouldBeNil)
		So(bufferBytes, ShouldNotBeNil)
	})

	Convey("Marshal should return an error unless given a pointer to a struct", t, func() {
		schema := &Schema{
			Definition: testSchema,
		}

		data := "string"
		bufferBytes, err := schema.Marshal(data)
		So(err, ShouldNotBeNil)
		So(err, ShouldHaveSameTypeAs, ErrUnsupportedType(reflect.ValueOf(data).Kind()))
		So(bufferBytes, ShouldBeNil)
	})

	Convey("Marshal should return an error if field is of the wrong type", t, func() {
		schema := &Schema{
			Definition: testSchema,
		}

		data := &testData2{
			Manager:        "Pardew, Alan",
			NumberOfYouths: 10,
		}

		bufferBytes, err := schema.Marshal(data)
		So(err, ShouldNotBeNil)
		So(err, ShouldEqual, ErrUnsupportedFieldType)
		So(bufferBytes, ShouldBeNil)
	})
}

// Test checkFieldType function
func TestUnitCheckFieldType(t *testing.T) {
	Convey("Successfully return without error", t, func() {
		_, v, t := setUp(testSchema, 1)

		err := checkFieldType(v, t)
		So(err, ShouldBeNil)
	})

	Convey("Error on unsupported field type", t, func() {
		_, v, t := setUp(testSchema, 2)

		err := checkFieldType(v, t)
		So(err, ShouldNotBeNil)
		So(err, ShouldResemble, ErrUnsupportedFieldType)
	})
}

// Test isValidType function
func TestUnitIsValidType(t *testing.T) {
	Convey("Return true for supported types", t, func() {
		Convey("returned true for boolean type", func() {
			isValid := isValidType(reflect.Bool)
			So(isValid, ShouldEqual, true)
		})

		Convey("returned true for int32 type", func() {
			isValid := isValidType(reflect.Int32)
			So(isValid, ShouldEqual, true)
		})

		Convey("returned true for int64 type", func() {
			isValid := isValidType(reflect.Int64)
			So(isValid, ShouldEqual, true)
		})

		Convey("returned true for slice type", func() {
			isValid := isValidType(reflect.Slice)
			So(isValid, ShouldEqual, true)
		})

		Convey("returned true for string type", func() {
			isValid := isValidType(reflect.String)
			So(isValid, ShouldEqual, true)
		})

		Convey("returned true for map type", func() {
			isValid := isValidType(reflect.Map)
			So(isValid, ShouldEqual, true)
		})

		Convey("returned true for struct type", func() {
			isValid := isValidType(reflect.Struct)
			So(isValid, ShouldEqual, true)
		})
	})

	Convey("Return false for unsupported type", t, func() {
		isValid := isValidType(reflect.Float32)
		So(isValid, ShouldEqual, false)
	})
}

// Test getRecord function
func TestUnitGetRecord(t *testing.T) {
	Convey("Successfully return a generic avro record", t, func() {
		Convey("record generated without a nested object", func() {
			avroSchema, v, typ := setUp(testSchema, 1)

			record, err := getRecord(avroSchema, v, typ)
			So(err, ShouldBeNil)
			So(record, ShouldNotBeNil)
			So(record, ShouldHaveSameTypeAs, avro.NewGenericRecord(avroSchema))
		})

		Convey("record generated with array of strings", func() {
			avroSchema, v, typ := setUp(testArraySchema, 3)

			record, err := getRecord(avroSchema, v, typ)
			So(err, ShouldBeNil)
			So(record, ShouldNotBeNil)
			So(record, ShouldHaveSameTypeAs, avro.NewGenericRecord(avroSchema))
			So(record.Get("winning_years"), ShouldResemble, []interface{}{"1934", "1972", "1999"})
		})

		Convey("record generated with missing array of strings", func() {
			avroSchema, v, typ := setUp(testArraySchema, 5)

			record, err := getRecord(avroSchema, v, typ)
			So(err, ShouldBeNil)
			So(record, ShouldNotBeNil)
			So(record, ShouldHaveSameTypeAs, avro.NewGenericRecord(avroSchema))
			So(record.Get("winning_years"), ShouldResemble, []interface{}(nil))
		})

		Convey("record generated with nested array", func() {
			avroSchema, v, typ := setUp(testNestedArraySchema, 4)

			record, err := getRecord(avroSchema, v, typ)

			So(err, ShouldBeNil)
			So(record, ShouldNotBeNil)
			So(record, ShouldHaveSameTypeAs, avro.NewGenericRecord(avroSchema))
			So(record.Get("team"), ShouldResemble, "Doncaster")
			So(record.Get("footballers"), ShouldNotBeEmpty)
		})
	})
}

// Test marshalSlice function
func TestUnitMarshalSlice(t *testing.T) {
	Convey("Successfully marshal string slice", t, func() {
		avroSchema, v, typ := setUp(testArraySchema, 3)
		record, _ := getRecord(avroSchema, v, typ)

		err := marshalSlice(record, v, 0, "string_slice", avroSchema)
		So(err, ShouldBeNil)
	})

	Convey("Successfully marshal struct slice", t, func() {
		avroSchema, v, typ := setUp(testNestedArraySchema, 4)
		record, _ := getRecord(avroSchema, v, typ)

		err := marshalSlice(record, v, 1, "footballers", avroSchema)
		So(err, ShouldBeNil)
	})

	Convey("Fail to marshal struct slice", t, func() {
		avroSchema, v, typ := setUp(testNestedArraySchema, 4)
		record, _ := getRecord(avroSchema, v, typ)

		err := marshalSlice(record, v, 1, "footballer", avroSchema)
		So(err, ShouldNotBeNil)
		So(err, ShouldResemble, ErrMissingNestedScema)
	})
}

// Test marshalStringSlice function
func TestUnitMarshalStringSlice(t *testing.T) {
	Convey("Successfully marshal string slice", t, func() {
		_, v, _ := setUp(testArraySchema, 3)

		slice := marshalStringSlice(v, 0)
		So(slice, ShouldNotBeEmpty)
		So(slice, ShouldResemble, []interface{}{"1934", "1972", "1999"})
	})
}

// Test marshalStructSlice function
func TestUnitMarshalStructSlice(t *testing.T) {
	Convey("Successfully marshal struct slice", t, func() {
		avroSchema, v, _ := setUp(testNestedArraySchema, 4)

		slice, err := marshalStructSlice(v, 1, avroSchema, "footballers")
		So(err, ShouldBeNil)
		So(slice, ShouldNotBeEmpty)
	})

	Convey("Throws error due to missing nested schema", t, func() {
		avroSchema, v, _ := setUp(testNestedArraySchema, 4)

		slice, err := marshalStructSlice(v, 1, avroSchema, "footballer")
		So(err, ShouldNotBeNil)
		So(err, ShouldResemble, ErrMissingNestedScema)
		So(slice, ShouldBeNil)
	})
}

// Test getArraySchema function
func TestUnitGetArraySchema(t *testing.T) {
	avroSchema, _, _ := setUp(testSchema, 0)

	Convey("Successfully retrieve array schema", t, func() {
		newSchema, err := getArraySchema(avroSchema, "manager")
		So(err, ShouldBeNil)
		So(newSchema, ShouldNotBeEmpty)
	})

	Convey("Error retrieving array schema", t, func() {
		newSchema, err := getArraySchema(avroSchema, "managers")
		So(err, ShouldNotBeNil)
		So(err, ShouldResemble, ErrMissingNestedScema)
		So(newSchema, ShouldBeNil)
	})
}

// Test getNestedSchema function
func TestUnitGetNestedSchema(t *testing.T) {
	Convey("Successfully retrieve nested schema", t, func() {
		avroSchema, v, t := setUp(testNestedArraySchema, 4)

		newSchema, err := getNestedSchema(avroSchema, "footballers", v, t)
		So(err, ShouldBeNil)
		So(newSchema, ShouldNotBeEmpty)
	})

	Convey("Error retrieving nested schema", t, func() {
		avroSchema, v, t := setUp(testNestedArraySchema, 4)

		newSchema, err := getNestedSchema(avroSchema, "footballer", v, t)
		So(err, ShouldNotBeNil)
		So(err, ShouldResemble, ErrMissingNestedScema)
		So(newSchema, ShouldBeNil)
	})
}

// TODO Test Unmarshal function

// TODO -- Test populateStructFromSchema function

// TODO -- -- Test generateDecodedRecord function

// TODO -- -- Test unmarshalStringSlice function

// TODO -- -- Test unmarshalStructSlice function

// TODO -- -- -- Test populateNestedArrayItem function

// TODO -- -- -- -- Test setNestedStructs function

func setUp(testSchema string, dataSet int) (avro.Schema, reflect.Value, reflect.Type) {
	avroSchema, _ := avro.ParseSchema(testSchema)

	var (
		v   reflect.Value
		typ reflect.Type
	)

	switch dataSet {
	case 1:
		testData := &testData1{
			Manager:         "Pardew, Alan",
			HasChangedName:  false,
			NumberOfPlayers: int32(24),
			PayPerWeek:      int64(539457394875390485),
		}

		v = reflect.ValueOf(testData)
		typ = reflect.TypeOf(testData)
	case 2:
		testData := &testData2{
			Manager:         "Pardew, Alan",
			HasChangedName:  false,
			NumberOfPlayers: int32(24),
			NumberOfYouths:  6,
		}

		v = reflect.ValueOf(testData)
		typ = reflect.TypeOf(testData)
	case 3:
		var winningYears []string
		winningYears = append(winningYears, "1934", "1972", "1999")
		testData := &testData3{
			WinningYears: winningYears,
		}

		v = reflect.ValueOf(testData)
		typ = reflect.TypeOf(testData)
	case 4:
		testData := &testData4{
			Team: "Doncaster",
			Footballers: []Footballer{
				{
					Email: "jgregory@gmail.com",
					Name:  "jack gregory",
				},
				{
					Email: "pdoherty@gmail.com",
					Name:  "paul doherty",
				},
			},
		}

		v = reflect.ValueOf(testData)
		typ = reflect.TypeOf(testData)
	case 5:
		var winningYears []string
		testData := &testData5{
			WinningYears: winningYears,
		}

		v = reflect.ValueOf(testData)
		typ = reflect.TypeOf(testData)
	default:
		testData := &testData1{
			Manager: "Pardew, Alan",
		}

		v = reflect.ValueOf(testData)
		typ = reflect.TypeOf(testData)
	}

	v = reflect.Indirect(v)
	typ = typ.Elem()

	return avroSchema, v, typ
}
