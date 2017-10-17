package kafka

import (
	"context"
	"fmt"
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

// Release signals the successful completion of a read incoming message
func (cg ConsumerGroup) CommitAndRelease(msg Message) {
	log.Trace("pre-commit", nil)
	msg.Commit()
	log.Trace("postCommit pre-release", nil)
	cg.Release()
	log.Trace("           postRelease", nil)
}

// StopListeningToConsumer stops any more messages being consumed off kafka topic
func (cg *ConsumerGroup) StopListeningToConsumer(ctx context.Context) (err error) {

	if ctx == nil {
		ctx = context.Background()
	}

	close(cg.closer)

	select {
	case <-cg.closed:
		log.Info("StopListeningToConsumer saw closed kafka consumer", log.Data{"topic": cg.topic, "group": cg.group})
	case <-ctx.Done():
		err = fmt.Errorf("StopListeningToConsumer context timed out for group[%s] topic[%s]: %s", cg.group, cg.topic, ctx.Err())
		log.Error(err, nil)
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

	select {
	case <-cg.closed:
		close(cg.errors)
		close(cg.incoming)

		if err = cg.consumer.Close(); err != nil {
			err = fmt.Errorf("Failed to close kafka consumer group for group[%s] topic[%s]: %s", cg.group, cg.topic, err)
			log.Error(err, nil)
		} else {
			log.Info("Successfully closed kafka consumer group", log.Data{"topic": cg.topic, "group": cg.group})
		}
	case <-ctx.Done():
		err = fmt.Errorf("Shutdown context timed out for group[%s] topic[%s]: %s", cg.group, cg.topic, ctx.Err())
		log.Error(err, nil)
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

	consumer, err := cluster.NewConsumer(brokers, group, []string{topic}, config)
	if err != nil {
		return nil, fmt.Errorf("newConsumer error: group[%s] topic[%s]: %s", group, topic, err)
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
		log.Info(fmt.Sprintf("Started kafka consumer listener of topic %q group %q", cg.topic, cg.group), nil)
		defer close(cg.closed)
		for looping := true; looping; {
			log.Trace("listener.....", nil)
			select {
			case <-cg.closer:
				log.Trace("listener.....CLOSER closed - exiting listener", nil)
				looping = false
			case <-time.After(tick):
				log.Trace("listener.....tick", nil)
			case msg := <-cg.consumer.Messages():
				log.Trace("listener.....msg sending upstream, may block...", nil)
				cg.Incoming() <- SaramaMessage{msg, cg.consumer}
				if cg.sync {
					// wait for msg-processed or close-consumer triggers
					log.Trace("listener.....msg upstreamed - waiting for sync", nil)
					for loopingForSync := true; looping && loopingForSync; {
						select {
						case <-cg.upstreamDone:
							log.Trace("listener.....sync DONE!", nil)
							loopingForSync = false
						case <-cg.closer:
							// XXX if we read closer here, this means that the release/upstreamDone blocks unless it is buffered
							log.Trace("listener.....sync closer triggered", nil)
							looping = false
						case <-time.After(tick):
							log.Trace("listener.....sync tick", nil)
						}
					}
				} else {
					log.Trace("listener.....msg upstreamed", nil)
				}
			}
		}
		cg.consumer.CommitOffsets()
		log.Info(fmt.Sprintf("Closed kafka consumer listener of topic %q group %q", cg.topic, cg.group), nil)
	}()

	// control goroutine - allows us to close consumer even if blocked while upstreaming a message (above)
	go func() {
		hasBalanced := false // avoid CommitOffsets() being called before we have balanced (otherwise causes a panic)
		for looping := true; looping; {
			select {
			case <-cg.closer:
				log.Info(fmt.Sprintf("Closing kafka consumer controller of topic %q group %q", cg.topic, cg.group), nil)
				<-cg.closed
				looping = false
			case err := <-cg.consumer.Errors():
				log.Trace("controller---err msg rx", nil)
				log.Error(err, nil)
				cg.Errors() <- err
			case <-time.After(tick):
				log.Trace("controller---tick", nil)
				if hasBalanced {
					cg.consumer.CommitOffsets()
				}
			case n, more := <-cg.consumer.Notifications():
				log.Trace("controller---note", log.Data{"n": n, "more": more})
				if more {
					hasBalanced = true
					log.Trace("Rebalancing group", log.Data{"topic": cg.topic, "group": cg.group, "partitions": n.Current[cg.topic]})
				}
			}
		}
		log.Info(fmt.Sprintf("Closed kafka consumer controller of topic %q group %q", cg.topic, cg.group), nil)
	}()

	return cg, nil
}
