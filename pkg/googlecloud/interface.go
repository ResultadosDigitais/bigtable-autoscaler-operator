package googlecloud

import (
	"context"

	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
)

type GoogleCloudClient interface {
	GetMetrics() (int32, error)
}

type MetricClientWrapper interface {
	NextMetric(ctx context.Context, req *monitoringpb.ListTimeSeriesRequest) (int32, error)
}
