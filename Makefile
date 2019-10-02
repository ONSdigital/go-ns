export KAFKA_CONSUMED_TOPIC?=input
export KAFKA_PRODUCED_TOPIC?=output
export KAFKA_CONSUME_MAX?=0
export KAFKA_PIPELINE?=0
export SNOOZE?=0
export POST_COMMIT_SNOOZE?=0s
export KAFKA_SYNC?=0
export KAFKA_ASYNC?=1
export OVERSLEEP?=0
export CHOMP_MSG?=1

export GRAPH_DRIVER_TYPE?=neptune
export GRAPH_ADDR?=ws://localhost:8182/gremlin

export HUMAN_LOG?=1

avro-test:
	go run -race cmd/avro-example/main.go --id INSTID --topic data-import-complete --dimension aggregate --codelistid f012ff16-4616-43c3-be3c-c29b27d4bb88

kafka-test:
	go run -race cmd/kafka-example/main.go

kafka-stuff:
	datetime="$(shell date '+%Y-%m-%d %H:%M:%S')";	\
	i=1; while [[ $$i -le 100 ]]; do		\
		echo $$datetime $$i;			\
		let i++;				\
	done | go run -race cmd/kafka-example/main.go

test:
	HUMAN_LOG= go test -v ./...


.PHONY: test kafka-test kafka-stuff avro-test
