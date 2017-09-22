package errorhandler

import (
	"github.com/ONSdigital/go-ns/errorhandler/models"
	"github.com/ONSdigital/go-ns/errorhandler/schema"
	"github.com/ONSdigital/go-ns/log"
)

//go:generate moq -out mock/handler.go -pkg mocks . Handler MessageProducer

var _ Handler = (*KafkaHandler)(nil)

// Handler is a generic interface for handling errors
type Handler interface {
	Handle(instanceID string, err error)
}

// KafkaHandler provides an error handler that writes to the kafka error topic
type KafkaHandler struct {
	messageProducer MessageProducer
}

//NewKafkaHandler returns a new KafkaHandler that sends error messages
func NewKafkaHandler(messageProducer MessageProducer) *KafkaHandler {
	return &KafkaHandler{
		messageProducer: messageProducer,
	}
}

//MessageProducer deoedency that writes messages to channels
type MessageProducer interface {
	Output() chan []byte
	Closer() chan bool
}

// Handle logs the error to the error handler via a kafka message
func (handler *KafkaHandler) Handle(instanceID string, err error) {

	data := log.Data{"instance_id": instanceID, "error": err.Error()}

	log.Info("Recieved error report", data)
	eventReport := errorModel.EventReport{
		InstanceID: instanceID,
		EventType:  "error",
		EventMsg:   err.Error(),
	}
	errMsg, err := errorschema.ReportedEventSchema.Marshal(eventReport)
	if err != nil {
		log.ErrorC("Failed to marshall error to event-reporter", err, data)
		return
	}
	handler.messageProducer.Output() <- errMsg
}
