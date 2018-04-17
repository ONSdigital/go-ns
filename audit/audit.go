package audit

//this can't import kafka, as kafka already imports log
//this either has to go in kafka or a separate package

import (
	"context"
	"encoding/json"
	"github.com/ONSdigital/go-ns/identity"
	"github.com/ONSdigital/go-ns/log"
	"net/http"
	"time"
)

type contextKey string

const eventContextKey = contextKey("audit")
const fromContextKey = contextKey("from")
const requestIDHeader = "X-Request-Id"
const fromHeader = "from"

type Event struct {
	Created         string         `avro:"created"`
	Service         string         `avro:"service"`
	Namespace       string         `avro:"name_space"`
	RequestID       string         `avro:"request_id"`
	User            string         `avro:"user"`
	AttemptedAction string         `avro:"attempted_action"`
	Result          string         `avro:"result"`
	Params          []keyValuePair `avro:"params"`
}

type keyValuePair struct {
	Key   string `avro:"key"`
	Value string `avro:"value"`
}

type OutboundProducer interface {
	Output() chan []byte
}

type Auditor struct {
	Namespace string
	TokenName string
	Producer  OutboundProducer
}

func New(namespace string, producer OutboundProducer, token string) *Auditor {
	return &Auditor{
		Namespace: namespace,
		Producer:  producer,
		TokenName: token,
	}
}

func (a *Auditor) Interceptor() func(handler http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			event := &Event{
				Params: make([]keyValuePair, 0),
			}

			if serviceToken := r.Header.Get(a.TokenName); len(serviceToken) > 0 {
				ctx = context.WithValue(ctx, a.TokenName, serviceToken)
				event.Service = serviceToken
			}

			if from := r.Header.Get(fromHeader); len(from) > 0 {
				ctx = context.WithValue(ctx, fromContextKey, from)
				event.User = from
			}

			event.RequestID = r.Header.Get(requestIDHeader)
			event.Namespace = a.Namespace

			eventString, err := json.Marshal(event)
			if err != nil {
				//we couldn't create the audit event. some input must be invalid
				//probably bomb out of the whole handler at this point.
				//Cancel the context? - dont want to allow an action we can't track
			}

			ctx = context.WithValue(ctx, eventContextKey, string(eventString))
			log.Debug("calling next handler", nil)
			handler.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func (a *Auditor) Record(ctx context.Context, action string, result string, params map[string]string) {
	e := GetEvent(ctx)

	//these things are coming from context in different ways - must be set
	//check that setting these would happen as part of the auth stuff we've designed
	e.AttemptedAction = action
	e.Result = result

	if params != nil {
		for k, v := range params {
			e.Params = append(e.Params, keyValuePair{Key: k, Value: v})
		}
	}

	//do this here - worst case we didn't unmarshal the prepared event
	//so the fields in the message are the only ones we have - best to have time to be safe
	e.Created = time.Now().String()
	e.Service = identity.Caller(ctx)
	e.User = identity.User(ctx)

	eventBytes, err := EventSchema.Marshal(e)
	if err != nil {
		log.Error(err, nil)
		return
	}

	log.Info("logging audit message...", nil)
	out := a.Producer.Output()
	out <- eventBytes
}

func GetEvent(input context.Context) *Event {
	a := &Event{}
	if s, ok := input.Value(eventContextKey).(string); ok {
		if err := json.Unmarshal([]byte(s), a); err != nil {
			log.ErrorC("error while attempting to unmarshal audit event from context", err, nil)
		}
	}
	return a
}
