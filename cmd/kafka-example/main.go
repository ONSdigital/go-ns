package main

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"strconv"

	"context"
	"time"

	"github.com/ONSdigital/go-ns/kafka"
	"github.com/ONSdigital/go-ns/log"
)

const timeOut = 5 * time.Second

func main() {
	log.Namespace = "kafka-example"

	brokers := []string{os.Getenv("KAFKA_ADDR")}
	if brokers[0] == "" {
		brokers = []string{"localhost:9092"}
	}
	consumedTopic := os.Getenv("KAFKA_CONSUMED_TOPIC")
	if consumedTopic == "" {
		consumedTopic = "input"
	}
	consumedGroup := os.Getenv("KAFKA_CONSUMED_GROUP")
	if consumedGroup == "" {
		consumedGroup = log.Namespace
	}
	producedTopic := os.Getenv("KAFKA_PRODUCED_TOPIC")
	if producedTopic == "" {
		producedTopic = "output"
	}
	consumeCount := 0
	consumeMax := 0
	maxMessagesString := os.Getenv("KAFKA_CONSUME_MAX")
	if maxMessagesString != "" {
		var err error
		consumeMax, err = strconv.Atoi(maxMessagesString)
		if err != nil {
			panic(err)
		}
	}
	maxMessageSize := 50 * 1024 * 1024 // 50MB

	log.Info(fmt.Sprintf("[KAFKA-TEST] Starting topics: %q -> stdout, stdin -> %q", consumedTopic, producedTopic), nil)

	kafka.SetMaxMessageSize(int32(maxMessageSize))
	producer, err := kafka.NewProducer(brokers, producedTopic, maxMessageSize)
	if err != nil {
		log.ErrorC("[KAFKA-TEST] Could not create producer", err, nil)
		panic("[KAFKA-TEST] Could not create producer")
	}

	consumer, err := kafka.NewConsumerGroup(brokers, consumedTopic, consumedGroup, kafka.OffsetNewest)
	if err != nil {
		log.ErrorC("[KAFKA-TEST] Could not create consumer", err, nil)
		panic("[KAFKA-TEST] Could not create consumer")
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	stdinChannel := make(chan string)

	eventLoopContext, eventLoopCancel := context.WithCancel(context.Background())
	eventLoopDone := make(chan bool)

	go func(ch chan string) {
		reader := bufio.NewReader(os.Stdin)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				break
			}
			ch <- line
		}
		eventLoopCancel()
		<-eventLoopDone
		close(ch)
	}(stdinChannel)

	// eventLoop
	go func() {
		defer close(eventLoopDone)
		for {
			select {
			case <-eventLoopContext.Done():
				return
			case consumedMessage := <-consumer.Incoming():
				log.Info("[KAFKA-TEST] Received message", nil)
				consumedData := consumedMessage.GetData()
				log.Info("[KAFKA-TEST] Message consumed", log.Data{
					"messageString": string(consumedData),
					"messageRaw":    consumedData,
					"messageLen":    len(consumedData),
				})
				producer.Output() <- consumedData
				consumedMessage.Commit()
				log.Info("[KAFKA-TEST] committed message", log.Data{"messageString": string(consumedData)})
				consumeCount++
				if consumeCount == consumeMax {
					log.Trace("consumed max - exiting eventLoop", nil)
					return
				}
			case stdinLine := <-stdinChannel:
				producer.Output() <- []byte(stdinLine)
				log.Info("[KAFKA-TEST] Message output", log.Data{"messageSent": stdinLine, "messageChars": []byte(stdinLine)})
			}
		}
	}()

	// block until a fatal error/signal or eventLoopDone - then proceed to shutdown
	select {
	case <-eventLoopDone:
		log.Info("[KAFKA-TEST] Quitting after event loop aborted", nil)
	case <-signals:
		log.Info("[KAFKA-TEST] os signal received", nil)
	case consumerError := <-consumer.Errors():
		log.Error(fmt.Errorf("[KAFKA-TEST] Aborting consumer"), log.Data{"messageReceived": consumerError})
	case producerError := <-producer.Errors():
		log.Error(fmt.Errorf("[KAFKA-TEST] Aborting producer"), log.Data{"messageReceived": producerError})
	}

	// give the eventLoop time to close gracefully before exiting
	ctx, cancel := context.WithTimeout(context.Background(), timeOut)

	// background graceful shutdown
	go func() {
		log.Info("[KAFKA-TEST] Attempting to stop listening to kafka consumer group", nil)
		consumer.StopListeningToConsumer(ctx)
		log.Info("[KAFKA-TEST] Successfully stopped listening to kafka consumer group", nil)
		eventLoopCancel()
		<-eventLoopDone
		log.Info("[KAFKA-TEST] Attempting to close kafka producer", nil)
		producer.Close(ctx)
		log.Info("[KAFKA-TEST] Successfully closed kafka producer", nil)
		log.Info("[KAFKA-TEST] Attempting to close kafka consumer group", nil)
		consumer.Close(ctx)
		log.Info("[KAFKA-TEST] Successfully closed kafka consumer group", nil)

		log.Info("[KAFKA-TEST] Successfully shutdown", nil)
		cancel() // stop timeout
	}()

	// wait for timeout or success
	<-ctx.Done()
	log.ErrorC("[KAFKA-TEST] shutdown done", ctx.Err(), nil)
	os.Exit(1)
}
