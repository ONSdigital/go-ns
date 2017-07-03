package kafka

import (
	"github.com/ONSdigital/go-ns/log"
	"github.com/Shopify/sarama"
)

type Producer struct {
	producer sarama.AsyncProducer
	Output   chan []byte
	Closer   chan bool
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
	//signals := make(chan os.Signal, 1)
	//signal.Notify(signals, os.Interrupt)
	go func() {
		defer producer.Close()
		log.Info("Started kafka producer", log.Data{"topic": topic})
		for {
			select {
			case err := <-producer.Errors():
				log.ErrorC("Producer[outer]", err, log.Data{"topic": topic})
				panic(err)
			case message := <-outputChannel:

				select {
				case err := <-producer.Errors():
					log.ErrorC("Producer[inner]", err, log.Data{"topic": topic})
					panic(err)
				case producer.Input() <- &sarama.ProducerMessage{Topic: topic, Value: sarama.StringEncoder(message)}:
				}

			case <-closerChannel:
				log.Info("Closing kafka producer", log.Data{"topic": topic})
				return
				//case <-signals:
				//log.Printf("Quitting kafka producer of topic %q", topic)
				//return
			}
		}
	}()
	return Producer{producer, outputChannel, closerChannel}
}
