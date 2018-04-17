package audit

import "github.com/ONSdigital/go-ns/avro"

var event = `
{
  "type": "record",
  "name": "audit-event",
  "namespace": "",
  "fields": [
    {
      "type": "string",
      "name": "service"
    },
    {
      "name": "name_space",
      "type": "string"
    },
    {
      "name": "request_id",
      "type": "string"
    },
    {
      "name": "user",
      "type": "string"
    },
    {
      "name": "attempted_action",
      "type": "string"
    },
    {
      "name": "outcome",
      "type": "string"
    },
    {
      "name": "response_status",
      "type": "string"
    },
    {
      "name": "request_uri",
      "type": "string"
    },
    {
      "name": "request_method",
      "type": "string"
    }
  ]
}`

var EventSchema *avro.Schema = &avro.Schema{
	Definition: event,
}
