package kafkatest

// MessageProducer provides a mock that allows injection of the required output channel.
type MessageProducer struct {
	outputChannel chan []byte
	closerChannel chan bool
}

//
func NewMessageProducer(outputChannel chan []byte, closerChannel chan bool) *MessageProducer {
	return &MessageProducer{
		closerChannel:closerChannel,
		outputChannel:outputChannel,
	}
}

// Output returns the injected output channel for testing.
func (messageProducer MessageProducer) Output() chan []byte {
	return messageProducer.outputChannel
}

// Closer returns the injected closer channel for testing.
func (messageProducer MessageProducer) Closer() chan bool {
	return messageProducer.closerChannel
}
