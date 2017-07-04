// Package avro provides a user functionality to return the
// avro encoding of s.
package avro

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"

	avro "github.com/go-avro"
	"github.com/ian-kent/go-log/log"
)

// Schema contains the schema definition necessary to generate an avro record
type Schema struct {
	Definition string
}

// ErrUnsupportedType is returned if the interface isn't a
// pointer to a struct
func ErrUnsupportedType(typ reflect.Kind) error {
	return fmt.Errorf("Unsupported interface type: %v", typ)
}

// ErrInvalidFieldName returned when the field name does not exist in avro schema
func ErrInvalidFieldName(field string) error {
	return errors.New("Incorrect field " + field + ", unable to get value for field")
}

// ErrUnsupportedFieldType is returned for unsupported field types.
var ErrUnsupportedFieldType = errors.New("Unsupported field type")

// Marshal is used to avro encode the interface of s.
func (schema *Schema) Marshal(s interface{}) ([]byte, error) {
	v := reflect.ValueOf(s)

	if v.Kind() == reflect.PtrTo(reflect.TypeOf(s)).Kind() {
		v = reflect.Indirect(v)
	}

	// Only structs are supported so return an empty result if the passed object
	// isn't a struct.
	if v.Kind() != reflect.Struct {
		return nil, ErrUnsupportedType(v.Kind())
	}

	//If a pointer to a struct is passed, get the type of the dereferenced object.
	typ := reflect.TypeOf(s)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	//Check for unsupported interface types
	err := checkFieldType(v, typ)
	if err != nil {
		return nil, err
	}

	avroSchema, err := avro.ParseSchema(schema.Definition)
	if err != nil {
		return nil, err
	}

	record := avro.NewGenericRecord(avroSchema)

	for i := 0; i < v.NumField(); i++ {
		fieldTag := typ.Field(i).Tag.Get("avro")
		if fieldTag == "-" {
			continue
		}
		fieldName := typ.Field(i).Name

		switch typ.Field(i).Type.Kind() {
		case reflect.Bool:
			value := v.FieldByName(fieldName).Bool()
			record.Set(fieldTag, value)
		case reflect.String:
			value := v.FieldByName(fieldName).String()
			record.Set(fieldTag, value)
		}
	}

	writer := avro.NewGenericDatumWriter()
	writer.SetSchema(avroSchema)

	buffer := new(bytes.Buffer)
	encoder := avro.NewBinaryEncoder(buffer)

	err = writer.Write(record, encoder)
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

// Unmarshal is used to parse the avro encoded data and store the
// result in the value pointed to by s.
func (schema *Schema) Unmarshal(message []byte, s interface{}) error {
	v := reflect.ValueOf(s)
	vp := reflect.ValueOf(s)

	if v.Kind() == reflect.PtrTo(reflect.TypeOf(s)).Kind() {
		v = reflect.Indirect(v)
	}

	// Only structs are supported so return an empty result if the passed object
	// isn't a struct.
	if v.Kind() != reflect.Struct {
		return ErrUnsupportedType(v.Kind())
	}

	decodedRecord, err := generateRecord(schema.Definition, message)
	if err != nil {
		log.Error(err, nil)
		return err
	}

	// If a pointer to a struct is passed, get the type of the dereferenced object.
	typ := reflect.TypeOf(s)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	for i := 0; i < v.NumField(); i++ {
		field := typ.Field(i).Tag.Get("avro")
		if field == "-" {
			continue
		}
		rawFieldName := typ.Field(i).Name
		fieldName := vp.Elem().FieldByName(rawFieldName)

		value := decodedRecord.Get(field)

		if fieldName.IsValid() {
			fieldName.Set(reflect.ValueOf(value))
		}
	}

	return nil
}

func checkFieldType(v reflect.Value, t reflect.Type) error {
	for i := 0; i < v.NumField(); i++ {
		fieldTag := t.Field(i).Tag.Get("avro")
		if fieldTag == "-" {
			continue
		}
		fieldType := t.Field(i)

		if fieldType.Type.Kind() != reflect.String && fieldType.Type.Kind() != reflect.Bool {
			return ErrUnsupportedFieldType
		}
	}

	return nil
}

func generateRecord(schema string, message []byte) (*avro.GenericRecord, error) {
	avroSchema, err := avro.ParseSchema(schema)
	if err != nil {
		return nil, err
	}

	reader := avro.NewGenericDatumReader()
	reader.SetSchema(avroSchema)
	decoder := avro.NewBinaryDecoder(message)
	decodedRecord := avro.NewGenericRecord(avroSchema)

	err = reader.Read(decodedRecord, decoder)
	if err != nil {
		return nil, err
	}

	return decodedRecord, nil
}
