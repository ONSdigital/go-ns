package errorschema

import "github.com/ONSdigital/go-ns/avro"

var reportedEvent = `{
  "type": "record",
  "name": "report-event",
  "fields": [
    {"name": "instance_id", "type": "string"},
    {"name": "event_type", "type": "string"},
    {"name": "event_message", "type": "string"}
  ]
}`

var ReportedEventSchema *avro.Schema = &avro.Schema{
	Definition: reportedEvent,
}
