package interfaces

import (
	"context"

	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
)

type GoogleCloudClient interface {
	GetMetrics() (int32, error)
}

type MetricClientWrapper interface {
	ListTimeSeries(ctx context.Context, req *monitoringpb.ListTimeSeriesRequest) TimeSeriesIteratorWrapper
}

type TimeSeriesIteratorWrapper interface {
	Points() ([]int32, error)
}
