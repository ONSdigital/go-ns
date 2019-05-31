package events

import "github.com/ONSdigital/go-ns/avro"

// HierarchyBuilt contains data related to a hierarchy that has just been built.
type HierarchyBuilt struct {
	InstanceID    string `avro:"instance_id"`
	DimensionName string `avro:"dimension_name"`
}

// from dp-hierarchy-builder
var hierarchyBuiltSchema = `{
  "type": "record",
  "name": "hierarchy-built",
  "fields": [
    {"name": "instance_id", "type": "string"},
    {"name": "dimension_name", "type": "string"}
  ]
}`

// HierarchyBuiltSchema is the Avro schema for HierarchyBuilt events.
var HierarchyBuiltSchema = avro.Schema{
	Definition: hierarchyBuiltSchema,
}
