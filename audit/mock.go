package audit

import (
	"context"
	"sync"

	"github.com/ONSdigital/go-ns/common"
)

var (
	lockMockAuditorRecord sync.RWMutex
)

//AuditorServiceMock is wrapper around the generated mock implementation of audit.AuditorService which can be configured
// to return an error when specified audit action / result values are passed into the Record method, also provides
//convenience test methods for asserting calls & params made to the mock.
type AuditorServiceMock struct {
	// RecordFunc mocks the Record method.
	RecordFunc func(ctx context.Context, action string, result string, params common.Params) error

	// calls tracks calls to the methods.
	calls struct {
		// Record holds details about calls to the Record method.
		Record []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Action is the action argument value.
			Action string
			// Result is the result argument value.
			Result string
			// Params is the params argument value.
			Params common.Params
		}
	}
}

// Record calls Copy on params and RecordFunc
func (mock *AuditorServiceMock) Record(ctx context.Context, action, result string, params common.Params) error {
	newParams := params.Copy()
	if mock.RecordFunc == nil {
		panic("MockAuditor.RecordFunc is nil but MockAuditor.Record was just called")
	}

	callInfo := struct {
		Ctx    context.Context
		Action string
		Result string
		Params common.Params
	}{
		Ctx:    ctx,
		Action: action,
		Result: result,
		Params: newParams,
	}

	lockMockAuditorRecord.Lock()
	mock.calls.Record = append(mock.calls.Record, callInfo)
	lockMockAuditorRecord.Unlock()

	return mock.RecordFunc(ctx, action, result, newParams)
}

// RecordCalls retrieves all the calls that were made to Record
func (mock *AuditorServiceMock) RecordCalls() []struct {
	Ctx    context.Context
	Action string
	Result string
	Params common.Params
} {
	var calls []struct {
		Ctx    context.Context
		Action string
		Result string
		Params common.Params
	}
	lockMockAuditorRecord.RLock()
	calls = mock.calls.Record
	lockMockAuditorRecord.RUnlock()
	return calls
}

var (
	lockOutboundProducerMockOutput sync.RWMutex
)

// OutboundProducerMock is a mock implementation of OutboundProducer.
type OutboundProducerMock struct {
	// OutputFunc mocks the Output method.
	OutputFunc func() chan []byte

	// calls tracks calls to the methods.
	calls struct {
		// Output holds details about calls to the Output method.
		Output []struct {
		}
	}
}

// Output calls OutputFunc.
func (mock *OutboundProducerMock) Output() chan []byte {
	if mock.OutputFunc == nil {
		panic("moq: OutboundProducerMock.OutputFunc is nil but OutboundProducer.Output was just called")
	}
	callInfo := struct {
	}{}
	lockOutboundProducerMockOutput.Lock()
	mock.calls.Output = append(mock.calls.Output, callInfo)
	lockOutboundProducerMockOutput.Unlock()
	return mock.OutputFunc()
}

// OutputCalls gets all the calls that were made to Output.
// Check the length with:
//     len(mockedOutboundProducer.OutputCalls())
func (mock *OutboundProducerMock) OutputCalls() []struct {
} {
	var calls []struct {
	}
	lockOutboundProducerMockOutput.RLock()
	calls = mock.calls.Output
	lockOutboundProducerMockOutput.RUnlock()
	return calls
}
