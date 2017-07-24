package kafka

import (
	"github.com/ONSdigital/go-ns/log"
	"github.com/Shopify/sarama"
)

type Producer struct {
	producer sarama.AsyncProducer
	output   chan []byte
	closer   chan bool
	errors   chan error
}

func (producer Producer) Output() chan []byte {
	return producer.output
}

func (producer Producer) Closer() chan bool {
	return producer.closer
}

func (producer Producer) Errors() chan error {
	return producer.errors
}

func NewProducer(brokers []string, topic string, envMax int) Producer {
	config := sarama.NewConfig()
	if envMax > 0 {
		config.Producer.MaxMessageBytes = envMax
	}
	producer, err := sarama.NewAsyncProducer(brokers, config)
	if err != nil {
		panic(err)
	}
	outputChannel := make(chan []byte)
	closerChannel := make(chan bool)
	errorChannel := make(chan error)
	go func() {
		defer producer.Close()
		log.Info("Started kafka producer", log.Data{"topic": topic})
		for {
			select {
			case err := <-producer.Errors():
				log.ErrorC("Producer", err, log.Data{"topic": topic})
				errorChannel <- err
			case message := <-outputChannel:
				producer.Input() <- &sarama.ProducerMessage{Topic: topic, Value: sarama.StringEncoder(message)}
			case <-closerChannel:
				log.Info("Closing kafka producer", log.Data{"topic": topic})
				return
			}
		}
	}()
	return Producer{producer, outputChannel, closerChannel, errorChannel}
}
