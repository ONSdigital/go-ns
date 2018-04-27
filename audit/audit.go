package audit

import (
	"context"
	"fmt"
	"github.com/ONSdigital/go-ns/common"
	"github.com/ONSdigital/go-ns/handlers/requestID"
	"github.com/ONSdigital/go-ns/identity"
	"github.com/ONSdigital/go-ns/log"
	"sort"
	"time"
)

//go:generate moq -out generated_mocks.go -pkg audit . AuditorService OutboundProducer

type contextKey string

type Error struct {
	Cause  string
	Action string
	Result string
	Params []keyValuePair
}

//Event holds data about the action being attempted
type Event struct {
	Created         string         `avro:"created" json:"created,omitempty"`
	Namespace       string         `avro:"namespace" json:"namespace,omitempty"`
	RequestID       string         `avro:"request_id" json:"request_id,omitempty"`
	User            string         `avro:"user" json:"user,omitempty"`
	AttemptedAction string         `avro:"attempted_action" json:"attempted_action,omitempty"`
	ActionResult    string         `avro:"action_result" json:"action_result,omitempty"`
	Params          []keyValuePair `avro:"params" json:"params,omitempty"`
}

type avroMarshaller func(s interface{}) ([]byte, error)

type keyValuePair struct {
	Key   string `avro:"key" json:"key,omitempty"`
	Value string `avro:"value" json:"value,omitempty"`
}

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
func (a *Auditor) Record(ctx context.Context, attemptedAction string, actionResult string, params common.Params) error {
	//NOTE: for now we are only auditing user actions - this may be subject to change
	user := identity.User(ctx)
	if user == "" {
		log.Debug("not user attempted action: skipping audit event", nil)
		return nil
	}

	if attemptedAction == "" {
		return NewAuditError("attemptedAction required but was empty", "", actionResult, params)
	}
	if actionResult == "" {
		return NewAuditError("actionResult required but was empty", attemptedAction, "", params)
	}

	e := Event{
		Namespace:       a.namespace,
		User:            user,
		AttemptedAction: attemptedAction,
		ActionResult:    actionResult,
		Created:         time.Now().String(),
	}

	reqID := requestID.Get(ctx)
	if reqID != "" {
		e.RequestID = reqID
	}

	if params != nil {
		for k, v := range params {
			e.Params = append(e.Params, keyValuePair{Key: k, Value: v})
		}
	}

	avroBytes, err := a.marshalToAvro(e)
	if err != nil {
		log.Error(err, nil)
		return NewAuditError("error marshalling event to arvo", attemptedAction, actionResult, params)
	}

	log.Info("logging audit message", log.Data{"auditEvent": e})
	a.producer.Output() <- avroBytes
	return nil
}

//NewAuditError creates new audit.Error with default field values where necessary and orders the params alphabetically.
func NewAuditError(cause string, attemptedAction string, actionResult string, params common.Params) Error {
	sortedParams := make([]keyValuePair, 0)

	// Params is a type alias for map and map does not guarantee the order in which the range iterates over the keyset.
	// To ensure Error() returns the same string each time it is called we convert the params to a array of
	// keyvaluepairs and sort by the key.
	if params != nil {
		for k, v := range params {
			sortedParams = append(sortedParams, keyValuePair{k, v})
		}
		sort.Slice(sortedParams, func(i, j int) bool {
			return sortedParams[i].Key < sortedParams[j].Key
		})
	}

	if cause == "" {
		cause = "nil"
	}

	if attemptedAction == "" {
		attemptedAction = "nil"
	}

	if actionResult == "" {
		actionResult = "nil"
	}

	return Error{
		Cause:  cause,
		Action: attemptedAction,
		Result: actionResult,
		Params: sortedParams,
	}
}

// fulfill the error interface contract
func (e Error) Error() string {
	return fmt.Sprintf("unable to audit event, attempted action: %s, action result: %s, cause: %s, params: %+v",
		e.Action, e.Result, e.Cause, e.Params)
}
