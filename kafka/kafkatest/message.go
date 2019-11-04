package kafkatest

import (
	"sync"

	"github.com/ONSdigital/go-ns/kafka"
)

var _ kafka.Message = (*Message)(nil)

// Message allows a mock message to return the configured data, and capture whether commit has been called.
type Message struct {
	data      []byte
	committed bool
	offset    int64
	mu        sync.Mutex
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
	m.mu.Lock()
	m.committed = true
	m.mu.Unlock()
}

// Committed returns true if commit was called on this message.
func (m *Message) Committed() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.committed
}

// Offset returns the message offset
func (m *Message) Offset() int64 {
	return m.offset
}
