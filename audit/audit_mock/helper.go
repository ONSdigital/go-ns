package audit_mock

import (
	"context"
	"errors"
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

//MockAuditor is wrapper around the generated mock implementation of audit.AuditorService which can be configured
// to return an error when specified audit action / result values are passed into the Record method, also provides
//convenience test methods for asserting calls & params made to the mock.
type MockAuditor struct {
	*audit.AuditorServiceMock
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
// match the specified errir trigger values.
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
	actual := m.RecordCalls()

	Convey("auditor.Record is called the expected number of times", func() {
		So(len(actual), ShouldEqual, len(expected))
	})

	Convey("audit.Record is called with the expected parameters ", func() {
		for i, call := range actual {
			So(call.Action, ShouldEqual, expected[i].Action)
			So(call.Result, ShouldEqual, expected[i].Result)
			So(call.Params, ShouldResemble, expected[i].Params)
		}
	})
}
