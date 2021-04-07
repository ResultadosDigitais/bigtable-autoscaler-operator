package googlecloud

import (
	"context"
	"fmt"
	"time"

	monitoring "cloud.google.com/go/monitoring/apiv3"
	"github.com/golang/protobuf/ptypes/timestamp"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"

	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
)

type metricClientWrapper struct {
	metricsClient *monitoring.MetricClient
}

func (m *metricClientWrapper) NextMetric(ctx context.Context, req *monitoringpb.ListTimeSeriesRequest) (int32, error) {
	it := m.metricsClient.ListTimeSeries(ctx, req)

	return m.nextPoint(it)
}

func (m *metricClientWrapper) nextPoint(it *monitoring.TimeSeriesIterator) (int32, error) {
	for {
		resp, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return -1, err
		}

		points := resp.Points
		return int32(points[0].GetValue().GetDoubleValue() * 100), nil
	}
	return -1, nil
}

type googleCloudClient struct {
	metricsClient MetricClientWrapper
	projectID string
	ctx context.Context
}

var _ MetricClientWrapper = (*metricClientWrapper)(nil)
var _ GoogleCloudClient = (*googleCloudClient)(nil)

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
		projectID: projectID,
		ctx: ctx,
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

	point, err := m.metricsClient.NextMetric(m.ctx, request)

	if err != nil {
		return -1, err
	}

	fmt.Println("Olá3")

	return point, nil
}
