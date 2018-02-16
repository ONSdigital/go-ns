package log

//this can't import kafka, as kafka already imports log
//this either has to go in kafka or a separate package

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"time"
)

type AuditEvent struct {
	Created         time.Time
	Service         string
	Namespace       string
	RequestID       string
	User            string
	AttemptedAction string
	Outcome         string
	Request         Request
	ResponseStatus  int
}

type Request struct {
	Method string
	URL    *url.URL
	Params map[string]interface{}
}

//
// type Response struct {
// 	Status int
// }

type Auditor struct {
	Namespace string
	TokenName string
	//kafka vars
}

func (a *Auditor) Prepare(handle func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		event := &AuditEvent{
			Request: Request{
				Method: r.Method,
				URL:    r.URL,
				//params - how do we get this flexibly? read body and parse form? or just query params
			},
		}

		if serviceToken := r.Header.Get(a.TokenName); len(serviceToken) > 0 {
			ctx = context.WithValue(ctx, a.TokenName, serviceToken)
			event.Service = serviceToken
		}

		if from := r.Header.Get("from"); len(from) > 0 {
			ctx = context.WithValue(ctx, "from", from)
			event.User = from
		}

		event.RequestID = Context(r)
		event.Namespace = a.Namespace

		eventString, err := json.Marshal(event)
		if err != nil {
			//we couldn't create the audit event. some input must be invalid
			//probably bomb out of the whole handler at this point.
			//Cancel the context? - dont want to allow an action we can't track
		}

		ctx = context.WithValue(ctx, "audit", eventString)

		r.WithContext(ctx)
		handle(w, r)
	})
}

func (event *AuditEvent) Audit(action string, response *AuditEvent) {
	//these things are coming from context in different ways - must be set
	//check that setting these would happen as part of the auth stuff we've designed

	event.AttemptedAction = action //probably set per handler - might even be the permission name that allows a thing to happen
	event.ResponseStatus = response.ResponseStatus
	event.Outcome = response.Outcome

	//do this here - worst case we didn't unmarshal the prepared event
	//so the fields in the message are the only ones we have - best to have time to be safe
	event.Created = time.Now()

	Info("replace this with writing to kafka", Data{"event": event})
}

func UnmarshalAudit(input context.Context) *AuditEvent {
	a := &AuditEvent{}
	if s, ok := input.Value("audit").(string); ok {
		if err := json.Unmarshal([]byte(s), a); err != nil {
			//log the error, but continue, and we'll just return an empty audit event
		}
	}

	return a
}
