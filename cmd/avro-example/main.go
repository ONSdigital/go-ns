package main

import (
	"context"
	"flag"
	"os"
	"time"

	"github.com/ONSdigital/dp-import/events"
	"github.com/ONSdigital/go-ns/avro"
	"github.com/ONSdigital/go-ns/kafka"
	"github.com/ONSdigital/go-ns/log"
)

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
	var (
		instance_id          = flag.String("id", "21", "instance id")
		producerTopic        = flag.String("topic", "input-file-available", "producer topic")
		fileURL              = flag.String("s3", "s3://dp-dimension-extractor/OCIGrowth.csv", "s3 file")
		insertedObservations = flag.Int("inserts", 2000, "inserted observations")
		dimensionName        = flag.String("dimension", "dimName", "dimension name")
		codeListID           = flag.String("codelistid", "codlistId", "code list ID")
	)
	flag.Parse()

	brokers := []string{"localhost:9092"}
	producer, err := kafka.NewProducer(brokers, *producerTopic, int(2000000))
	if err != nil {
		panic(err)
	}

	var schema *avro.Schema
	var producerMessage []byte

	if *producerTopic == "input-file-available" {
		schema = events.InputFileAvailableSchema
		producerMessage, err = schema.Marshal(&events.InputFileAvailable{
			URL:        *fileURL,
			InstanceID: *instance_id,
		})
	} else if *producerTopic == "import-observations-inserted" {
		schema = &events.ObservationsInsertedSchema
		insertedObservationsMsg := events.ObservationsInserted{
			InstanceID:           *instance_id,
			ObservationsInserted: int32(*insertedObservations),
		}
		log.Debug("msg", log.Data{"iom": insertedObservationsMsg})
		producerMessage, err = schema.Marshal(&insertedObservationsMsg)
	} else if *producerTopic == "data-import-complete" {
		schema = &events.DataImportCompleteSchema
		dataImportCompleteMsg := events.DataImportComplete{
			InstanceID:    *instance_id,
			DimensionName: *dimensionName,
			CodeListID:    *codeListID,
		}
		log.Debug("msg", log.Data{"dimp": dataImportCompleteMsg})
		producerMessage, err = schema.Marshal(&dataImportCompleteMsg)
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
		os.Exit(1)
	}

	producer.Output() <- producerMessage
	time.Sleep(time.Duration(1000 * time.Millisecond))
	producer.Close(context.Background())
}
