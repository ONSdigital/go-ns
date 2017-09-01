package kafka

import (
	"fmt"
	"time"

	"context"
	"github.com/ONSdigital/go-ns/log"
	"github.com/bsm/sarama-cluster"
	"github.com/johnnadratowski/golang-neo4j-bolt-driver/errors"
)

var tick = time.Millisecond * 4000

// ConsumerGroup represents a Kafka consumer group instance.
type ConsumerGroup struct {
	consumer *cluster.Consumer
	incoming chan Message
	errors   chan error
	closer   chan struct{}
	closed   chan struct{}
}

// Incoming provides a channel of incoming messages.
func (cg ConsumerGroup) Incoming() chan Message {
	return cg.incoming
}

// Errors provides a channel of incoming errors.
func (cg ConsumerGroup) Errors() chan error {
	return cg.errors
}

// Close safely closes the consumer and releases all resources.
// pass in a context with a timeout or deadline.
// Passing a nil context will provide no timeout but is not recommended
func (cg *ConsumerGroup) Close(ctx context.Context) (err error) {

	if ctx == nil {
		ctx = context.Background()
	}

	close(cg.closer)

	select {
	case <-cg.closed:
		log.Info(fmt.Sprintf("Successfully closed kafka consumer group"), nil)
		close(cg.errors)
		close(cg.incoming)
		return cg.consumer.Close()

	case <-ctx.Done():
		log.Info(fmt.Sprintf("Shutdown context time exceeded, skipping graceful shutdown of consumer group"), nil)
		return errors.New("Shutdown context timed out")
	}
}

// NewConsumerGroup returns a new consumer group using default configuration.
func NewConsumerGroup(brokers []string, topic string, group string, offset int64) (*ConsumerGroup, error) {

	config := cluster.NewConfig()
	config.Group.Return.Notifications = true
	config.Consumer.Return.Errors = true
	config.Consumer.MaxWaitTime = 50 * time.Millisecond
	config.Consumer.Offsets.Initial = offset

	consumer, err := cluster.NewConsumer(brokers, group, []string{topic}, config)
	if err != nil {
		return nil, fmt.Errorf("Bad NewConsumer of %q: %s", topic, err)
	}

	cg := ConsumerGroup{
		consumer: consumer,
		incoming: make(chan Message),
		closer:   make(chan struct{}),
		closed:   make(chan struct{}),
		errors:   make(chan error),
	}

	go func() {
		defer close(cg.closed)
		log.Info(fmt.Sprintf("Started kafka consumer of topic %q group %q", topic, group), nil)
		for {
			select {
			case err := <-consumer.Errors():
				log.Error(err, nil)
				cg.Errors() <- err
			default:
				select {
				case msg := <-consumer.Messages():
					cg.Incoming() <- SaramaMessage{msg, consumer}
				case n, more := <-consumer.Notifications():
					if more {
						log.Trace("Rebalancing group", log.Data{"topic": topic, "group": group, "partitions": n.Current[topic]})
					}
				case <-time.After(tick):
					consumer.CommitOffsets()
				case <-cg.closer:
					consumer.CommitOffsets()
					log.Info(fmt.Sprintf("Closing kafka consumer of topic %q group %q", topic, group), nil)
					return
				}
			}
		}
	}()
	return &cg, nil
}
