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
      "name": "namespace",
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
      "default": "",
      "type": {
        "type": "map",
        "values": "string"
      }
    }
  ]
}`

var EventSchema *avro.Schema = &avro.Schema{
	Definition: event,
}
