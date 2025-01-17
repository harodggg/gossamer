// Code generated by mockery v2.14.0. DO NOT EDIT.

package mocks

import (
	mock "github.com/stretchr/testify/mock"

	transaction "github.com/ChainSafe/gossamer/lib/transaction"

	types "github.com/ChainSafe/gossamer/dot/types"
)

// TransactionStateAPI is an autogenerated mock type for the TransactionStateAPI type
type TransactionStateAPI struct {
	mock.Mock
}

// FreeStatusNotifierChannel provides a mock function with given fields: ch
func (_m *TransactionStateAPI) FreeStatusNotifierChannel(ch chan transaction.Status) {
	_m.Called(ch)
}

// GetStatusNotifierChannel provides a mock function with given fields: ext
func (_m *TransactionStateAPI) GetStatusNotifierChannel(ext types.Extrinsic) chan transaction.Status {
	ret := _m.Called(ext)

	var r0 chan transaction.Status
	if rf, ok := ret.Get(0).(func(types.Extrinsic) chan transaction.Status); ok {
		r0 = rf(ext)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(chan transaction.Status)
		}
	}

	return r0
}

// Pending provides a mock function with given fields:
func (_m *TransactionStateAPI) Pending() []*transaction.ValidTransaction {
	ret := _m.Called()

	var r0 []*transaction.ValidTransaction
	if rf, ok := ret.Get(0).(func() []*transaction.ValidTransaction); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*transaction.ValidTransaction)
		}
	}

	return r0
}

type mockConstructorTestingTNewTransactionStateAPI interface {
	mock.TestingT
	Cleanup(func())
}

// NewTransactionStateAPI creates a new instance of TransactionStateAPI. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewTransactionStateAPI(t mockConstructorTestingTNewTransactionStateAPI) *TransactionStateAPI {
	mock := &TransactionStateAPI{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
