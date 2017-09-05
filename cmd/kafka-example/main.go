package main

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"

	"context"
	"github.com/ONSdigital/go-ns/kafka"
	"github.com/ONSdigital/go-ns/log"
	"time"
)

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

	log.Info(fmt.Sprintf("Starting topics: %q -> stdout, stdin -> %q", consumedTopic, producedTopic), nil)

	kafka.SetMaxMessageSize(int32(maxMessageSize))
	producer, err := kafka.NewProducer(brokers, producedTopic, maxMessageSize)
	if err != nil {
		log.ErrorC("Could not create producer", err, nil)
		panic("Could not create producer")
	}
	consumer, err := kafka.NewConsumerGroup(brokers, consumedTopic, log.Namespace, kafka.OffsetNewest)
	if err != nil {
		log.ErrorC("Could not create consumer", err, nil)
		panic("Could not create consumer")
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	stdinChannel := make(chan string)

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

	shutdownGracefully := func() {

		// give the app 3 seconds to close gracefully before killing it.
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		producer.Close(ctx)
		consumer.Close(ctx)
		log.Info("Service kafka example stopped", nil)
		os.Exit(1)
	}

	for {
		select {
		case consumedMessage := <-consumer.Incoming():
			consumedData := consumedMessage.GetData()
			log.Info("Message consumed", log.Data{
				"messageString": string(consumedData),
				"messageRaw":    consumedData,
				"messageLen":    len(consumedData),
			})
			consumedMessage.Commit()
		case consumerError := <-consumer.Errors():
			log.Error(fmt.Errorf("Aborting"), log.Data{"messageReceived": consumerError})
			shutdownGracefully()
		case producerError := <-producer.Errors():
			log.Error(fmt.Errorf("Aborting"), log.Data{"messageReceived": producerError})
			shutdownGracefully()
		case stdinLine := <-stdinChannel:
			producer.Output() <- []byte(stdinLine)
			log.Info("Message output", log.Data{"messageSent": stdinLine})
		case <-signals:
			log.Info("Quitting after signal", nil)
			shutdownGracefully()
		}
	}
}
