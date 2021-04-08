package interfaces

import (
	"context"

	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
)

type GoogleCloudClient interface {
	GetCurrentCPULoad() (int32, error)
	GetCurrentNodeCount() (int32, error)
}

type MetricClientWrapper interface {
	ListTimeSeries(ctx context.Context, req *monitoringpb.ListTimeSeriesRequest) TimeSeriesIteratorWrapper
}

type TimeSeriesIteratorWrapper interface {
	Points() ([]int32, error)
}

type BigtableClientWrapper interface {
	Clusters(ctx context.Context, instanceID string) ([]ClusterInfoWrapper, error)
}

type ClusterInfoWrapper interface {
	Name() (string)
	ServerNodes() (int32)
}
