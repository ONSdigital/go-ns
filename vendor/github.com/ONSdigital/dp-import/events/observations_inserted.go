package events

import "github.com/ONSdigital/go-ns/avro"

// ObservationsInserted is the data that is output for each observation batch inserted.
type ObservationsInserted struct {
	ObservationsInserted int32  `avro:"observations_inserted"`
	InstanceID           string `avro:"instance_id"`
}

// from dp-observation-importer
var observationsInsertedSchema = `{
  "type": "record",
  "name": "import-observations-inserted",
  "fields": [
    {"name": "instance_id", "type": "string"},
    {"name": "observations_inserted", "type": "int"}
  ]
}`

// ObservationsInsertedSchema is the Avro schema for each observation batch inserted.
var ObservationsInsertedSchema = avro.Schema{
	Definition: observationsInsertedSchema,
}
