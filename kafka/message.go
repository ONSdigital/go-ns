package kafka

import "github.com/Shopify/sarama"

type Message interface {
	GetData() []byte
	Commit()
}

type SaramaMessage struct {
	message  *sarama.ConsumerMessage
	consumer SaramaClusterConsumer
}

func (M SaramaMessage) GetData() []byte {
	return M.message.Value
}

func (M SaramaMessage) Commit() {
	M.consumer.MarkOffset(M.message, "metadata")
}
