package googlecloud

import (
	"context"

	monitoring "cloud.google.com/go/monitoring/apiv3"
	gax "github.com/googleapis/gax-go/v2"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
)

type GoogleCloudClient interface {
	GetMetrics() (int32, error)
}

type MonitoringMetricClient interface {
  ListTimeSeries(ctx context.Context, req *monitoringpb.ListTimeSeriesRequest, opts ...gax.CallOption) *monitoring.TimeSeriesIterator
}
