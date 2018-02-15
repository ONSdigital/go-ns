package log

//this can't import kafka, as kafka already imports log
//this either has to go in kafka or a separate package

import (
	"net/http"
	"net/url"
	"time"
)

type AuditEvent struct {
	created         time.Time
	service         string
	namespace       string
	requestID       string
	user            string
	attemptedAction string
	responseReason  string
	request         Request
	response        Response
}

type Request struct {
	method string
	url    *url.URL
	params map[string]interface{}
}

type Response struct {
	status int
}

func Audit(r http.Request, responseStatus int, action, reason string) {
	//these things are coming from context in different ways - must be set
	//check that setting these would happen as part of the auth stuff we've designed

	namespace := r.Context().Value("namespace").(string) //this is probably just the namespace of the service? maybe this isnt context
	service := r.Context().Value("Internal-token").(string)
	user := r.Context().Value("from").(string) //from or origin headers? though the auth middleware may take that from header to context

	id := Context(&r)

	a := AuditEvent{
		created:         time.Now(),
		namespace:       namespace,
		requestID:       id,
		attemptedAction: action, //probably set per handler - might even be the permission name that allows a thing to happen
		responseReason:  reason,
	}

	var identifiable bool
	if len(service) > 0 {
		a.service = service
		identifiable = true
	}

	if len(user) > 0 {
		a.user = user
		identifiable = true
	}

	if !identifiable {
		//freak out/throw big error but still audit what we do know
		//throw an error that causes the calling function to fail - do not permit this action if we cant identify?
	}

	a.request = Request{
		method: r.Method,
		url:    r.URL,
		//params - how do we get this flexibly? read body and parse form? or just query params
	}

	//probably least needed
	a.response = Response{
		status: responseStatus,
	}

	Info("replace this with writing to kafka", Data{"event": a})
}
