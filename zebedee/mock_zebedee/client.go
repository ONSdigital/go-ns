// Automatically generated by MockGen. DO NOT EDIT!
// Source: ../client.go

package mock_zebedee

import (
	"github.com/ONSdigital/go-ns/zebedee"
	gomock "github.com/golang/mock/gomock"
)

// MockClient is a mock of Client interface
type MockClient struct {
	ctrl     *gomock.Controller
	recorder *MockClientMockRecorder
}

// MockClientMockRecorder is the mock recorder for MockClient
type MockClientMockRecorder struct {
	mock *MockClient
}

// NewMockClient creates a new mock instance
func NewMockClient(ctrl *gomock.Controller) *MockClient {
	mock := &MockClient{ctrl: ctrl}
	mock.recorder = &MockClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (_m *MockClient) EXPECT() *MockClientMockRecorder {
	return _m.recorder
}

// Get mocks base method
func (_m *MockClient) Get(path string) ([]byte, error) {
	ret := _m.ctrl.Call(_m, "Get", path)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get
func (_mr *MockClientMockRecorder) Get(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Get", arg0)
}

// GetLanding mocks base method
func (_m *MockClient) GetLanding(path string) (zebedee.StaticDatasetLandingPage, error) {
	ret := _m.ctrl.Call(_m, "GetLanding", path)
	ret0, _ := ret[0].(zebedee.StaticDatasetLandingPage)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetLanding indicates an expected call of GetLanding
func (_mr *MockClientMockRecorder) GetLanding(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetLanding", arg0)
}

// SetAccessToken mocks base method
func (_m *MockClient) SetAccessToken(token string) {
	_m.ctrl.Call(_m, "SetAccessToken", token)
}

// SetAccessToken indicates an expected call of SetAccessToken
func (_mr *MockClientMockRecorder) SetAccessToken(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "SetAccessToken", arg0)
}
