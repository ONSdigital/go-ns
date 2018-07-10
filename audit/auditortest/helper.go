package auditortest

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/ONSdigital/go-ns/audit"
	"github.com/ONSdigital/go-ns/common"
	. "github.com/smartystreets/goconvey/convey"
)

//ErrAudit is the test error returned from a MockAuditor if the audit action & result match error trigger criteria
var ErrAudit = errors.New("auditing error")

//Expected is a struct encapsulating the method parameters to audit.Record
type Expected struct {
	Action string
	Result string
	Params common.Params
}

// actual the actual calls to auditor.Record
type actual Expected

//MockAuditor is wrapper around the generated mock implementation of audit.AuditorService which can be configured
// to return an error when specified audit action / result values are passed into the Record method, also provides
//convenience test methods for asserting calls & params made to the mock.
type MockAuditor struct {
	*audit.AuditorServiceMock
}

//NewExpectation is a constructor for creating a new Expected struct
func NewExpectation(action string, result string, params common.Params) Expected {
	return Expected{
		Action: action,
		Result: result,
		Params: params,
	}
}

//actualCalls convenience method for converting the call values to the right format.
func (m *MockAuditor) actualCalls() []actual {
	if len(m.RecordCalls()) == 0 {
		return []actual{}
	}

	actuals := make([]actual, 0)
	for _, a := range m.RecordCalls() {
		actuals = append(actuals, actual{Action: a.Action, Result: a.Result, Params: a.Params})
	}
	return actuals
}

//New creates new instance of MockAuditor that does not return any errors
func New() *MockAuditor {
	return &MockAuditor{
		&audit.AuditorServiceMock{
			RecordFunc: func(ctx context.Context, action string, result string, params common.Params) error {
				return nil
			},
		},
	}
}

//NewErroring creates new instance of MockAuditor that will return ErrAudit if the supplied audit action and result
// match the specified error trigger values.
func NewErroring(a string, r string) *MockAuditor {
	return &MockAuditor{
		&audit.AuditorServiceMock{
			RecordFunc: func(ctx context.Context, action string, result string, params common.Params) error {
				if action == a && r == result {
					audit.LogActionFailure(ctx, a, r, ErrAudit, audit.ToLogData(params))
					return ErrAudit
				}
				return nil
			},
		},
	}
}

//AssertRecordCalls is a convenience method which asserts the expected number of Record calls are made and
// the parameters of each match the expected values.
func (m *MockAuditor) AssertRecordCalls(expected ...Expected) {
	Convey("auditor.Record is called the expected number of times with the expected parameters", func() {

		// shouldAuditAsExpected is a custom implementation of a Goconvey assertion which adds additional context to
		// test failures reports to aid understanding/debugging test failures.
		shouldAuditAsExpected := func(a interface{}, e ...interface{}) string {
			if a == nil {
				return "auditor.Record could not assert audit.Record calls: actual parameter required but was empty"
			}

			if e == nil || len(e) == 0 || e[0] == nil {
				return "auditor.Record could not assert audit.Record calls: expected parameter required but was empty"
			}

			actualCalls, ok := a.([]actual)
			if !ok {
				return fmt.Sprintf("auditor.Record could not assert audit.Record calls: incorrect type for actual parameter expected: %s, actual: %s", reflect.TypeOf(actual{}), reflect.TypeOf(a))
			}

			expectedCalls, ok := e[0].([]Expected)
			if !ok {
				return fmt.Sprintf("auditor.Record could not assert audit.Record calls: incorrect type for expected parameter: expected: %s, actual: %s", reflect.TypeOf(Expected{}), reflect.TypeOf(e[0]))
			}

			if len(actualCalls) != len(expectedCalls) {
				return fmt.Sprintf("auditor.Record incorrect number of invocations, expected: %d, actual: %d", len(expectedCalls), len(actualCalls))
			}

			total := len(actualCalls)
			var invocation int
			for i, call := range actualCalls {
				invocation = i + 1

				action := expectedCalls[i].Action
				if equalErr := ShouldEqual(call.Action, action); equalErr != "" {
					return fmt.Sprintf("auditor.Record invocation %d/%d incorrect audit action - expected: %q, actual: %q", invocation, total, action, call.Action)
				}

				result := expectedCalls[i].Result
				if equalErr := ShouldEqual(call.Result, result); equalErr != "" {
					return fmt.Sprintf("auditor.Record invocation %d/%d incorrect audit result - expected: %q, actual: %q", invocation, total, result, call.Result)
				}

				params := expectedCalls[i].Params
				if equalErr := ShouldResemble(call.Params, params); equalErr != "" {
					return fmt.Sprintf("auditor.Record invocation %d/%d incorrect auditParams - expected: %+v, actual: %+v", invocation, total, params, call.Params)
				}
			}
			return ""
		}

		So(m.actualCalls(), shouldAuditAsExpected, expected)
	})
}
