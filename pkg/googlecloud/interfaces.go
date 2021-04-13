package googlecloud

import (
	"context"

	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
)

type GoogleCloudClient interface {
	GetCurrentCPULoad() (int32, error)
	GetCurrentNodeCount(clusterID string) (int32, error)
}

type MetricClient interface {
	ListTimeSeries(ctx context.Context, req *monitoringpb.ListTimeSeriesRequest) TimeSeriesIterator
}

type TimeSeriesIterator interface {
	Points() ([]int32, error)
}

type BigtableClient interface {
	Clusters(ctx context.Context, instanceID string) ([]ClusterInfo, error)
}

type ClusterInfo interface {
	Name() string
	ServerNodes() int32
}
