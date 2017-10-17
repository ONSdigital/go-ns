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

	// Offset returns the message offset
	Offset() int64
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

// Offset returns the message offset
func (M SaramaMessage) Offset() int64 {
	return M.message.Offset
}

// Commit the message's offset.
func (M SaramaMessage) Commit() {
	M.consumer.MarkOffset(M.message, "metadata")
}
