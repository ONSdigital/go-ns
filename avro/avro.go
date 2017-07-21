// Package avro provides a user functionality to return the
// avro encoding of s.
package avro

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/go-avro/avro"
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

// ErrUnsupportedFieldType is returned for unsupported field types.
var ErrUnsupportedFieldType = errors.New("Unsupported field type")

// ErrMissingNestedScema is returned when nested schemas are missing from the parent
var ErrMissingNestedScema = errors.New("nested schema missing from parent")

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

	// If a pointer to a struct is passed, get the type of the dereferenced object.
	typ := reflect.TypeOf(s)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	// Check for unsupported interface types
	err := checkFieldType(v, typ)
	if err != nil {
		return nil, err
	}

	avroSchema, err := avro.ParseSchema(schema.Definition)
	if err != nil {
		return nil, err
	}

	record, err := getRecord(avroSchema, v, typ)
	if err != nil {
		return nil, err
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

	// If a pointer to a struct is passed, get the type of the dereferenced object.
	typ := reflect.TypeOf(s)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	return populateStructFromSchema(schema.Definition, message, typ, v, vp)
}

func checkFieldType(v reflect.Value, t reflect.Type) error {
	for i := 0; i < v.NumField(); i++ {
		fieldTag := t.Field(i).Tag.Get("avro")
		if fieldTag == "-" {
			continue
		}
		fieldType := t.Field(i)

		if !isValidType(fieldType.Type.Kind()) {
			return ErrUnsupportedFieldType
		}
	}

	return nil
}

func generateDecodedRecord(schema string, message []byte) (*avro.GenericRecord, error) {
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

func getNestedSchema(avroSchema avro.Schema, fieldTag string, v reflect.Value, typ reflect.Type) (avro.Schema, error) {
	// Unmarshal parent avro schema into map
	var schemaMap map[string]interface{}
	if err := json.Unmarshal([]byte(avroSchema.String()), &schemaMap); err != nil {
		return nil, err
	}

	// Get fields section from parent schema map
	fields := schemaMap["fields"].([]interface{})
	for _, field := range fields {
		var fld map[string]interface{}
		var ok bool

		if fld, ok = field.(map[string]interface{}); !ok {
			continue
		}

		// Iterate through each field until the nested schema field is found
		if fld["name"].(string) == fieldTag {
			var avroFieldType map[string]interface{}

			// The nested schema is inside the type element of the required field
			if avroFieldType, ok = fld["type"].(map[string]interface{}); !ok {
				var avroFieldTypes []interface{}
				if avroFieldTypes, ok = fld["type"].([]interface{}); !ok {
					continue
				}

				// If the nested schema could potentially be "null", then the schema is the second type
				// element rather than the first
				if avroFieldType = avroFieldTypes[1].(map[string]interface{}); !ok {
					continue
				}
			}

			// Marshal the nested schema map into json
			nestedSchemaBytes, err := json.Marshal(avroFieldType)
			if err != nil {
				return nil, err
			}

			return avro.ParseSchema(string(nestedSchemaBytes))
		}
	}
	return nil, ErrMissingNestedScema
}

func getRecord(avroSchema avro.Schema, v reflect.Value, typ reflect.Type) (*avro.GenericRecord, error) {
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
		case reflect.Int32:
			value := v.FieldByName(fieldName).Interface().(int32)
			record.Set(fieldTag, value)
		case reflect.Slice:
			if err := marshalSlice(record, v, i, fieldTag, avroSchema); err != nil {
				return nil, err
			}
		case reflect.Struct:
			nestedSchema, err := getNestedSchema(avroSchema, fieldTag, v, typ)
			if err != nil {
				return nil, err
			}

			nestedRecord, err := getRecord(nestedSchema, v.Field(i), typ.Field(i).Type)
			if err != nil {
				return nil, err
			}

			record.Set(fieldTag, nestedRecord)
		}
	}

	return record, nil
}

func isValidType(kind reflect.Kind) bool {
	supportedTypes := []reflect.Kind{
		reflect.Bool,
		reflect.Int32,
		reflect.Slice,
		reflect.String,
		reflect.Struct,
	}

	for _, supportedType := range supportedTypes {
		if supportedType == kind {
			return true
		}
	}
	return false
}

func marshalSlice(record *avro.GenericRecord, v reflect.Value, i int, fieldTag string, avroSchema avro.Schema) error {
	// This switch statement will need to be extended to support other native types,
	// Currently supports strings and structs.
	switch v.Field(i).Type().Elem().Kind() {
	case reflect.String:
		slice := marshalStringSlice(v, i)
		record.Set(fieldTag, slice)
	case reflect.Struct:
		slice, err := marshalStructSlice(v, i, avroSchema, fieldTag)
		if err != nil {
			return err
		}
		record.Set(fieldTag, slice)
	}
	return nil
}

func marshalStringSlice(v reflect.Value, i int) []interface{} {
	vals := v.Field(i)
	var slice []interface{}
	for j := 0; j < vals.Len(); j++ {
		slice = append(slice, vals.Index(j).Interface())
	}
	return slice
}

func marshalStructSlice(v reflect.Value, i int, avroSchema avro.Schema, fieldTag string) ([]interface{}, error) {
	vals := v.Field(i)
	var slice []interface{}
	for j := 0; j < vals.Len(); j++ {
		arraySchema, err := getArraySchema(avroSchema, fieldTag)
		if err != nil {
			return nil, err
		}

		arrayRecord, err := getRecord(arraySchema, vals.Index(j), v.Field(i).Type().Elem())
		if err != nil {
			return nil, err
		}

		slice = append(slice, arrayRecord)
	}
	return slice, nil
}

func getArraySchema(avroSchema avro.Schema, fieldTag string) (avro.Schema, error) {
	var schemaMap map[string]interface{}
	// Unmarshal the parent schema into a map
	if err := json.Unmarshal([]byte(avroSchema.String()), &schemaMap); err != nil {
		return nil, err
	}

	fields := schemaMap["fields"].([]interface{})
	for _, field := range fields {
		var fld map[string]interface{}
		var ok bool

		if fld, ok = field.(map[string]interface{}); !ok {
			continue
		}

		// Iterate through fields in schema until fieldTag matches the name element
		// of the field
		if fld["name"].(string) == fieldTag {
			// The array schema will be inside the type element of the requested field.
			// Marshal this schema back to json and return
			arraySchemaBytes, err := json.Marshal(fld["type"])
			if err != nil {
				return nil, err
			}

			return avro.ParseSchema(string(arraySchemaBytes))
		}
	}

	return nil, ErrMissingNestedScema
}

func populateNestedArrayItem(nestedMap map[string]interface{}, typ reflect.Type) reflect.Value {
	// Create a new instance of required struct type
	v := reflect.Indirect(reflect.New(typ))
	for i := 0; i < v.NumField(); i++ {
		field := typ.Field(i).Tag.Get("avro")
		fieldValue := nestedMap[field]
		if fieldValue != nil {
			if v.Field(i).Kind() == reflect.Struct {
				setNestedStructs(fieldValue.(map[string]interface{}), v.Field(i), typ.Field(i).Type)
				continue
			}
			value := reflect.ValueOf(fieldValue)
			if typ.Field(i).Type.Kind() == reflect.Slice {
				sliceInterface := fieldValue.([]interface{})
				sliceString := make([]string, len(sliceInterface))
				for _, val := range sliceInterface {
					sliceString = append(sliceString, val.(string))
				}
				value = reflect.ValueOf(sliceString)
			}
			v.Field(i).Set(value)
		}
	}
	return v
}

func populateStructFromSchema(schema string, message []byte, typ reflect.Type, v, vp reflect.Value) error {
	decodedRecord, err := generateDecodedRecord(schema, message)
	if err != nil {
		return err
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
			if v.Field(i).Type().Kind() == reflect.Slice {
				switch v.Field(i).Type().Elem().Kind() {
				case reflect.String:
					value = unmarshalStringSlice(value)
				case reflect.Struct:
					v, err = unmarshalStructSlice(value, v, i)
					if err != nil {
						return err
					}
					continue
				default:
					return ErrUnsupportedType(v.Field(i).Type().Elem().Kind())
				}
			}
			fieldName.Set(reflect.ValueOf(value))
		}
	}

	return nil
}

func setNestedStructs(nestedMap map[string]interface{}, v reflect.Value, typ reflect.Type) {
	for i := 0; i < v.NumField(); i++ {
		field := typ.Field(i).Tag.Get("avro")
		fieldValue := nestedMap[field]
		if fieldValue != nil {
			if v.Field(i).Kind() == reflect.Struct {
				setNestedStructs(fieldValue.(map[string]interface{}), v.Field(i), typ.Field(i).Type)
				continue
			}
			value := reflect.ValueOf(fieldValue)
			if typ.Field(i).Type.Kind() == reflect.Slice {
				sliceInterface := fieldValue.([]interface{})
				sliceString := make([]string, len(sliceInterface))
				for i, val := range sliceInterface {
					sliceString[i] = val.(string)
				}
				value = reflect.ValueOf(sliceString)
			}
			v.Field(i).Set(value)
		}
	}
}

func unmarshalStringSlice(value interface{}) []string {
	sliceInterface := value.([]interface{})
	sliceString := make([]string, len(sliceInterface))
	for _, val := range sliceInterface {
		sliceString = append(sliceString, val.(string))
	}
	return sliceString
}

func unmarshalStructSlice(value interface{}, v reflect.Value, i int) (reflect.Value, error) {
	sliceInterface := value.([]interface{})
	sliceType := v.Field(i).Type().Elem()
	emptySlice := reflect.MakeSlice(reflect.SliceOf(sliceType), 0, 0)
	for _, val := range sliceInterface {
		record := val.(*avro.GenericRecord)

		var dataMap map[string]interface{}
		if err := json.Unmarshal([]byte(record.String()), &dataMap); err != nil {
			return v, err
		}
		item := populateNestedArrayItem(dataMap, v.Field(i).Type().Elem())
		emptySlice = reflect.Append(emptySlice, item)
		v.Field(i).Set(emptySlice)
	}
	return v, nil
}
