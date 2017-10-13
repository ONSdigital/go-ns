package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/ONSdigital/go-ns/log"
	"github.com/bsm/sarama-cluster"
)

var tick = time.Millisecond * 4000

// ConsumerGroup represents a Kafka consumer group instance.
type ConsumerGroup struct {
	consumer *cluster.Consumer
	incoming chan Message
	errors   chan error
	closer   chan struct{}
	closed   chan struct{}
	topic    string
	group    string
}

// Incoming provides a channel of incoming messages.
func (cg ConsumerGroup) Incoming() chan Message {
	return cg.incoming
}

// Errors provides a channel of incoming errors.
func (cg ConsumerGroup) Errors() chan error {
	return cg.errors
}

// StopListeningToConsumer stops any more messages being consumed off kafka topic
func (cg *ConsumerGroup) StopListeningToConsumer(ctx context.Context) (err error) {

	if ctx == nil {
		ctx = context.Background()
	}

	close(cg.closer)

	select {
	case <-cg.closed:
		log.Info("Stopped listening to kafka consumer group", log.Data{"topic": cg.topic, "group": cg.group})
		return
	case <-ctx.Done():
		err = fmt.Errorf("StopListeningToConsumer context timed out for group[%s] topic[%s]: %s", cg.group, cg.topic, ctx.Err())
		log.Error(err, nil)
		return err
	}
}

// Close safely closes the consumer and releases all resources.
// pass in a context with a timeout or deadline.
// Passing a nil context will provide no timeout but is not recommended
func (cg *ConsumerGroup) Close(ctx context.Context) (err error) {

	if ctx == nil {
		ctx = context.Background()
	}

	// close(closer) - select avoids panic if already closed
	select {
	case <-cg.closer:
	default:
		close(cg.closer)
	}

	select {
	case <-cg.closed:
		close(cg.errors)
		close(cg.incoming)

		if err = cg.consumer.Close(); err != nil {
			err = fmt.Errorf("Failed to close kafka consumer group for group[%s] topic[%s]: %s", cg.group, cg.topic, err)
			log.Error(err, nil)
			return
		}

		log.Info("Successfully closed kafka consumer group", log.Data{"topic": cg.topic, "group": cg.group})
		return
	case <-ctx.Done():
		err = fmt.Errorf("Shutdown context timed out for group[%s] topic[%s]: %s", cg.group, cg.topic, ctx.Err())
		log.Error(err, nil)
		return err
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
		return nil, fmt.Errorf("Bad NewConsumer of group[%s] topic[%s]: %s", group, topic, err)
	}

	cg := ConsumerGroup{
		consumer: consumer,
		incoming: make(chan Message),
		closer:   make(chan struct{}),
		closed:   make(chan struct{}),
		errors:   make(chan error),
		topic:    topic,
		group:    group,
	}

	go func() {
		log.Info(fmt.Sprintf("Started kafka consumer of topic %q group %q", cg.topic, cg.group), nil)
		defer close(cg.closed)
		select {
		case <-cg.closer:
			cg.consumer.CommitOffsets()
			log.Info(fmt.Sprintf("Closing kafka consumer of topic %q group %q", cg.topic, cg.group), nil)
			return
		case msg := <-cg.consumer.Messages():
			cg.Incoming() <- SaramaMessage{msg, cg.consumer}
		}
	}()

	go func() {
		hasBalanced := false
		for {
			select {
			case <-cg.closer:
				return
			case err := <-cg.consumer.Errors():
				log.Error(err, nil)
				cg.Errors() <- err
			case <-time.After(tick):
				if hasBalanced {
					cg.consumer.CommitOffsets()
				}
			case n, more := <-cg.consumer.Notifications():
				if more {
					hasBalanced = true
					log.Trace("Rebalancing group", log.Data{"topic": cg.topic, "group": cg.group, "partitions": n.Current[topic]})
				}
			}
		}
	}()
	return &cg, nil
}
