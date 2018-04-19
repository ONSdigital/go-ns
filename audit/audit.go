package audit

import (
	"context"
	"encoding/json"
	"github.com/ONSdigital/go-ns/identity"
	"github.com/ONSdigital/go-ns/log"
	"github.com/pkg/errors"
	"net/http"
	"time"
)

//go:generate moq -out ./audit_test/audit_generated_mocks.go -pkg audit_test . OutboundProducer

type contextKey string

const eventContextKey = contextKey("audit")
const fromContextKey = contextKey("from")
const requestIDHeader = "X-Request-Id"
const fromHeader = "from"

var noEventInContextError = errors.New("unable to audit no event found in context")

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

type jsonUnmarshaller func(data []byte, v interface{}) error

//Params key value pair for additional audit data not included in the event mode.
type Params map[string]string

type keyValuePair struct {
	Key   string `avro:"key" json:"key,omitempty"`
	Value string `avro:"value" json:"value,omitempty"`
}

// OutboundProducer defines a producer for sending outbound audit events to a kafka topic
type OutboundProducer interface {
	Output() chan []byte
}

//Auditor provides functions for interception HTTP requests and populating the context with a base audit event and
// recording audit events
type Auditor struct {
	namespace     string
	tokenName     string
	marshalToAvro avroMarshaller
	unmarshalJSON jsonUnmarshaller
	producer      OutboundProducer
}

//New creates a new Auditor with the namespace, producer and token provided.
func New(namespace string, producer OutboundProducer, token string) *Auditor {
	return &Auditor{
		namespace:     namespace,
		producer:      producer,
		tokenName:     token,
		marshalToAvro: EventSchema.Marshal,
		unmarshalJSON: json.Unmarshal,
	}
}

//Interceptor returns an http.Handler which populates a http.Request.Context with a base audit event - fields common
// to all events regardless of context before passing it on to the handler in the chain.
func (a *Auditor) Interceptor() func(handler http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			event := &Event{
				Params: make([]keyValuePair, 0),
			}

			if serviceToken := r.Header.Get(a.tokenName); len(serviceToken) > 0 {
				ctx = context.WithValue(ctx, a.tokenName, serviceToken)
				event.Service = serviceToken
			}

			if from := r.Header.Get(fromHeader); len(from) > 0 {
				ctx = context.WithValue(ctx, fromContextKey, from)
				event.User = from
			}

			event.RequestID = r.Header.Get(requestIDHeader)
			event.Namespace = a.namespace

			b, err := json.Marshal(event)
			if err != nil {
				log.ErrorC("cannot audit, failing request ", err, log.Data{
					"requestedURI": r.URL.RequestURI(),
					"queryString":  r.URL.RawQuery,
				})
				//we couldn't create the audit event. some input must be invalid
				//probably bomb out of the whole handler at this point.
				//Cancel the context? - dont want to allow an action we can't track
			}

			ctx = context.WithValue(ctx, eventContextKey, string(b))
			handler.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func (a *Auditor) Record(ctx context.Context, action string, result string, params Params) error {
	logData := log.Data{
		"action": action,
		"result": result,
		"params": params,
	}

	e, err := a.getEvent(ctx)
	if err != nil {
		return err
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

	eventBytes, err := a.marshalToAvro(e)
	if err != nil {
		log.ErrorC("unable to audit error marshalling event to arvo", err, logData)
		return err
	}

	log.Info("logging audit message...", nil)
	a.producer.Output() <- eventBytes
	return nil
}

func (a *Auditor) getEvent(input context.Context) (*Event, error) {
	eventBytes := input.Value(eventContextKey)
	if eventBytes == nil {
		return nil, noEventInContextError
	}

	var event Event
	if s, ok := eventBytes.(string); ok {
		b := []byte(s)
		if err := a.unmarshalJSON(b, &event); err != nil {
			return nil, err
		}
	}
	return &event, nil
}
