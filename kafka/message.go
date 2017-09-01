package kafka

import (
	"github.com/Shopify/sarama"
	"github.com/bsm/sarama-cluster"
)

type Message interface {
	GetData() []byte
	Commit()
}

type SaramaMessage struct {
	message  *sarama.ConsumerMessage
	consumer *cluster.Consumer
}

func (M SaramaMessage) GetData() []byte {
	return M.message.Value
}

func (M SaramaMessage) Commit() {
	M.consumer.MarkOffset(M.message, "metadata")
}
