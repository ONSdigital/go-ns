package audit

import (
	"context"
	"fmt"
	"time"

	"github.com/ONSdigital/go-ns/common"
	"github.com/ONSdigital/go-ns/handlers/requestID"
	"github.com/ONSdigital/go-ns/identity"
	"github.com/ONSdigital/go-ns/log"
)

//go:generate moq -out generated_mocks.go -pkg audit . AuditorService OutboundProducer

type contextKey string

type Error struct {
	Cause  string
	Action string
	Result string
	Params common.Params
}

//Event holds data about the action being attempted
type Event struct {
	Created         string        `avro:"created"          json:"created,omitempty"`
	Namespace       string        `avro:"namespace"        json:"namespace,omitempty"`
	RequestID       string        `avro:"request_id"       json:"request_id,omitempty"`
	User            string        `avro:"user"             json:"user,omitempty"`
	AttemptedAction string        `avro:"attempted_action" json:"attempted_action,omitempty"`
	Result          string        `avro:"result"           json:"result,omitempty"`
	Params          common.Params `avro:"params"           json:"params,omitempty"`
}

type avroMarshaller func(s interface{}) ([]byte, error)

// OutboundProducer defines a producer for sending outbound audit events to a kafka topic
type OutboundProducer interface {
	Output() chan []byte
}

//AuditorService defines the behaviour of an auditor
type AuditorService interface {
	Record(ctx context.Context, action string, result string, params common.Params) error
}

//Auditor provides functions for interception HTTP requests and populating the context with a base audit event and
// recording audit events
type Auditor struct {
	namespace     string
	marshalToAvro avroMarshaller
	producer      OutboundProducer
}

//New creates a new Auditor with the namespace, producer and token provided.
func New(producer OutboundProducer, namespace string) *Auditor {
	return &Auditor{
		producer:      producer,
		namespace:     namespace,
		marshalToAvro: EventSchema.Marshal,
	}
}

//Record captures the provided action, result and parameters and an audit event. Common fields - time, user, service
// are added automatically. An error is returned if there is a problem recording the event it is up to the caller to
// decide what do with the error in these cases.
func (a *Auditor) Record(ctx context.Context, action string, result string, params common.Params) error {
	//NOTE: for now we are only auditing user actions - this may be subject to change
	user := identity.User(ctx)
	if user == "" {
		log.Info("not user action: skipping audit event", nil)
		return nil
	}

	if action == "" {
		return newAuditError("attempted action required but was empty", "", result, params)
	}
	if result == "" {
		return newAuditError("result required but was empty", action, "", params)
	}

	e := Event{
		Namespace:       a.namespace,
		User:            user,
		AttemptedAction: action,
		Result:          result,
		Created:         time.Now().String(),
	}

	reqID := requestID.Get(ctx)
	if reqID != "" {
		e.RequestID = reqID
	}

	if params != nil {
		e.Params = params
	}

	avroBytes, err := a.marshalToAvro(e)
	if err != nil {
		log.Error(err, nil)
		return newAuditError("error marshalling event to arvo", action, result, params)
	}

	log.Info("logging audit message", log.Data{"auditEvent": e})
	a.producer.Output() <- avroBytes
	return nil
}

//newAuditError creates new audit.Error with default field values where necessary
func newAuditError(cause string, action string, result string, params common.Params) Error {
	if params == nil {
		params = make(common.Params, 0)
	}

	if cause == "" {
		cause = "nil"
	}

	if action == "" {
		action = "nil"
	}

	if result == "" {
		result = "nil"
	}

	return Error{
		Cause:  cause,
		Action: action,
		Result: result,
		Params: params,
	}
}

// fulfill the error interface contract
func (e Error) Error() string {
	return fmt.Sprintf("unable to audit event, action: %s, result: %s, cause: %s, params: %+v",
		e.Action, e.Result, e.Cause, e.Params)
}
