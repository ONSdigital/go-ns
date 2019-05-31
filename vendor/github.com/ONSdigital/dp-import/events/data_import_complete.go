package events

import "github.com/ONSdigital/go-ns/avro"

// DataImportComplete contains the data required for each hierarchy to be imported.
type DataImportComplete struct {
	InstanceID    string `avro:"instance_id"`
	DimensionName string `avro:"dimension_name"`
	CodeListID    string `avro:"code_list_id"`
}

// from dp-import-tracker
var dataImportCompleteSchema = `{
  "type": "record",
  "name": "data-import-complete",
  "fields": [
    {"name": "instance_id", "type": "string"},
    {"name": "dimension_name", "type": "string"},
    {"name": "code_list_id", "type": "string"}
  ]
}`

// DataImportCompleteSchema is the Avro schema for DataImportComplete
var DataImportCompleteSchema = avro.Schema{
	Definition: dataImportCompleteSchema,
}
