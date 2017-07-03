package kafka

import (
	"os"

	"github.com/ONSdigital/go-ns/log"
	"github.com/Shopify/sarama"
)

type Consumer struct {
	Master   sarama.Consumer
	Consumer sarama.PartitionConsumer
	Incoming chan []byte
	Closer   chan bool
}

func NewConsumer(brokers []string, topic string, offset int64) Consumer {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	master, err := sarama.NewConsumer(brokers, config)
	if err != nil {
		panic(err)
	}

	consumer, err := master.ConsumePartition(topic, 0, offset)
	if err != nil {
		panic(err)
	}

	messageChannel := make(chan []byte)
	closerChannel := make(chan bool)
	signals := make(chan os.Signal, 1)
	//signal.Notify(signals, os.Interrupt)

	go func() {
		defer consumer.Close()
		log.Info("Started kafka consumer", log.Data{"topic": topic})
		for {
			select {
			case err := <-consumer.Errors():
				log.Error(err, log.Data{"topic": topic})
				return
			default:
				select {
				case msg := <-consumer.Messages():
					messageChannel <- msg.Value
				case <-signals:
					log.Info("Quitting kafka consumer", log.Data{"topic": topic})
					return
				case <-closerChannel:
					log.Info("Closing kafka consumer", log.Data{"topic": topic})
					return
				}
			}
		}
	}()
	return Consumer{Master: master, Consumer: consumer, Incoming: messageChannel, Closer: closerChannel}
}
