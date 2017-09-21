package main

import (
	"context"
	"flag"
	"time"

	"github.com/ONSdigital/go-ns/avro"
	"github.com/ONSdigital/go-ns/kafka"
	"github.com/ONSdigital/go-ns/log"
)

// InputFileAvailableSchema schema
var InputFileAvailableSchema = `{
	"type": "record",
	"name": "input-file-available",
	"fields": [
		{"name": "file_url", "type": "string"},
		{"name": "instance_id", "type": "string"}
	]
}`

type inputFileAvailable struct {
	fileURL    string `avro:"file_url"`
	InstanceID string `avro:"instance_id"`
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

// ObservationsInsertedEvent is the Avro schema for each observation batch inserted.
var ObservationsInsertedEvent avro.Schema = avro.Schema{Definition: observationsInsertedSchema}

type insertedObservationsMessage struct {
	InstanceID           string `avro:"instance_id"`
	ObservationsInserted int32  `avro:"observations_inserted"`
}

/*
var TestNestedArraySchema = `{
	"type": "record",
	"name": "publish-dataset",
	"fields": [
		{
			"name": "instance_ids",
			"type": {
				"type": "array",
				"items": "string"
			}
		},
		{
			"name": "files",
			"type": {
				"type": "array",
				"items": {
					"name": "file",
					"type": "record",
					"fields": [
						{
							"name": "alias-name",
							"type": "string"
						},
						{
							"name": "url",
							"type": "string"
						}
					]
				}
			}
		}
	]
}`

type testNestedArray struct {
	InstanceIDs []string `avro:"instance_ids"`
	Files       []File   `avro:"files"`
}

type File struct {
	AliasName string `avro:"alias-name"`
	URL       string `avro:"url"`
}
*/

func main() {
	instance_id := flag.String("id", "21", "instance id")
	producerTopic := flag.String("topic", "input-file-available", "producer topic")
	fileURL := flag.String("s3", "s3://dp-dimension-extractor/OCIGrowth.csv", "s3 file")
	insertedObservations := flag.Int("inserts", 2000, "inserted observations")
	flag.Parse()

	brokers := []string{"localhost:9092"}
	// inputFileAvailableProducer, err := kafka.NewProducer(brokers, "dimensions-extracted", int(2000000))
	// inputFileAvailableProducer, err := kafka.NewProducer(brokers, "input-test-file", int(2000000))
	producer, err := kafka.NewProducer(brokers, *producerTopic, int(2000000))
	if err != nil {
		panic(err)
	}

	var schema *avro.Schema
	var producerMessage []byte

	if *producerTopic == "input-file-available" {
		schema = &avro.Schema{Definition: InputFileAvailableSchema}
		producerMessage, err = schema.Marshal(&inputFileAvailable{
			//fileURL:    "s3://dp-dimension-extractor/UKBAA01a.csv",
			fileURL:    *fileURL,
			InstanceID: *instance_id,
		})
	} else if *producerTopic == "import-observations-inserted" {
		schema = &avro.Schema{Definition: observationsInsertedSchema}
		insertedObservationsMsg := insertedObservationsMessage{
			InstanceID:           *instance_id,
			ObservationsInserted: int32(*insertedObservations),
		}
		log.Debug("msg", log.Data{"iom": insertedObservationsMsg})
		producerMessage, err = schema.Marshal(&insertedObservationsMsg)
	}

	/* nestStruct := testNestedArray{
		InstanceIDs: []string{
			"1",
			"2",
		},
		Files: []File{
			File{
				AliasName: "some-alias-name",
				URL:       "some-made-up-url",
			},
			File{
				AliasName: "some-other-alias-name",
				URL:       "some-other-made-up-url",
			},
		},
	}

	producerMessage, err := schema.Marshal(nestStruct)
	*/

	if err != nil {
		log.ErrorC("error marshalling", err, nil)
		panic(err)
		// os.Exit(1)
	}

	producer.Output() <- producerMessage
	time.Sleep(time.Duration(1000 * time.Millisecond))
	ctx, cancel := context.WithTimeout(context.Background(), 5)
	producer.Close(ctx)
	cancel()
}
