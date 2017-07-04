package kafka

import (
	"github.com/Shopify/sarama"
)

const OffsetNewest = sarama.OffsetNewest

func SetMaxMessageSize(maxSize int32) {
	sarama.MaxRequestSize = maxSize
	sarama.MaxResponseSize = maxSize
}
