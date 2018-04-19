package audit

import (
	"context"
	"github.com/ONSdigital/go-ns/identity"
	"github.com/ONSdigital/go-ns/log"
	"github.com/pkg/errors"
	"net/http"
	"time"
)

//go:generate moq -out generated_mocks.go -pkg audit . AuditorService OutboundProducer

type contextKey string

const (
	eventContextKey = contextKey("audit")
	fromContextKey  = contextKey("from")
	requestIDHeader = "X-Request-Id"
	fromHeader      = "from"
)

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

//Params key value pair for additional audit data not included in the event model.
type Params map[string]string

type keyValuePair struct {
	Key   string `avro:"key" json:"key,omitempty"`
	Value string `avro:"value" json:"value,omitempty"`
}

// OutboundProducer defines a producer for sending outbound audit events to a kafka topic
type OutboundProducer interface {
	Output() chan []byte
}

//AuditorService defines the behaviour of an audior
type AuditorService interface {
	GetEvent(input context.Context) (*Event, error)
	Record(ctx context.Context, action string, result string, params Params) error
}

//Auditor provides functions for interception HTTP requests and populating the context with a base audit event and
// recording audit events
type Auditor struct {
	namespace     string
	tokenName     string
	marshalToAvro avroMarshaller
	producer      OutboundProducer
}

//New creates a new Auditor with the namespace, producer and token provided.
func New(producer OutboundProducer, namespace string, token string) *Auditor {
	return &Auditor{
		producer:      producer,
		namespace:     namespace,
		tokenName:     token,
		marshalToAvro: EventSchema.Marshal,
	}
}

//Interceptor returns an http.Handler which populates a http.Request.Context with a base audit event - fields common
// to all events regardless of context before passing it on to the handler in the chain.
func (a *Auditor) Interceptor() func(handler http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			event := Event{
				RequestID: r.Header.Get(requestIDHeader),
				Namespace: a.namespace,
				Params:    make([]keyValuePair, 0),
			}

			if serviceToken := r.Header.Get(a.tokenName); len(serviceToken) > 0 {
				ctx = context.WithValue(ctx, a.tokenName, serviceToken)
				event.Service = serviceToken
			}

			if from := r.Header.Get(fromHeader); len(from) > 0 {
				ctx = context.WithValue(ctx, fromContextKey, from)
				event.User = from
			}

			ctx = context.WithValue(ctx, eventContextKey, event)
			handler.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

//Record captures the provided action, result and parameters and an audit event. Common fields - time, user, service
// are added automatically. An error is returned if there is a problem recording the event it is up to the caller to
// decide what do with the error in these cases.
func (a *Auditor) Record(ctx context.Context, action string, result string, params Params) error {
	if action == "" {
		return auditError("attempted action is required field", "nil", params)
	}
	if result == "" {
		return auditError("result is required field", action, params)
	}

	e, err := a.GetEvent(ctx)
	if err != nil {
		return auditError(err.Error(), action, params)
	}

	e.AttemptedAction = action
	e.Result = result

	if params != nil {
		for k, v := range params {
			e.Params = append(e.Params, keyValuePair{Key: k, Value: v})
		}
	}

	e.Created = time.Now().String()
	e.Service = identity.Caller(ctx)
	e.User = identity.User(ctx)

	avroBytes, err := a.marshalToAvro(e)
	if err != nil {
		log.Error(err, nil)
		return auditError("error marshalling event to arvo", action, params)
	}

	log.Info("logging audit message", log.Data{"auditEvent": e})
	a.producer.Output() <- avroBytes
	return nil
}

//GetEvent convenience method for getting the audit event struct from the provided context, returns an error if there is
// no event in the provided context or if the value in the context is not of the correct type.
func (a *Auditor) GetEvent(input context.Context) (*Event, error) {
	event := input.Value(eventContextKey)
	if event == nil {
		return nil, errors.New("no event found in context.Context")
	}

	auditEvent, ok := event.(Event)
	if !ok {
		return nil, errors.New("context.Context audit event was not in the expected format")
	}

	return &auditEvent, nil
}

func auditError(context string, action string, params Params) error {
	if params == nil || len(params) == 0 {
		return errors.Errorf("unable to audit action: %s, %s, enforcing failure response with status code 500", action, context)
	}
	return errors.Errorf("unable to audit action: %s, params: %+v, %s, enforcing failure response with status code 500", action, params, context)
}
