// Code generated by MockGen. DO NOT EDIT.
// Source: pkg/events.go

// Package mock is a generated GoMock package.
package mock

import (
	gomock "github.com/golang/mock/gomock"
	pkg "github.com/nuts-foundation/nuts-event-octopus/pkg"
	reflect "reflect"
)

// MockIEventPublisher is a mock of IEventPublisher interface
type MockIEventPublisher struct {
	ctrl     *gomock.Controller
	recorder *MockIEventPublisherMockRecorder
}

// MockIEventPublisherMockRecorder is the mock recorder for MockIEventPublisher
type MockIEventPublisherMockRecorder struct {
	mock *MockIEventPublisher
}

// NewMockIEventPublisher creates a new mock instance
func NewMockIEventPublisher(ctrl *gomock.Controller) *MockIEventPublisher {
	mock := &MockIEventPublisher{ctrl: ctrl}
	mock.recorder = &MockIEventPublisherMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockIEventPublisher) EXPECT() *MockIEventPublisherMockRecorder {
	return m.recorder
}

// Publish mocks base method
func (m *MockIEventPublisher) Publish(subject string, event pkg.Event) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Publish", subject, event)
	ret0, _ := ret[0].(error)
	return ret0
}

// Publish indicates an expected call of Publish
func (mr *MockIEventPublisherMockRecorder) Publish(subject, event interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Publish", reflect.TypeOf((*MockIEventPublisher)(nil).Publish), subject, event)
}

// MockEventOctopusClient is a mock of EventOctopusClient interface
type MockEventOctopusClient struct {
	ctrl     *gomock.Controller
	recorder *MockEventOctopusClientMockRecorder
}

// MockEventOctopusClientMockRecorder is the mock recorder for MockEventOctopusClient
type MockEventOctopusClientMockRecorder struct {
	mock *MockEventOctopusClient
}

// NewMockEventOctopusClient creates a new mock instance
func NewMockEventOctopusClient(ctrl *gomock.Controller) *MockEventOctopusClient {
	mock := &MockEventOctopusClient{ctrl: ctrl}
	mock.recorder = &MockEventOctopusClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockEventOctopusClient) EXPECT() *MockEventOctopusClientMockRecorder {
	return m.recorder
}

// EventPublisher mocks base method
func (m *MockEventOctopusClient) EventPublisher(clientId string) (pkg.IEventPublisher, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EventPublisher", clientId)
	ret0, _ := ret[0].(pkg.IEventPublisher)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// EventPublisher indicates an expected call of EventPublisher
func (mr *MockEventOctopusClientMockRecorder) EventPublisher(clientId interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EventPublisher", reflect.TypeOf((*MockEventOctopusClient)(nil).EventPublisher), clientId)
}

// Subscribe mocks base method
func (m *MockEventOctopusClient) Subscribe(service, subject string, callbacks map[string]pkg.EventHandlerCallback) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Subscribe", service, subject, callbacks)
	ret0, _ := ret[0].(error)
	return ret0
}

// Subscribe indicates an expected call of Subscribe
func (mr *MockEventOctopusClientMockRecorder) Subscribe(service, subject, callbacks interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Subscribe", reflect.TypeOf((*MockEventOctopusClient)(nil).Subscribe), service, subject, callbacks)
}
