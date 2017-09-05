package kafkatest

import (
	"github.com/ONSdigital/go-ns/kafka"
)

var _ kafka.Message = (*Message)(nil)

// Message allows a mock message to return the configured data, and capture whether commit has been called.
type Message struct {
	data      []byte
	committed bool
}

// NewMessage returns a new mock message containing the given data.
func NewMessage(data []byte) *Message {
	return &Message{
		data: data,
	}
}

// GetData returns the data that was added to the struct.
func (m *Message) GetData() []byte {
	return m.data
}

// Commit captures the fact that the method was called.
func (m *Message) Commit() {
	m.committed = true
}

// Committed returns true if commit was called on this message.
func (m *Message) Committed() bool {
	return m.committed
}
