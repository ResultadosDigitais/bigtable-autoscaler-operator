// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package mocks

import (
	context "context"

	googlecloud "bigtable-autoscaler.com/m/v2/pkg/googlecloud"
	mock "github.com/stretchr/testify/mock"
)

// BigtableClientWrapper is an autogenerated mock type for the BigtableClientWrapper type
type BigtableClientWrapper struct {
	mock.Mock
}

// Clusters provides a mock function with given fields: ctx, instanceID
func (_m *BigtableClientWrapper) Clusters(ctx context.Context, instanceID string) ([]googlecloud.ClusterInfoWrapper, error) {
	ret := _m.Called(ctx, instanceID)

	var r0 []googlecloud.ClusterInfoWrapper
	if rf, ok := ret.Get(0).(func(context.Context, string) []googlecloud.ClusterInfoWrapper); ok {
		r0 = rf(ctx, instanceID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]googlecloud.ClusterInfoWrapper)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, instanceID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}