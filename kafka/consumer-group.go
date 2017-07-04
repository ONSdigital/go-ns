package kafka

import (
	"fmt"
	"os"
	"time"

	"github.com/ONSdigital/go-ns/log"
	"github.com/Shopify/sarama"
	"github.com/bsm/sarama-cluster"
)

var tick = time.Millisecond * 4000

type ConsumerGroup struct {
	Consumer *cluster.Consumer
	Incoming chan Message
	Closer   chan bool
	Errors   chan error
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
	//M.consumer.CommitOffsets()
	//log.Printf("Offset : %d, Partition : %d", M.message.Offset, M.message.Partition)
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
		Consumer: consumer,
		Incoming: make(chan Message),
		Closer:   make(chan bool),
		Errors:   make(chan error),
	}
	signals := make(chan os.Signal, 1)
	//signal.Notify(signals, os.)

	go func() {
		defer cg.Consumer.Close()
		log.Info(fmt.Sprintf("Started kafka consumer of topic %q group %q", topic, group), nil)
		for {
			select {
			case err := <-cg.Consumer.Errors():
				log.Error(err, nil)
				cg.Errors <- err
			default:
				select {
				case msg := <-cg.Consumer.Messages():
					cg.Incoming <- Message{msg, cg.Consumer}
				case n, more := <-cg.Consumer.Notifications():
					if more {
						log.Trace("Rebalancing group", log.Data{"topic": topic, "group": group, "partitions": n.Current[topic]})
					}
				case <-time.After(tick):
					cg.Consumer.CommitOffsets()
				case <-signals:
					log.Info(fmt.Sprintf("Quitting kafka consumer of topic %q group %q", topic, group), nil)
					return
				case <-cg.Closer:
					log.Info(fmt.Sprintf("Closing kafka consumer of topic %q group %q", topic, group), nil)
					return
				}
			}
		}
	}()
	return &cg, nil
}
