package audit_mock

import (
	"context"
	"errors"
	"github.com/ONSdigital/go-ns/audit"
	"github.com/ONSdigital/go-ns/common"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

//ErrAudit is the test error returned from a MockAuditor if the audit action & result match error trigger criteria
var ErrAudit = errors.New("auditing error")

//Expected is a struct encapsulating the method parameters to audit.Record
type Expected struct {
	Action string
	Result string
	Params common.Params
}

//MockAuditor is wrapper around the generated mock implementation of audit.AuditorService which can be configured
// to return an error when specified audit action / result values are passed into the Record method, also provides
//convenience test methods for asserting calls & params made to the mock.
type MockAuditor struct {
	*audit.AuditorServiceMock
	t *testing.T
}

//New creates new instance of MockAuditor that does not return any errors
func New(t *testing.T) *MockAuditor {
	return &MockAuditor{
		&audit.AuditorServiceMock{
			RecordFunc: func(ctx context.Context, action string, result string, params common.Params) error {
				return nil
			},
		},
		t,
	}
}

//NewErroring creates new instance of MockAuditor that will return ErrAudit if the supplied audit action and result
// match the specified errir trigger values.
func NewErroring(t *testing.T, a string, r string) *MockAuditor {
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
		t,
	}
}

//AssertRecordCalls is a convenience method which asserts the expected number of Record calls are made and
// the parameters of each match the expected values.
func (m *MockAuditor) AssertRecordCalls(expected ...Expected) {
	actual := m.RecordCalls()

	Convey("auditor.Record is called the expected number of times with the expected parameters", func() {
		if len(actual) != len(expected) {
			m.t.Fatalf("audit.Record incorrect number of invocations, expected: %d, actual: %d", len(expected), len(actual))
		}
		for i, call := range actual {

			if result := ShouldResemble(call.Action, expected[i].Action); result != "" {
				m.t.Fatalf("audit.Record invocation %d incorrect audit action - expected: \"%s\", actual: \"%s\"", i, expected[i].Action, call.Action)
			}

			if result := ShouldResemble(call.Result, expected[i].Result); result != "" {
				m.t.Fatalf("audit.Record invocation %d incorrect audit result - expected: \"%s\", actual: \"%s\"", i, expected[i].Result, call.Result)
			}

			if result := ShouldResemble(call.Params, expected[i].Params); result != "" {
				m.t.Fatalf("audit.Record invocation %d incorrect auditParams - expected: %+v, actual:  %+v", i, expected[i].Params, call.Params)
			}
		}
	})
}
