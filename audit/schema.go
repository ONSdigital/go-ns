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
      "name": "created",
      "default": ""
    },
    {
      "type": "string",
      "name": "service",
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
      "name": "result",
      "type": "string",
      "default": ""
    },
    {
      "name": "params",
      "default": [],
      "type": {
        "type": "array",
        "items": {
          "name": "Params",
          "type": "record",
          "fields": [
            {
              "name": "key",
              "type": "string",
              "default": ""
            },
            {
              "name": "value",
              "type": "string",
              "default": ""
            }
          ]
        }
      }
    }
  ]
}`

var EventSchema *avro.Schema = &avro.Schema{
	Definition: event,
}
