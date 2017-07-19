package kafka

import (
	"github.com/Shopify/sarama"
	cluster "github.com/bsm/sarama-cluster"
)

const OffsetNewest = sarama.OffsetNewest

func SetMaxMessageSize(maxSize int32) {
	sarama.MaxRequestSize = maxSize
	sarama.MaxResponseSize = maxSize
}

// MessageConsumer provides a generic interface for consuming []byte messages
type MessageConsumer interface {
	Consumer() *cluster.Consumer
	Incoming() chan Message
	Closer() chan bool
	Errors() chan error
}

// MessageProducer provides a generic interface for producing []byte messages
type MessageProducer interface {
	Producer() sarama.AsyncProducer
	Output() chan Message
	Closer() chan bool
}
