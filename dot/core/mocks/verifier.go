// Code generated by mockery v2.8.0. DO NOT EDIT.

package core

import (
	types "github.com/ChainSafe/gossamer/dot/types"
	mock "github.com/stretchr/testify/mock"
)

// MockVerifier is an autogenerated mock type for the Verifier type
type MockVerifier struct {
	mock.Mock
}

// SetOnDisabled provides a mock function with given fields: authorityIndex, block
func (_m *MockVerifier) SetOnDisabled(authorityIndex uint32, block *types.Header) error {
	ret := _m.Called(authorityIndex, block)

	var r0 error
	if rf, ok := ret.Get(0).(func(uint32, *types.Header) error); ok {
		r0 = rf(authorityIndex, block)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}