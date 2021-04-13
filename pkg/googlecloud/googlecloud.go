package googlecloud

import (
	"context"
	"errors"
	"fmt"
	"time"

	"cloud.google.com/go/bigtable"
	monitoring "cloud.google.com/go/monitoring/apiv3"
	"github.com/golang/protobuf/ptypes/timestamp"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"

	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
)

type googleCloudClient struct {
	metricsClient  MetricClient
	bigtableClient BigtableClient
	projectID      string
	instanceID     string
	ctx            context.Context
}

func NewClientFromCredentials(ctx context.Context, credentialsJSON []byte, projectID, instanceID string) (GoogleCloudClient, error) {
	metricClient, err := monitoring.NewMetricClient(ctx, option.WithCredentialsJSON(credentialsJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create metrics client: %w", err)
	}
	metricClientWrapped := metricClientWrapper{
		metricsClient: metricClient,
	}

	bigtableClient, err := bigtable.NewInstanceAdminClient(ctx, projectID, option.WithCredentialsJSON(credentialsJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create bigtable client: %w", err)
	}
	bigtableClientWrapped := bigtableClientWrapper{
		bigtableClient: bigtableClient,
	}

	return NewClient(ctx, projectID, instanceID, &metricClientWrapped, &bigtableClientWrapped), nil
}

func NewClient(ctx context.Context, projectID, instanceID string, metricClientWrapped MetricClient,
	bigtableClientWrapped BigtableClient) GoogleCloudClient {
	return &googleCloudClient{
		metricsClient:  metricClientWrapped,
		bigtableClient: bigtableClientWrapped,
		projectID:      projectID,
		instanceID:     instanceID,
		ctx:            ctx,
	}
}

func (m *googleCloudClient) GetCurrentCPULoad() (int32, error) {
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
			return -1, fmt.Errorf("failed get points data from time series: %w", err)
		}
		return points[0], nil
	}
	return -1, nil
}

func (m *googleCloudClient) GetCurrentNodeCount(clusterID string) (int32, error) {
	clustersInfo, err := m.bigtableClient.Clusters(m.ctx, m.instanceID)
	if err != nil {
		return -1, fmt.Errorf("failed to get clusters info: %w", err)
	}

	for _, clusterInfo := range clustersInfo {
		if clusterInfo.Name() == clusterID {
			return clusterInfo.ServerNodes(), nil
		}
	}
	message := fmt.Sprintf("Cluster of id %s not found", clusterID)
	return -1, errors.New(message)
}
