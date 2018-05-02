package audit

import "github.com/ONSdigital/go-ns/avro"

var event = `{
  "type": "record",
  "name": "audit-event",
  "namespace": "",
  "fields": [
    {
      "type": "string",
      "name": "created",
      "default": ""
    },
    {
      "name": "service",
      "type": "string",
      "default": ""
    },
    {
      "name": "request_id",
      "type": "string",
      "default": ""
    },
    {
      "name": "user",
      "type": "string",
      "default": ""
    },
    {
      "name": "attempted_action",
      "type": "string",
      "default": ""
    },
    {
      "name": "action_result",
      "type": "string",
      "default": ""
    },
    {
      "name": "params",
      "default": null,
      "type": [
        "null",
        {
          "type": "map",
          "values": "string"
        }
      ]
    }
  ]
}`

// EventSchema defines the avro schema for an audit event.
var EventSchema *avro.Schema = &avro.Schema{
	Definition: event,
}
