package kafka

import (
	"fmt"
	"time"

	"github.com/ONSdigital/go-ns/log"
	"github.com/Shopify/sarama"
	"github.com/bsm/sarama-cluster"
)

var tick = time.Millisecond * 4000

type ConsumerGroup struct {
	consumer *cluster.Consumer
	incoming chan Message
	closer   chan bool
	errors   chan error
}

type Message struct {
	message  *sarama.ConsumerMessage
	consumer *cluster.Consumer
}

func (M Message) GetData() []byte {
	return M.message.Value
}

func (M Message) Commit() {
	M.consumer.MarkOffset(M.message, "metadata")
}

func (cg ConsumerGroup) Consumer() *cluster.Consumer {
	return cg.consumer
}

func (cg ConsumerGroup) Incoming() chan Message {
	return cg.incoming
}

func (cg ConsumerGroup) Errors() chan error {
	return cg.errors
}

func (cg ConsumerGroup) Closer() chan bool {
	return cg.closer
}

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
		closer:   make(chan bool),
		errors:   make(chan error),
	}

	go func() {
		defer cg.Consumer().Close()
		log.Info(fmt.Sprintf("Started kafka consumer of topic %q group %q", topic, group), nil)
		for {
			select {
			case err := <-cg.Consumer().Errors():
				log.Error(err, nil)
				cg.Errors() <- err
			default:
				select {
				case msg := <-cg.Consumer().Messages():
					cg.Incoming() <- Message{msg, cg.Consumer()}
				case n, more := <-cg.Consumer().Notifications():
					if more {
						log.Trace("Rebalancing group", log.Data{"topic": topic, "group": group, "partitions": n.Current[topic]})
					}
				case <-time.After(tick):
					cg.Consumer().CommitOffsets()
				case <-cg.Closer():
					log.Info(fmt.Sprintf("Closing kafka consumer of topic %q group %q", topic, group), nil)
					return
				}
			}
		}
	}()
	return &cg, nil
}
