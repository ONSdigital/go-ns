package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync/atomic"

	"context"
	"time"

	"github.com/ONSdigital/go-ns/kafka"
	"github.com/ONSdigital/go-ns/log"
)

const timeOut = 5 * time.Second

func main() {
	log.Namespace = "kafka-example"

	var brokers []string
	brokers = append(brokers, "localhost:9092")
	consumedTopic := os.Getenv("KAFKA_CONSUMED_TOPIC")
	if consumedTopic == "" {
		consumedTopic = "input"
	}
	producedTopic := os.Getenv("KAFKA_PRODUCED_TOPIC")
	if producedTopic == "" {
		producedTopic = "output"
	}
	maxMessageSize := 50 * 1024 * 1024 // 50MB

	log.Info(fmt.Sprintf("[KAFKA-TEST] Starting topics: %q -> stdout, stdin -> %q", consumedTopic, producedTopic), nil)

	kafka.SetMaxMessageSize(int32(maxMessageSize))
	producer, err := kafka.NewProducer(brokers, producedTopic, maxMessageSize)
	if err != nil {
		log.ErrorC("[KAFKA-TEST] Could not create producer", err, nil)
		panic("[KAFKA-TEST] Could not create producer")
	}

	consumer, err := kafka.NewConsumerGroup(brokers, consumedTopic, log.Namespace, kafka.OffsetNewest)
	if err != nil {
		log.ErrorC("[KAFKA-TEST] Could not create consumer", err, nil)
		panic("[KAFKA-TEST] Could not create consumer")
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	stdinChannel := make(chan string)
	readyToCloseConsumer := make(chan bool)
	readyToCloseProducer := make(chan bool)

	go func(ch chan string) {
		reader := bufio.NewReader(os.Stdin)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				break
			}
			ch <- line
		}
		close(ch)
	}(stdinChannel)

	done := make(chan bool)
	go func() {
		for {
			select {
			case consumedMessage := <-consumer.Incoming():
				log.Info("[KAFKA-TEST] Received message", nil)
				consumedData := consumedMessage.GetData()
				log.Info("[KAFKA-TEST] Message consumed", log.Data{
					"messageString": string(consumedData),
					"messageRaw":    consumedData,
					"messageLen":    len(consumedData),
				})
				time.Sleep(1 * time.Second)
				producer.Output() <- consumedData
				consumedMessage.Commit()
				log.Info("[KAFKA-TEST] committed message", log.Data{"messageString": string(consumedData)})
			case consumerError := <-consumer.Errors():
				log.Error(fmt.Errorf("[KAFKA-TEST] Aborting consumer"), log.Data{"messageReceived": consumerError})
				close(done)
				return
			case producerError := <-producer.Errors():
				log.Error(fmt.Errorf("[KAFKA-TEST] Aborting producer"), log.Data{"messageReceived": producerError})
				close(done)
				return
			case stdinLine := <-stdinChannel:
				producer.Output() <- []byte(stdinLine)
				log.Info("[KAFKA-TEST] Message output", log.Data{"messageSent": stdinLine})
			}
		}
	}()

	select {
	case <-done:
		log.Info("[KAFKA-TEST] Quitting after done was closed", nil)
	case <-signals:
		log.Info("[KAFKA-TEST] Quitting after signal", nil)
	}

	// give the app 3 seconds to close gracefully before killing it.
	ctx, cancel := context.WithTimeout(context.Background(), timeOut)
	defer cancel()

	waitGroup := int32(0)

	// background a wait for the instance handler to stop
	atomic.AddInt32(&waitGroup, 1)
	go func() {
		log.Info("[KAFKA-TEST] Attempting to close kafka consumer group", nil)
		consumer.Close(ctx, readyToCloseConsumer, readyToCloseProducer)
		atomic.AddInt32(&waitGroup, -1)

		log.Info("[KAFKA-TEST] Closed kafka consumer", nil)
	}()

	// background a wait for the instance handler to stop
	atomic.AddInt32(&waitGroup, 1)
	go func() {
		log.Info("[KAFKA-TEST] Stop writing to producer", nil)
		<-readyToCloseProducer
		producer.Close(ctx)
		atomic.AddInt32(&waitGroup, -1)
		log.Info("[KAFKA-TEST] Closed kafka producer", nil)
		// Allow service to commit to consumer before closing consumerGroup
		readyToCloseConsumer <- true
	}()

	// setup a timer to zero waitGroup after timeout
	go func() {
		<-time.After(timeOut)
		log.Error(errors.New("[KAFKA-TEST] timeout while shutting down"), nil)
		atomic.AddInt32(&waitGroup, -atomic.LoadInt32(&waitGroup))
	}()

	for atomic.LoadInt32(&waitGroup) > 0 {
	}

	log.Info("[KAFKA-TEST] Service kafka example stopped", nil)
	os.Exit(1)
}
