// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/0x24CaptainParrot/gophermart-service/internal/pkg/service (interfaces: Order)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	models "github.com/0x24CaptainParrot/gophermart-service/internal/models"
	service "github.com/0x24CaptainParrot/gophermart-service/internal/pkg/service"
	gomock "github.com/golang/mock/gomock"
)

// MockOrder is a mock of Order interface.
type MockOrder struct {
	ctrl     *gomock.Controller
	recorder *MockOrderMockRecorder
}

// MockOrderMockRecorder is the mock recorder for MockOrder.
type MockOrderMockRecorder struct {
	mock *MockOrder
}

// NewMockOrder creates a new mock instance.
func NewMockOrder(ctrl *gomock.Controller) *MockOrder {
	mock := &MockOrder{ctrl: ctrl}
	mock.recorder = &MockOrderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockOrder) EXPECT() *MockOrderMockRecorder {
	return m.recorder
}

// CheckOrderStatus mocks base method.
func (m *MockOrder) CheckOrderStatus(arg0 context.Context, arg1 int64, arg2 int) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CheckOrderStatus", arg0, arg1, arg2)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CheckOrderStatus indicates an expected call of CheckOrderStatus.
func (mr *MockOrderMockRecorder) CheckOrderStatus(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CheckOrderStatus", reflect.TypeOf((*MockOrder)(nil).CheckOrderStatus), arg0, arg1, arg2)
}

// CreateOrder mocks base method.
func (m *MockOrder) CreateOrder(arg0 context.Context, arg1 models.Order) (*service.ResponseInfo, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateOrder", arg0, arg1)
	ret0, _ := ret[0].(*service.ResponseInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateOrder indicates an expected call of CreateOrder.
func (mr *MockOrderMockRecorder) CreateOrder(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateOrder", reflect.TypeOf((*MockOrder)(nil).CreateOrder), arg0, arg1)
}

// ListOrders mocks base method.
func (m *MockOrder) ListOrders(arg0 context.Context, arg1 int) ([]models.Order, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListOrders", arg0, arg1)
	ret0, _ := ret[0].([]models.Order)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListOrders indicates an expected call of ListOrders.
func (mr *MockOrderMockRecorder) ListOrders(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListOrders", reflect.TypeOf((*MockOrder)(nil).ListOrders), arg0, arg1)
}
