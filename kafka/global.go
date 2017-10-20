package kafka

import (
	"github.com/Shopify/sarama"
)

const (
	OffsetNewest = sarama.OffsetNewest
	OffsetOldest = sarama.OffsetOldest
)

func SetMaxMessageSize(maxSize int32) {
	sarama.MaxRequestSize = maxSize
	sarama.MaxResponseSize = maxSize
}

// MessageConsumer provides a generic interface for consuming []byte messages
type MessageConsumer interface {
	Incoming() chan Message
	Closer() chan bool
	Errors() chan error
}

// MessageProducer provides a generic interface for producing []byte messages
type MessageProducer interface {
	Output() chan []byte
	Closer() chan bool
	Errors() chan error
}
