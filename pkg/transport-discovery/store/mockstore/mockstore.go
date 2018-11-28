// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/watercompany/skywire-services/pkg/transport-discovery/store (interfaces: Store)

// Package mockstore is a generated GoMock package.
package mockstore

import (
	context "context"
	gomock "github.com/golang/mock/gomock"
	cipher "github.com/skycoin/skycoin/src/cipher"
	store "github.com/watercompany/skywire-services/pkg/transport-discovery/store"
	reflect "reflect"
)

// MockStore is a mock of Store interface
type MockStore struct {
	ctrl     *gomock.Controller
	recorder *MockStoreMockRecorder
}

// MockStoreMockRecorder is the mock recorder for MockStore
type MockStoreMockRecorder struct {
	mock *MockStore
}

// NewMockStore creates a new mock instance
func NewMockStore(ctrl *gomock.Controller) *MockStore {
	mock := &MockStore{ctrl: ctrl}
	mock.recorder = &MockStoreMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockStore) EXPECT() *MockStoreMockRecorder {
	return m.recorder
}

// DeregisterTransport mocks base method
func (m *MockStore) DeregisterTransport(arg0 context.Context, arg1 store.ID) error {
	ret := m.ctrl.Call(m, "DeregisterTransport", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeregisterTransport indicates an expected call of DeregisterTransport
func (mr *MockStoreMockRecorder) DeregisterTransport(arg0, arg1 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeregisterTransport", reflect.TypeOf((*MockStore)(nil).DeregisterTransport), arg0, arg1)
}

// GetNonce mocks base method
func (m *MockStore) GetNonce(arg0 context.Context, arg1 cipher.PubKey) (store.Nonce, error) {
	ret := m.ctrl.Call(m, "GetNonce", arg0, arg1)
	ret0, _ := ret[0].(store.Nonce)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetNonce indicates an expected call of GetNonce
func (mr *MockStoreMockRecorder) GetNonce(arg0, arg1 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetNonce", reflect.TypeOf((*MockStore)(nil).GetNonce), arg0, arg1)
}

// GetTransportByID mocks base method
func (m *MockStore) GetTransportByID(arg0 context.Context, arg1 store.ID) (*store.Transport, error) {
	ret := m.ctrl.Call(m, "GetTransportByID", arg0, arg1)
	ret0, _ := ret[0].(*store.Transport)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetTransportByID indicates an expected call of GetTransportByID
func (mr *MockStoreMockRecorder) GetTransportByID(arg0, arg1 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetTransportByID", reflect.TypeOf((*MockStore)(nil).GetTransportByID), arg0, arg1)
}

// GetTransportsByEdge mocks base method
func (m *MockStore) GetTransportsByEdge(arg0 context.Context, arg1 cipher.PubKey) ([]*store.Transport, error) {
	ret := m.ctrl.Call(m, "GetTransportsByEdge", arg0, arg1)
	ret0, _ := ret[0].([]*store.Transport)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetTransportsByEdge indicates an expected call of GetTransportsByEdge
func (mr *MockStoreMockRecorder) GetTransportsByEdge(arg0, arg1 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetTransportsByEdge", reflect.TypeOf((*MockStore)(nil).GetTransportsByEdge), arg0, arg1)
}

// IncrementNonce mocks base method
func (m *MockStore) IncrementNonce(arg0 context.Context, arg1 cipher.PubKey) (store.Nonce, error) {
	ret := m.ctrl.Call(m, "IncrementNonce", arg0, arg1)
	ret0, _ := ret[0].(store.Nonce)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// IncrementNonce indicates an expected call of IncrementNonce
func (mr *MockStoreMockRecorder) IncrementNonce(arg0, arg1 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IncrementNonce", reflect.TypeOf((*MockStore)(nil).IncrementNonce), arg0, arg1)
}

// RegisterTransport mocks base method
func (m *MockStore) RegisterTransport(arg0 context.Context, arg1 *store.Transport) error {
	ret := m.ctrl.Call(m, "RegisterTransport", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// RegisterTransport indicates an expected call of RegisterTransport
func (mr *MockStoreMockRecorder) RegisterTransport(arg0, arg1 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RegisterTransport", reflect.TypeOf((*MockStore)(nil).RegisterTransport), arg0, arg1)
}
