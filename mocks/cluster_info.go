// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// ClusterInfo is an autogenerated mock type for the ClusterInfo type
type ClusterInfo struct {
	mock.Mock
}

// Name provides a mock function with given fields:
func (_m *ClusterInfo) Name() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// ServerNodes provides a mock function with given fields:
func (_m *ClusterInfo) ServerNodes() int32 {
	ret := _m.Called()

	var r0 int32
	if rf, ok := ret.Get(0).(func() int32); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int32)
	}

	return r0
}
