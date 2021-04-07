package googlecloud

import (
	"context"
	"time"

	monitoring "cloud.google.com/go/monitoring/apiv3"
	"github.com/golang/protobuf/ptypes/timestamp"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"bigtable-autoscaler.com/m/v2/pkg/interfaces"

	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
)

type metricClientWrapper struct {
	metricsClient *monitoring.MetricClient
}

type timeSeriesIteratorWrapper struct {
	iterator *monitoring.TimeSeriesIterator
}

func (w *timeSeriesIteratorWrapper) Points() ([]int32, error) {
	ts, err := w.iterator.Next()

	if err != nil {
		return nil, err
	}

	normalized_points := make([]int32, 0)

	for _, point := range ts.Points {
		value := point.GetValue().GetDoubleValue() * 100
		normalized_points = append(normalized_points, int32(value))
	}

	return normalized_points, nil
}

var _ interfaces.TimeSeriesIteratorWrapper = (*timeSeriesIteratorWrapper)(nil)

func (w *metricClientWrapper) ListTimeSeries(ctx context.Context, req *monitoringpb.ListTimeSeriesRequest) interfaces.TimeSeriesIteratorWrapper {
	it := w.metricsClient.ListTimeSeries(ctx, req)

	ts := timeSeriesIteratorWrapper{iterator: it}

	return &ts
}

var _ interfaces.MetricClientWrapper = (*metricClientWrapper)(nil)

type googleCloudClient struct {
	metricsClient interfaces.MetricClientWrapper
	projectID     string
	ctx           context.Context
}

var _ interfaces.GoogleCloudClient = (*googleCloudClient)(nil)

func NewClient(ctx context.Context, credentialsJSON []byte, projectID string) (*googleCloudClient, error) {
	client, err := monitoring.NewMetricClient(ctx, option.WithCredentialsJSON(credentialsJSON))
	clientWrapped := metricClientWrapper{
		metricsClient: client,
	}

	if err != nil {
		return nil, err
	}

	return &googleCloudClient{
		metricsClient: &clientWrapped,
		projectID:     projectID,
		ctx:           ctx,
	}, nil
}

func (m *googleCloudClient) GetMetrics() (int32, error) {
	const timeWindow = 5 * time.Minute

	startTime := time.Now().UTC().Add(-timeWindow)
	endTime := time.Now().UTC()
	request := &monitoringpb.ListTimeSeriesRequest{
		Name:   "projects/" + m.projectID,
		Filter: `metric.type="bigtable.googleapis.com/cluster/cpu_load"`,
		Interval: &monitoringpb.TimeInterval{
			StartTime: &timestamp.Timestamp{
				Seconds: startTime.Unix(),
			},
			EndTime: &timestamp.Timestamp{
				Seconds: endTime.Unix(),
			},
		},
	}

	it := m.metricsClient.ListTimeSeries(m.ctx, request)

	for {
		points, err := it.Points()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return -1, err
		}
		return points[0], nil
	}
	return -1, nil
}
