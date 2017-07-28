package main

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"

	"github.com/ONSdigital/go-ns/kafka"
	"github.com/ONSdigital/go-ns/log"
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
	exitChannel := make(chan bool)
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

	go func() {
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
				producer.Closer() <- true
				consumer.Closer() <- true
				exitChannel <- true
				return
			case producerError := <-producer.Errors():
				log.Error(fmt.Errorf("Aborting"), log.Data{"messageReceived": producerError})
				producer.Closer() <- true
				consumer.Closer() <- true
				exitChannel <- true
				return
			case stdinLine := <-stdinChannel:
				producer.Output() <- []byte(stdinLine)
				log.Info("Message output", log.Data{"messageSent": stdinLine})
			case <-signals:
				log.Info("Quitting after signal", nil)
				producer.Closer() <- true
				consumer.Closer() <- true
				exitChannel <- true
				return
			}
		}
	}()
	<-exitChannel
	log.Info("Service kafka example stopped", nil)
}
