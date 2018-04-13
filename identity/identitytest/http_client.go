// Code generated by moq; DO NOT EDIT
// github.com/matryer/moq

package identitytest

import (
	"context"
	"net/http"
	"sync"
)

var (
	lockHttpClientMockDo sync.RWMutex
)

// HttpClientMock is a mock implementation of HttpClient.
//
//     func TestSomethingThatUsesHttpClient(t *testing.T) {
//
//         // make and configure a mocked HttpClient
//         mockedHttpClient := &HttpClientMock{
//             DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
// 	               panic("TODO: mock out the Do method")
//             },
//         }
//
//         // TODO: use mockedHttpClient in code that requires HttpClient
//         //       and then make assertions.
//
//     }
type HttpClientMock struct {
	// DoFunc mocks the Do method.
	DoFunc func(ctx context.Context, req *http.Request) (*http.Response, error)

	// calls tracks calls to the methods.
	calls struct {
		// Do holds details about calls to the Do method.
		Do []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Req is the req argument value.
			Req *http.Request
		}
	}
}

// Do calls DoFunc.
func (mock *HttpClientMock) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	if mock.DoFunc == nil {
		panic("moq: HttpClientMock.DoFunc is nil but HttpClient.Do was just called")
	}
	callInfo := struct {
		Ctx context.Context
		Req *http.Request
	}{
		Ctx: ctx,
		Req: req,
	}
	lockHttpClientMockDo.Lock()
	mock.calls.Do = append(mock.calls.Do, callInfo)
	lockHttpClientMockDo.Unlock()
	return mock.DoFunc(ctx, req)
}

// DoCalls gets all the calls that were made to Do.
// Check the length with:
//     len(mockedHttpClient.DoCalls())
func (mock *HttpClientMock) DoCalls() []struct {
	Ctx context.Context
	Req *http.Request
} {
	var calls []struct {
		Ctx context.Context
		Req *http.Request
	}
	lockHttpClientMockDo.RLock()
	calls = mock.calls.Do
	lockHttpClientMockDo.RUnlock()
	return calls
}
