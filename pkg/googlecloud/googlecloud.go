package googlecloud

import (
	"context"
	"errors"
	"fmt"
	"time"

	"bigtable-autoscaler.com/m/v2/pkg/interfaces"
	"cloud.google.com/go/bigtable"
	monitoring "cloud.google.com/go/monitoring/apiv3"
	"github.com/golang/protobuf/ptypes/timestamp"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"

	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
)

type googleCloudClient struct {
	metricsClient  interfaces.MetricClientWrapper
	bigtableClient interfaces.BigtableClientWrapper
	projectID      string
	instanceID     string
	clusterID      string
	ctx            context.Context
}

func NewClient(ctx context.Context, credentialsJSON []byte, projectID, instanceID, clusterID string) (interfaces.GoogleCloudClient, error) {
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

	return &googleCloudClient{
		metricsClient:  &metricClientWrapped,
		bigtableClient: &bigtableClientWrapped,
		projectID:      projectID,
		instanceID:     instanceID,
		clusterID:      clusterID,
		ctx:            ctx,
	}, nil
}

func (m *googleCloudClient) GetLastCPUMeasure() (int32, error) {
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

func (m *googleCloudClient) GetCurrentNodeCount() (int32, error) {
	clustersInfo, err := m.bigtableClient.Clusters(m.ctx, m.instanceID)
	if err != nil {
		return -1, fmt.Errorf("failed to get clusters info: %w", err)
	}

	for _, clusterInfo := range clustersInfo {
		if clusterInfo.Name() == m.clusterID {
			return clusterInfo.ServerNodes(), nil
		}
	}
	message := fmt.Sprintf("Cluster of id %s not found", m.clusterID)
	return -1, errors.New(message)
}

// Make sure the real implementation complies with its interface.
var _ interfaces.GoogleCloudClient = (*googleCloudClient)(nil)
