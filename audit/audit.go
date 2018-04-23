package audit

import (
	"context"
	"fmt"
	"github.com/ONSdigital/go-ns/common"
	"github.com/ONSdigital/go-ns/identity"
	"github.com/ONSdigital/go-ns/log"
	"github.com/pkg/errors"
	"net/http"
	"sort"
	"time"
)

//go:generate moq -out generated_mocks.go -pkg audit . AuditorService OutboundProducer

type contextKey string

const (
	auditContextKey = contextKey("audit")
	requestIDHeader = "X-Request-Id"
)

type Error struct {
	Cause  string
	Action string
	Result string
	Params []keyValuePair
}

//Event holds data about the action being attempted
type Event struct {
	Created         string         `avro:"created" json:"created,omitempty"`
	Service         string         `avro:"service" json:"service,omitempty"`
	Namespace       string         `avro:"namespace" json:"namespace,omitempty"`
	RequestID       string         `avro:"request_id" json:"request_id,omitempty"`
	User            string         `avro:"user" json:"user,omitempty"`
	AttemptedAction string         `avro:"attempted_action" json:"attempted_action,omitempty"`
	Result          string         `avro:"result" json:"result,omitempty"`
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
	GetEvent(input context.Context) (*Event, error)
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

//Interceptor returns an http.Handler which populates a http.Request.Context with a base audit event - fields common
// to all events regardless of context before passing it to the next handler in the chain. It's recommended Interceptor
// if the first middleware in handle chain.
func (a *Auditor) Interceptor() func(handler http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			event := Event{
				RequestID: r.Header.Get(requestIDHeader),
				Namespace: a.namespace,
				Params:    make([]keyValuePair, 0),
			}

			ctx = context.WithValue(ctx, auditContextKey, event)
			handler.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

//Record captures the provided action, result and parameters and an audit event. Common fields - time, user, service
// are added automatically. An error is returned if there is a problem recording the event it is up to the caller to
// decide what do with the error in these cases.
func (a *Auditor) Record(ctx context.Context, action string, result string, params common.Params) error {
	if action == "" {
		return newAuditError("attempted action required but was empty", "", result, params)
	}
	if result == "" {
		return newAuditError("result required but was empty", action, "", params)
	}

	e, err := a.GetEvent(ctx)
	if err != nil {
		return newAuditError(err.Error(), action, result, params)
	}

	e.AttemptedAction = action
	e.Result = result
	e.Created = time.Now().String()

	if e.Service == "" {
		e.Service = identity.Caller(ctx)
	}
	if e.User == "" {
		e.User = identity.User(ctx)
	}

	if params != nil {
		for k, v := range params {
			e.Params = append(e.Params, keyValuePair{Key: k, Value: v})
		}
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

//GetEvent convenience method for getting the audit event struct from the provided context, returns an error if there is
// no event in the provided context or if the value in the context is not of the correct type.
func (a *Auditor) GetEvent(input context.Context) (*Event, error) {
	event := input.Value(auditContextKey)
	if event == nil {
		return nil, errors.New("no event found in context.Context")
	}

	auditEvent, ok := event.(Event)
	if !ok {
		return nil, errors.New("context.Context audit event was not in the expected format")
	}

	return &auditEvent, nil
}

//newAuditError creates new audit.Error with default field values where necessary and orders the params alphabetically.
func newAuditError(cause string, action string, result string, params common.Params) Error {
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
		Params: sortedParams,
	}
}

// fulfill the error interface contract
func (e Error) Error() string {
	return fmt.Sprintf("unable to audit event, action: %s, result: %s, cause: %s, params: %+v",
		e.Action, e.Result, e.Cause, e.Params)
}
