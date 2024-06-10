// Code generated by MockGen. DO NOT EDIT.
// Source: secret_handler.go

// Package handler is a generated GoMock package.
package handler

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockSecretStorage is a mock of SecretStorage interface.
type MockSecretStorage struct {
	ctrl     *gomock.Controller
	recorder *MockSecretStorageMockRecorder
}

// MockSecretStorageMockRecorder is the mock recorder for MockSecretStorage.
type MockSecretStorageMockRecorder struct {
	mock *MockSecretStorage
}

// NewMockSecretStorage creates a new mock instance.
func NewMockSecretStorage(ctrl *gomock.Controller) *MockSecretStorage {
	mock := &MockSecretStorage{ctrl: ctrl}
	mock.recorder = &MockSecretStorageMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockSecretStorage) EXPECT() *MockSecretStorageMockRecorder {
	return m.recorder
}

// CreateSecret mocks base method.
func (m *MockSecretStorage) CreateSecret(ctx context.Context, user *User, secret *Secret) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateSecret", ctx, user, secret)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateSecret indicates an expected call of CreateSecret.
func (mr *MockSecretStorageMockRecorder) CreateSecret(ctx, user, secret interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateSecret", reflect.TypeOf((*MockSecretStorage)(nil).CreateSecret), ctx, user, secret)
}

// DeleteSecret mocks base method.
func (m *MockSecretStorage) DeleteSecret(ctx context.Context, user *User, secretKey string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteSecret", ctx, user, secretKey)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteSecret indicates an expected call of DeleteSecret.
func (mr *MockSecretStorageMockRecorder) DeleteSecret(ctx, user, secretKey interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteSecret", reflect.TypeOf((*MockSecretStorage)(nil).DeleteSecret), ctx, user, secretKey)
}

// GetSecret mocks base method.
func (m *MockSecretStorage) GetSecret(ctx context.Context, user *User, secretKey string) (*Secret, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetSecret", ctx, user, secretKey)
	ret0, _ := ret[0].(*Secret)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetSecret indicates an expected call of GetSecret.
func (mr *MockSecretStorageMockRecorder) GetSecret(ctx, user, secretKey interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSecret", reflect.TypeOf((*MockSecretStorage)(nil).GetSecret), ctx, user, secretKey)
}