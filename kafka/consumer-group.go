package kafka

import (
	"context"
	"time"

	"github.com/ONSdigital/go-ns/log"
	"github.com/bsm/sarama-cluster"
)

var tick = time.Millisecond * 1500

// ConsumerGroup represents a Kafka consumer group instance.
type ConsumerGroup struct {
	consumer     *cluster.Consumer
	incoming     chan Message
	errors       chan error
	closer       chan struct{}
	closed       chan struct{}
	topic        string
	group        string
	sync         bool
	upstreamDone chan bool
}

// Incoming provides a channel of incoming messages.
func (cg ConsumerGroup) Incoming() chan Message {
	return cg.incoming
}

// Errors provides a channel of incoming errors.
func (cg ConsumerGroup) Errors() chan error {
	return cg.errors
}

// Release signals that upstream has completed an incoming message
// i.e. move on to read the next message
func (cg ConsumerGroup) Release() {
	cg.upstreamDone <- true
}

// CommitAndRelease commits the consumed message and release the consumer listener to read another message
func (cg ConsumerGroup) CommitAndRelease(msg Message) {
	msg.Commit()
	cg.Release()
}

// StopListeningToConsumer stops any more messages being consumed off kafka topic
func (cg *ConsumerGroup) StopListeningToConsumer(ctx context.Context) (err error) {

	if ctx == nil {
		ctx = context.Background()
	}

	close(cg.closer)

	logData := log.Data{"topic": cg.topic, "group": cg.group}
	select {
	case <-cg.closed:
		log.Info("StopListeningToConsumer got confirmation of closed kafka consumer listener", logData)
	case <-ctx.Done():
		err = ctx.Err()
		log.ErrorC("StopListeningToConsumer abandoned: context done", err, logData)
	}
	return
}

// Close safely closes the consumer and releases all resources.
// pass in a context with a timeout or deadline.
// Passing a nil context will provide no timeout but is not recommended
func (cg *ConsumerGroup) Close(ctx context.Context) (err error) {

	if ctx == nil {
		ctx = context.Background()
	}

	// close(closer) - the select{} avoids panic if already closed (by StopListeningToConsumer)
	select {
	case <-cg.closer:
	default:
		close(cg.closer)
	}

	logData := log.Data{"topic": cg.topic, "group": cg.group}
	select {
	case <-cg.closed:
		close(cg.errors)
		close(cg.incoming)

		if err = cg.consumer.Close(); err != nil {
			log.ErrorC("Close failed of kafka consumer group", err, logData)
		} else {
			log.Info("Successfully closed kafka consumer group", logData)
		}
	case <-ctx.Done():
		err = ctx.Err()
		log.ErrorC("Close abandoned: context done", err, logData)
	}
	return
}

// NewSyncConsumer returns a new synchronous consumer group using default configuration.
func NewSyncConsumer(brokers []string, topic string, group string, offset int64) (*ConsumerGroup, error) {
	return newConsumer(brokers, topic, group, offset, true)
}

// NewConsumerGroup returns a new asynchronous consumer group using default configuration.
func NewConsumerGroup(brokers []string, topic string, group string, offset int64) (*ConsumerGroup, error) {
	return newConsumer(brokers, topic, group, offset, false)
}

// newConsumer returns a new consumer group using default configuration.
func newConsumer(brokers []string, topic string, group string, offset int64, sync bool) (*ConsumerGroup, error) {

	config := cluster.NewConfig()
	config.Group.Return.Notifications = true
	config.Consumer.Return.Errors = true
	config.Consumer.MaxWaitTime = 50 * time.Millisecond
	config.Consumer.Offsets.Initial = offset
	config.Consumer.Offsets.Retention = 0 // indefinite retention

	logData := log.Data{"topic": topic, "group": group, "config": config}

	consumer, err := cluster.NewConsumer(brokers, group, []string{topic}, config)
	if err != nil {
		log.ErrorC("newConsumer failed", err, logData)
		return nil, err
	}

	var upstream chan Message
	if sync {
		// make the upstream channel buffered, so we can send-and-wait for upstreamDone
		upstream = make(chan Message, 1)
	} else {
		upstream = make(chan Message)
	}

	cg := &ConsumerGroup{
		consumer:     consumer,
		incoming:     upstream,
		closer:       make(chan struct{}),
		closed:       make(chan struct{}),
		errors:       make(chan error),
		topic:        topic,
		group:        group,
		sync:         sync,
		upstreamDone: make(chan bool, 1),
	}

	// listener goroutine - listen to consumer.Messages() and upstream them
	// if this blocks while upstreaming a message, we can shutdown consumer via the following goroutine
	go func() {
		logData := log.Data{"topic": topic, "group": group, "config": config}

		log.Info("Started kafka consumer listener", logData)
		defer close(cg.closed)
		for looping := true; looping; {
			select {
			case <-cg.closer:
				looping = false
			case msg := <-cg.consumer.Messages():
				cg.Incoming() <- SaramaMessage{msg, cg.consumer}
				if cg.sync {
					// wait for msg-processed or close-consumer triggers
					for loopingForSync := true; looping && loopingForSync; {
						select {
						case <-cg.upstreamDone:
							loopingForSync = false
						case <-cg.closer:
							// XXX if we read closer here, this means that the release/upstreamDone blocks unless it is buffered
							looping = false
						}
					}
				}
			}
		}
		cg.consumer.CommitOffsets()
		log.Info("Closed kafka consumer listener", logData)
	}()

	// control goroutine - allows us to close consumer even if blocked while upstreaming a message (above)
	go func() {
		logData := log.Data{"topic": topic, "group": group}

		hasBalanced := false // avoid CommitOffsets() being called before we have balanced (otherwise causes a panic)
		for looping := true; looping; {
			select {
			case <-cg.closer:
				log.Info("Closing kafka consumer controller", logData)
				<-cg.closed
				looping = false
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
					log.Trace("Rebalancing group", log.Data{"topic": cg.topic, "group": cg.group, "partitions": n.Current[cg.topic]})
				}
			}
		}
		log.Info("Closed kafka consumer controller", logData)
	}()

	return cg, nil
}
