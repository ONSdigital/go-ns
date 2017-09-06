package kafka

import (
	"github.com/Shopify/sarama"
	"github.com/bsm/sarama-cluster"
)

// Message represents a single kafka message.
type Message interface {

	// GetData returns the message contents.
	GetData() []byte

	// Commit the message's offset.
	Commit()
}

// SaramaMessage represents a Sarama specific Kafka message
type SaramaMessage struct {
	message  *sarama.ConsumerMessage
	consumer *cluster.Consumer
}

// GetData returns the message contents.
func (M SaramaMessage) GetData() []byte {
	return M.message.Value
}

// Commit the message's offset.
func (M SaramaMessage) Commit() {
	M.consumer.MarkOffset(M.message, "metadata")
}
