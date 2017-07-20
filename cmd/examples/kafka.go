package main

import (
	"fmt"

	"github.com/ONSdigital/dp-publish-pipeline/utils"
	"github.com/ONSdigital/go-ns/kafka"
	"github.com/ONSdigital/go-ns/log"
)

func main() {
	log.Namespace = "kafka-example"

	brokers := utils.GetEnvironmentVariableAsArray("KAFKA_ADDR", "localhost:9092")
	consumedTopic := utils.GetEnvironmentVariable("CONSUMED_TOPIC", "input")
	producedTopic := utils.GetEnvironmentVariable("PRODUCED_TOPIC", "output")
	maxMessageSize, err := utils.GetEnvironmentVariableInt("KAFKA_MESSAGE_SIZE", 50*1024*1024) // default to 50MB
	if err != nil {
		log.ErrorC("Could not create consumer", err, nil)
		panic("Could not create consumer")
	}

	log.Info(fmt.Sprintf("Starting topics: %q -> stdout, stdin -> %q", consumedTopic, producedTopic), nil)

	kafka.SetMaxMessageSize(int32(maxMessageSize))
	producer := kafka.NewProducer(brokers, producedTopic, maxMessageSize)
	consumer, err := kafka.NewConsumerGroup(brokers, consumedTopic, log.Namespace, kafka.OffsetNewest)
	if err != nil {
		log.ErrorC("Could not create consumer", err, nil)
		panic("Could not create consumer")
	}

	exitChannel := make(chan bool)

	go func() {
		for {
			select {
			case consumedMessage := <-consumer.Incoming:
				log.Info(string(consumedMessage.GetData()), nil)
				producer.Output <- consumedMessage.GetData()
				consumedMessage.Commit()
			case errorMessage := <-consumer.Errors:
				log.Error(fmt.Errorf("Aborting"), log.Data{"messageReceived": errorMessage})
				producer.Closer <- true
				consumer.Closer <- true
				exitChannel <- true
				return
			}
		}
	}()
	<-exitChannel
	log.Info("Service kafka example stopped", nil)
}
