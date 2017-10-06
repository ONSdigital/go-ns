package kafka

import (
	"context"
	"errors"
	"fmt"

	"github.com/ONSdigital/go-ns/log"
	"github.com/Shopify/sarama"
)

// Producer provides a producer of Kafka messages
type Producer struct {
	producer sarama.AsyncProducer
	output   chan []byte
	errors   chan error
	closer   chan struct{}
	closed   chan struct{}
}

// Output is the channel to send outgoing messages to.
func (producer Producer) Output() chan []byte {
	return producer.output
}

// Errors provides errors returned from Kafka.
func (producer Producer) Errors() chan error {
	return producer.errors
}

// Close safely closes the consumer and releases all resources.
// pass in a context with a timeout or deadline.
// Passing a nil context will provide no timeout but is not recommended
func (producer *Producer) Close(ctx context.Context) (err error) {

	if ctx == nil {
		ctx = context.Background()
	}

	close(producer.closer)

	select {
	case <-producer.closed:
		close(producer.errors)
		close(producer.output)
		log.Info(fmt.Sprintf("Successfully closed kafka producer"), nil)
		return producer.producer.Close()

	case <-ctx.Done():
		log.Info(fmt.Sprintf("Shutdown context time exceeded, skipping graceful shutdown of consumer group"), nil)
		return errors.New("Shutdown context timed out")
	}
}

// NewProducer returns a new producer instance using the provided config. The rest of the config is set to defaults.
func NewProducer(brokers []string, topic string, envMax int) (Producer, error) {
	config := sarama.NewConfig()
	if envMax > 0 {
		config.Producer.MaxMessageBytes = envMax
	}
	producer, err := sarama.NewAsyncProducer(brokers, config)
	if err != nil {
		return Producer{}, err
	}

	outputChannel := make(chan []byte)
	errorChannel := make(chan error)
	closerChannel := make(chan struct{})
	closedChannel := make(chan struct{})

	go func() {
		defer close(closedChannel)
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

	return Producer{producer, outputChannel, errorChannel, closerChannel, closedChannel}, nil
}
