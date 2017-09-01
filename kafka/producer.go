package kafka

import (
	"github.com/ONSdigital/go-ns/log"
	"github.com/Shopify/sarama"
	"sync"
)

//go:generate moq -out kafkatest/sarama_async_producer.go -pkg kafkatest . SaramaAsyncProducer
type SaramaAsyncProducer sarama.AsyncProducer

type Producer struct {
	producer SaramaAsyncProducer
	output   chan []byte
	errors   chan error
	closer   chan bool
	wg       *sync.WaitGroup
}

func (producer Producer) Output() chan []byte {
	return producer.output
}

func (producer Producer) Errors() chan error {
	return producer.errors
}

// Close safely closes the consumer and releases all resources
func (producer *Producer) Close() (err error) {

	producer.closer <- true
	producer.wg.Wait()

	close(producer.errors)
	close(producer.output)

	return producer.producer.Close()
}

func NewProducer(brokers []string, topic string, envMax int) (Producer, error) {
	config := sarama.NewConfig()
	if envMax > 0 {
		config.Producer.MaxMessageBytes = envMax
	}
	producer, err := sarama.NewAsyncProducer(brokers, config)
	if err != nil {
		return Producer{}, err
	}

	var wg sync.WaitGroup
	wg.Add(1)

	outputChannel := make(chan []byte)
	closerChannel := make(chan bool)
	errorChannel := make(chan error)

	go func() {

		defer wg.Done()
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

	return Producer{producer, outputChannel, errorChannel, closerChannel, &wg}, nil
}
