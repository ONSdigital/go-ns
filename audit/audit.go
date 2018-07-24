package audit

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/ONSdigital/go-ns/common"
	"github.com/ONSdigital/go-ns/log"
)

// List of audit messages
const (
	Attempted    = "attempted"
	Successful   = "successful"
	Unsuccessful = "unsuccessful"

	AuditError     = "error while attempting to record audit event, failing request"
	AuditActionErr = "failed to audit action"
)

// Error represents containing details of an attempt to audit and action that failed.
type Error struct {
	Cause  string
	Action string
	Result string
	Params common.Params
}

//Event holds data about the action being attempted
type Event struct {
	Created         string        `avro:"created" json:"created,omitempty"`
	Service         string        `avro:"service" json:"service,omitempty"`
	RequestID       string        `avro:"request_id" json:"request_id,omitempty"`
	User            string        `avro:"user" json:"user,omitempty"`
	AttemptedAction string        `avro:"attempted_action" json:"attempted_action,omitempty"`
	ActionResult    string        `avro:"action_result" json:"action_result,omitempty"`
	Params          common.Params `avro:"params" json:"params,omitempty"`
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
	service       string
	marshalToAvro avroMarshaller
	producer      OutboundProducer
}

// NopAuditor is an no op implementation of the AuditorService.
type NopAuditor struct{}

//Record is a no op implementation of Auditor.Record
func (a *NopAuditor) Record(ctx context.Context, attemptedAction string, actionResult string, params common.Params) error {
	return nil
}

//New creates a new Auditor with the namespace, producer and token provided.
func New(producer OutboundProducer, namespace string) *Auditor {
	log.Debug("auditing enabled for service", log.Data{"service": namespace})
	return &Auditor{
		producer:      producer,
		service:       namespace,
		marshalToAvro: EventSchema.Marshal,
	}
}

//Record captures the provided action, result and parameters and an audit event. Common fields - time, user, service
// are added automatically. An error is returned if there is a problem recording the event it is up to the caller to
// decide what do with the error in these cases.
// NOTE: Record relies on the identity middleware having run first. If no user / service identity is available in the
// provided context an error will be returned.
func (a *Auditor) Record(ctx context.Context, attemptedAction string, actionResult string, params common.Params) (err error) {
	var e Event
	defer func() {
		if err != nil {
			LogActionFailure(ctx, attemptedAction, actionResult, err, ToLogData(params))
		} else {
			LogInfo(ctx, "captured audit event", log.Data{"auditEvent": e})
		}
	}()

	//NOTE: for now we are only auditing user actions - this may be subject to change
	user := common.User(ctx)
	service := common.Caller(ctx)

	if user == "" && service == "" {
		err = NewAuditError("expected user or caller identity but none found", attemptedAction, actionResult, params)
		return
	}

	if user == "" {
		log.DebugCtx(ctx, "not user attempted action: skipping audit event", nil)
		return
	}

	if attemptedAction == "" {
		err = NewAuditError("attemptedAction required but was empty", "", actionResult, params)
		return
	}
	if actionResult == "" {
		err = NewAuditError("actionResult required but was empty", attemptedAction, "", params)
		return
	}

	e = Event{
		Service:         a.service,
		User:            user,
		AttemptedAction: attemptedAction,
		ActionResult:    actionResult,
		Created:         time.Now().String(),
		Params:          params,
	}

	e.RequestID = common.GetRequestId(ctx)

	avroBytes, err := a.marshalToAvro(e)
	if err != nil {
		err = NewAuditError("error marshalling event to avro", attemptedAction, actionResult, params)
		return
	}

	a.producer.Output() <- avroBytes
	return
}

//NewAuditError creates new audit.Error with default field values where necessary and orders the params alphabetically.
func NewAuditError(cause string, attemptedAction string, actionResult string, params common.Params) Error {
	return Error{
		Cause:  cause,
		Action: attemptedAction,
		Result: actionResult,
		Params: params,
	}
}

// fulfill the error interface contract
func (e Error) Error() string {
	return fmt.Sprintf("unable to audit event, attempted action: %s, action result: %s, cause: %s, params: %s",
		e.Action, e.Result, e.Cause, e.formatParams())
}

//formatParams returns the params as a string - ensure the param keys are returned in a consistent order (alphabetical)
func (e Error) formatParams() string {
	if e.Params == nil || len(e.Params) == 0 {
		return "[]"
	}

	var keyValuePairs []struct {
		key   string
		value string
	}

	for k, v := range e.Params {
		keyValuePairs = append(keyValuePairs, struct {
			key   string
			value string
		}{
			key:   k,
			value: v,
		})
	}
	sort.Slice(keyValuePairs, func(i, j int) bool {
		return keyValuePairs[i].key < keyValuePairs[j].key
	})

	result := "["
	l := len(keyValuePairs)
	for i, kvp := range keyValuePairs {
		result += kvp.key + ":" + kvp.value
		if i < l-1 {
			result += ", "
		}
	}
	result += "]"
	return result
}

//ToLogData convert common.Params to log.Data
func ToLogData(p common.Params) log.Data {
	if len(p) == 0 {
		return nil
	}

	data := log.Data{}
	for k, v := range p {
		data[k] = v
	}
	return data
}
