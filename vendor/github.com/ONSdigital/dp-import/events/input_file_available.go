package events

import "github.com/ONSdigital/go-ns/avro"

// InputFileAvailable is an event produced when a new input file is available to import.
type InputFileAvailable struct {
	URL        string `avro:"file_url"`
	InstanceID string `avro:"instance_id"`
	JobID      string `avro:"job_id"`
}

var inputFileAvailableSchema = `{
  "type": "record",
  "name": "input-file-available",
  "fields": [
    {"name": "file_url", "type": "string"},
    {"name": "instance_id", "type": "string"},
    {"name": "job_id", "type": "string"}
  ]
}`

// InputFileAvailableSchema provides an Avro schema for the InputFileAvailable event.
var InputFileAvailableSchema = &avro.Schema{
	Definition: inputFileAvailableSchema,
}
