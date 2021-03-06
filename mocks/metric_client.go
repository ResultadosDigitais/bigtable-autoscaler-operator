// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package mocks

import (
	context "context"

	googlecloud "bigtable-autoscaler.com/m/v2/pkg/googlecloud"
	mock "github.com/stretchr/testify/mock"

	monitoring "google.golang.org/genproto/googleapis/monitoring/v3"
)

// MetricClient is an autogenerated mock type for the MetricClient type
type MetricClient struct {
	mock.Mock
}

// ListTimeSeries provides a mock function with given fields: ctx, req
func (_m *MetricClient) ListTimeSeries(ctx context.Context, req *monitoring.ListTimeSeriesRequest) googlecloud.TimeSeriesIterator {
	ret := _m.Called(ctx, req)

	var r0 googlecloud.TimeSeriesIterator
	if rf, ok := ret.Get(0).(func(context.Context, *monitoring.ListTimeSeriesRequest) googlecloud.TimeSeriesIterator); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(googlecloud.TimeSeriesIterator)
		}
	}

	return r0
}
