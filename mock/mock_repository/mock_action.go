// Code generated by MockGen. DO NOT EDIT.
// Source: action.go

// Package mock_repository is a generated GoMock package.
package mock_repository

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	model "github.com/ww24/calendar-notifier/domain/model"
	repository "github.com/ww24/calendar-notifier/domain/repository"
)

// MockActionConfigurator is a mock of ActionConfigurator interface.
type MockActionConfigurator struct {
	ctrl     *gomock.Controller
	recorder *MockActionConfiguratorMockRecorder
}

// MockActionConfiguratorMockRecorder is the mock recorder for MockActionConfigurator.
type MockActionConfiguratorMockRecorder struct {
	mock *MockActionConfigurator
}

// NewMockActionConfigurator creates a new mock instance.
func NewMockActionConfigurator(ctrl *gomock.Controller) *MockActionConfigurator {
	mock := &MockActionConfigurator{ctrl: ctrl}
	mock.recorder = &MockActionConfiguratorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockActionConfigurator) EXPECT() *MockActionConfiguratorMockRecorder {
	return m.recorder
}

// Configure mocks base method.
func (m *MockActionConfigurator) Configure(arg0 model.ActionConfig) (repository.Action, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Configure", arg0)
	ret0, _ := ret[0].(repository.Action)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Configure indicates an expected call of Configure.
func (mr *MockActionConfiguratorMockRecorder) Configure(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Configure", reflect.TypeOf((*MockActionConfigurator)(nil).Configure), arg0)
}

// MockAction is a mock of Action interface.
type MockAction struct {
	ctrl     *gomock.Controller
	recorder *MockActionMockRecorder
}

// MockActionMockRecorder is the mock recorder for MockAction.
type MockActionMockRecorder struct {
	mock *MockAction
}

// NewMockAction creates a new mock instance.
func NewMockAction(ctrl *gomock.Controller) *MockAction {
	mock := &MockAction{ctrl: ctrl}
	mock.recorder = &MockActionMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockAction) EXPECT() *MockActionMockRecorder {
	return m.recorder
}

// List mocks base method.
func (m *MockAction) List(arg0 context.Context) (model.ScheduleEvents, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "List", arg0)
	ret0, _ := ret[0].(model.ScheduleEvents)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// List indicates an expected call of List.
func (mr *MockActionMockRecorder) List(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*MockAction)(nil).List), arg0)
}

// Register mocks base method.
func (m *MockAction) Register(arg0 context.Context, arg1 ...model.ScheduleEvent) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0}
	for _, a := range arg1 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Register", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// Register indicates an expected call of Register.
func (mr *MockActionMockRecorder) Register(arg0 interface{}, arg1 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0}, arg1...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Register", reflect.TypeOf((*MockAction)(nil).Register), varargs...)
}

// Unregister mocks base method.
func (m *MockAction) Unregister(arg0 context.Context, arg1 ...model.ScheduleEvent) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0}
	for _, a := range arg1 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Unregister", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// Unregister indicates an expected call of Unregister.
func (mr *MockActionMockRecorder) Unregister(arg0 interface{}, arg1 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0}, arg1...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Unregister", reflect.TypeOf((*MockAction)(nil).Unregister), varargs...)
}