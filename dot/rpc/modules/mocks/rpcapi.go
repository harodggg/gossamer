// Code generated by mockery v2.10.0. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// RPCAPI is an autogenerated mock type for the RPCAPI type
type RPCAPI struct {
	mock.Mock
}

// BuildMethodNames provides a mock function with given fields: rcvr, name
func (_m *RPCAPI) BuildMethodNames(rcvr interface{}, name string) {
	_m.Called(rcvr, name)
}

// Methods provides a mock function with given fields:
func (_m *RPCAPI) Methods() []string {
	ret := _m.Called()

	var r0 []string
	if rf, ok := ret.Get(0).(func() []string); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	return r0
}
