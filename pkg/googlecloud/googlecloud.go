package googlecloud

import (
	"context"
	"errors"
	"time"

	monitoring "cloud.google.com/go/monitoring/apiv3"
	"github.com/golang/protobuf/ptypes/timestamp"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
)

type googleCloudClient struct {
	credentialsJSON []byte
	projectID string
}

var _ GoogleCloudClient = (*googleCloudClient)(nil)

func NewClient(credentialsJSON []byte, projectID string) (*googleCloudClient) {
	return &googleCloudClient{
		credentialsJSON: credentialsJSON,
		projectID: projectID,
	}
}

func (m *googleCloudClient) GetMetrics() (int32, error) {
	ctx := context.Background()

	client, err := monitoring.NewMetricClient(ctx, option.WithCredentialsJSON(m.credentialsJSON))

	if err != nil {
		return -1, err
	}

	const timeWindow = 2

	startTime := time.Now().UTC().Add(-timeWindow * time.Minute)
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

	it := client.ListTimeSeries(ctx, request)

	for {
		resp, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return -1, err
		}

		points := resp.Points

		if len(points) > 0 {
			return int32(points[0].GetValue().GetDoubleValue() * 100), nil
		}
		return 0, errors.New("Empty metrics points")
	}

	return -1, nil
}
